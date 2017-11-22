package ga

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"

	"github.com/pkg/errors"

	"github.com/sunfish-shogi/sunfish4-expt/util"
)

type Config struct {
	Repository string
	Branch     string
	Directory  string
}

type CSAConfig struct {
	Host      string
	Port      int
	Pass      string
	Floodgate int
	User      string

	Depth    int
	Limit    int
	Repeat   int
	Worker   int
	Ponder   int
	UseBook  int
	HashMem  int
	MarginMs int
}

type Sunfish struct {
	Config    Config
	CSAConfig CSAConfig

	cmdCSA *exec.Cmd
}

func New() *Sunfish {
	return &Sunfish{
		Config: Config{
			Repository: "https://github.com/sunfish-shogi/sunfish4.git",
			Branch:     "master",
			Directory:  "sunfish4",
		},
		CSAConfig: CSAConfig{
			Host:      "localhost",
			Port:      4081,
			Pass:      "test-600-10,SunTest",
			Floodgate: 1,
			User:      "Sunfish4",

			Depth:    48,
			Limit:    1,
			Repeat:   1e6,
			Worker:   1,
			Ponder:   0,
			UseBook:  1,
			HashMem:  128,
			MarginMs: 500,
		},
	}
}

func (s *Sunfish) Setup() error {
	err := s.clone()
	if err != nil {
		return errors.Wrap(err, "failed to clone sunfish4")
	}

	err := s.writeCSAIni()
	if err != nil {
		return errors.Wrap(err, "failed to write csa.ini")
	}

	return nil
}

func (s *Sunfish) clone() error {
	return util.Command("git", "clone", "--depth", "1", "--branch", s.Config.Branch, s.Config.Repository, s.Config.Directory).Run()
}

func (s *Sunfish) writeCSAIni() error {
	f, err := os.OpenFile(path.Join(s.Dir(), "config/csa.ini"), os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	f.WriteString("[Server]\n")
	f.WriteString("Host      = " + s.CSAConfig.Host + "\n")
	f.WriteString("Port      = " + strconv.Itoa(s.CSAConfig.Port) + "\n")
	f.WriteString("Pass      = " + s.CSAConfig.Pass + "\n")
	f.WriteString("Floodgate = " + strconv.Itoa(s.CSAConfig.Floodgate) + "\n")
	f.WriteString("User      = " + s.CSAConfig.User + "\n")
	f.WriteString("\n")
	f.WriteString("[Search]\n")
	f.WriteString("Depth    = " + strconv.Itoa(s.CSAConfig.Depth) + "\n")
	f.WriteString("Limit    = " + strconv.Itoa(s.CSAConfig.Limit) + "\n")
	f.WriteString("Repeat   = " + strconv.Itoa(s.CSAConfig.Repeat) + "\n")
	f.WriteString("Worker   = " + strconv.Itoa(s.CSAConfig.Worker) + "\n")
	f.WriteString("Ponder   = " + strconv.Itoa(s.CSAConfig.Ponder) + "\n")
	f.WriteString("UseBook  = " + strconv.Itoa(s.CSAConfig.UseBook) + "\n")
	f.WriteString("HashMem  = " + strconv.Itoa(s.CSAConfig.HashMem) + "\n")
	f.WriteString("MarginMs = " + strconv.Itoa(s.CSAConfig.MarginMs) + "\n")
	f.WriteString("\n")
	f.WriteString("[KeepAlive]\n")
	f.WriteString("KeepAlive = 1\n")
	f.WriteString("KeepIdle  = 10\n")
	f.WriteString("KeepIntvl = 5\n")
	f.WriteString("KeepCnt   = 10\n")
	f.WriteString("\n")
	f.WriteString("[File]\n")
	f.WriteString("KifuDir   = out/csa_kifu\n")

	return f.Close()
}

func (s *Sunfish) SymlinkEvalBin(orgPath string) error {
	return util.Symlink(orgPath, path.Join(s.Dir(), "eval.bin"))
}

func (s *Sunfish) SymlinkBookBin(orgPath string) error {
	return util.Symlink(orgPath, path.Join(s.Dir(), "book.bin"))
}

func (s *Sunfish) StartCSA() error {
	if s.cmdCSA != nil {
		return fmt.Errorf("sunfish_csa already started")
	}

	cmd := util.Command("make", "csa")
	cmd.Dir = s.Dir()
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to build sunfish_csa")
	}

	s.cmdCSA = util.Command(path.Join(s.Dir(), "sunfish_csa"), "-s")
	s.cmdCSA.Dir = path.Join(s.Dir())
	return s.cmdCSA.Start()
}

func (s *Sunfish) StopCSA() error {
	if s.cmdCSA != nil {
		if s.cmdCSA.Process != nil {
			if err := s.cmdCSA.Process.Kill(); err != nil {
				return err
			} else if _, err := s.cmdCSA.Process.Wait(); err != nil {
				return err
			}
		}
		s.cmdCSA = nil
	}
	return nil
}

func (s *Sunfish) WriteParamHpp(params map[string]int) error {
	f, err := os.OpenFile(path.Join(s.Dir(), "src/search/Param.hpp"), os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	for key, value := range params {
		fmt.Fprintf(f, "#define %s %d\n", key, value)
	}

	return f.Close()
}

var regexpLogRecvWin = regexp.MustCompile(`^[^ ]* \[RECV\] +#WIN$`)
var regexpLogRecvLose = regexp.MustCompile(`^[^ ]* \[RECV\] +#LOSE$`)

type Score struct {
	Win  int
	Lose int
}

func (s *Sunfish) GetScore() (Score, error) {
	f, err := os.Open(path.Join(s.Dir(), "out/csa.log"))
	if err != nil {
		return Score{}, err
	}
	defer f.Close()

	var score Score

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if regexpLogRecvWin.MatchString(line) {
			score.Win++
		} else if regexpLogRecvLose.MatchString(line) {
			score.Lose++
		}
	}
	if err := scanner.Err(); err != nil {
		return Score{}, err
	}

	return score, nil
}

func (s *Sunfish) Clean() {
	os.RemoveAll(s.Dir())
}

func (s *Sunfish) Dir() string {
	return path.Join(util.WorkDir(), s.Config.Directory)
}

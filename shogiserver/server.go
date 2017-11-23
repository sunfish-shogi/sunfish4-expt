package shogiserver

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/sunfish-shogi/sunfish4-expt/util"
)

var recordDir = regexp.MustCompile(`^[0-9]+$`)

type Config struct {
	Repository string
	Branch     string
	Directory  string
}

type ShogiServer struct {
	Config Config

	cmd *exec.Cmd
}

func New() *ShogiServer {
	return &ShogiServer{
		Config: Config{
			Repository: "git://git.pf.osdn.jp/gitroot/s/su/sunfish-shogi/shogi-server.git",
			Branch:     "master",
			Directory:  "shogi-server",
		},
	}
}

func (s *ShogiServer) Dir() string {
	return path.Join(util.WorkDir(), s.Config.Directory)
}

func (s *ShogiServer) Setup() error {
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", s.Config.Branch, s.Config.Repository, s.Config.Directory)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to clone shogi-server")
	}

	s.cmd = exec.Command("ruby", "shogi-server", "test", "4081")
	s.cmd.Dir = s.Dir()
	if err := s.cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start shogi-server")
	}
	return nil
}

func (s *ShogiServer) Stop() {
	if s.cmd != nil {
		if s.cmd.Process != nil {
			if err := s.cmd.Process.Kill(); err != nil {
				log.Println(err)
			} else if _, err := s.cmd.Process.Wait(); err != nil {
				log.Println(err)
			}
		}
		s.cmd = nil
	}
}

func (s *ShogiServer) MakeRate() (Rate, error) {
	files, err := ioutil.ReadDir(s.Dir())
	if err != nil {
		return Rate{}, err
	}

	cmdParams := make([]string, 0, 8)
	cmdParams = append(cmdParams, path.Join(s.Dir(), "mk_game_results"))
	for i := range files {
		if files[i].IsDir() && recordDir.MatchString(files[i].Name()) {
			cmdParams = append(cmdParams, files[i].Name())
		}
	}
	cmdParams = append(cmdParams, "|")
	cmdParams = append(cmdParams, "grep", "-v", "abnormal")
	cmdParams = append(cmdParams, "|")
	cmdParams = append(cmdParams, path.Join(s.Dir(), "mk_rate"))

	cmd := exec.Command("sh", "-c", strings.Join(cmdParams, " "))
	cmd.Dir = s.Dir()
	buf, err := cmd.Output()
	if err != nil {
		return Rate{}, err
	}

	return UnmarshalRate(buf)
}

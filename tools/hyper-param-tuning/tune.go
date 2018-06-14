package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	server "github.com/sunfish-shogi/sunfish4-expt/shogiserver"
	"github.com/sunfish-shogi/sunfish4-expt/sunfish"
	"github.com/sunfish-shogi/sunfish4-expt/util"
)

type player struct {
	name    string
	sunfish *sunfish.Sunfish
	values  []int32
}

type Manager struct {
	Config Config

	server      *server.ShogiServer
	newPlayers  []*player
	currPlayers []*player
	values      []int32
	procMutex   sync.Mutex
}

func NewManager(config Config) *Manager {
	return &Manager{
		Config: config,
		values: generateNormalValues(config),
	}
}

func (m *Manager) Run() error {
	checkFiles()

	if err := m.start(); err != nil {
		log.Fatal(err)
	}

	defer m.Destroy()

	for {
		for pi := range m.Config.Params {
			m.next(pi)
		}
	}
}

func (m *Manager) start() error {
	m.procMutex.Lock()
	defer m.procMutex.Unlock()

	m.server = server.New()
	return m.server.Setup()
}

func (m *Manager) next(pi int) {
	log.Printf("%d: %s curr=%d\n", pi, m.Config.Params[pi].Name, m.values[pi])
	log.Println()

	m.currPlayers = make([]*player, m.Config.Concurrency)
	for ci := 0; ci < m.Config.Concurrency; ci++ {
		name := fmt.Sprintf("c-%d", ci)
		m.currPlayers[ci] = m.generatePlayer(name, ci, m.values)
	}

	vl := make([]int32, len(m.Config.Params))
	vh := make([]int32, len(m.Config.Params))

	copy(vl, m.values)
	copy(vh, m.values)

	if m.values[pi] < m.Config.Params[pi].Minimum+m.Config.Params[pi].Step {
		vl[pi] += m.Config.Params[pi].Step
		vh[pi] += m.Config.Params[pi].Step * 2
	} else if m.values[pi] > m.Config.Params[pi].Maximum-m.Config.Params[pi].Step {
		vl[pi] -= m.Config.Params[pi].Step * 2
		vh[pi] -= m.Config.Params[pi].Step
	} else {
		vl[pi] -= m.Config.Params[pi].Step
		vh[pi] += m.Config.Params[pi].Step
	}
	log.Printf("low  values %s\n", stringifyValues(vl))
	log.Printf("high values %s\n", stringifyValues(vh))
	log.Println()

	m.newPlayers = make([]*player, m.Config.Concurrency)
	for ci := 0; ci < m.Config.Concurrency; ci++ {
		if ci < m.Config.Concurrency/2 {
			name := fmt.Sprintf("l-%d", ci)
			m.newPlayers[ci] = m.generatePlayer(name, ci, vl)
		} else {
			name := fmt.Sprintf("h-%d", ci)
			m.newPlayers[ci] = m.generatePlayer(name, ci, vh)
		}
	}

	if err := m.setupPlayers(m.Config); err != nil {
		log.Println(err)
	}

	if err := m.startPlayers(); err != nil {
		log.Println(err)
	}

	var lscore sunfish.Score
	var hscore sunfish.Score
	for {
		time.Sleep(time.Minute * 10)

		lscore = sunfish.Score{}
		hscore = sunfish.Score{}
		for ci := 0; ci < m.Config.Concurrency; ci++ {
			p := m.newPlayers[ci]
			if score, err := p.sunfish.GetScore(); err != nil {
				log.Println(err)
				continue
			} else {
				if ci < m.Config.Concurrency/2 {
					lscore.Win += score.Win
					lscore.Lose += score.Lose
				} else {
					hscore.Win += score.Win
					hscore.Lose += score.Lose
				}
			}
		}
		if lscore.Win+lscore.Lose >= m.Config.NumberOfGames &&
			hscore.Win+hscore.Lose >= m.Config.NumberOfGames {
			break
		}
	}
	lrate := float64(lscore.Win) / float64(lscore.Win+lscore.Lose)
	hrate := float64(hscore.Win) / float64(hscore.Win+hscore.Lose)
	log.Printf("low  total %d - %d (%f)\n", lscore.Win, lscore.Lose, lrate)
	log.Printf("high total %d - %d (%f)\n", hscore.Win, hscore.Lose, hrate)
	log.Println()

	if lrate <= 0.5 && hrate <= 0.5 {
		log.Println("do not update")
	} else {
		if lrate > hrate {
			m.values[pi] = vl[pi]
		} else {
			m.values[pi] = vh[pi]
		}
		log.Printf("update %s %d\n", m.Config.Params[pi].Name, m.values[pi])
		log.Printf("values %s\n", stringifyValues(m.values))
	}
	log.Println()

	// Stop Players
	stopPlayers(append(m.newPlayers, m.currPlayers...))
}

func (m *Manager) generatePlayer(name string, gameNumber int, values []int32) *player {
	s := sunfish.New()
	s.Config.Directory = name
	s.Config.Branch = m.Config.Branch
	s.CSAConfig.Pass = "test" + strconv.Itoa(gameNumber) + "-600-10,SunTest"
	s.CSAConfig.User = name
	s.CSAConfig.Limit = m.Config.MoveLimit
	return &player{
		name:    name,
		sunfish: s,
		values:  values,
	}
}

func (m *Manager) Destroy() {
	m.procMutex.Lock()
	defer m.procMutex.Unlock()

	stopPlayers(append(m.newPlayers, m.currPlayers...))
	m.server.Stop()
}

var evalBinPath = path.Join(util.WorkDir(), "eval.bin")
var bookBinPath = path.Join(util.WorkDir(), "book.bin")

func checkFiles() {
	if _, err := os.Stat(evalBinPath); err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(bookBinPath); err != nil {
		log.Fatal(err)
	}
}

func (m *Manager) setupPlayers(config Config) error {
	m.procMutex.Lock()
	defer m.procMutex.Unlock()

	var eg errgroup.Group
	for _, _p := range append(m.newPlayers, m.currPlayers...) {
		p := _p
		eg.Go(func() error {
			if err := p.sunfish.Setup(); err != nil {
				err = errors.Wrap(err, fmt.Sprintf("failed to setup sunfish %s", p.name))
				log.Println(err)
				return err
			}

			if err := p.sunfish.SymlinkEvalBin(evalBinPath); err != nil {
				return err
			}

			if err := p.sunfish.SymlinkBookBin(bookBinPath); err != nil {
				return err
			}

			if err := p.sunfish.WriteParamHpp(config.Params.list(p.values)); err != nil {
				return err
			}

			if err := p.sunfish.BuildCSA(); err != nil {
				err = errors.Wrap(err, fmt.Sprintf("failed to start sunfish %s", p.name))
				log.Println(err)
				return err
			}

			return nil
		})
	}
	return eg.Wait()
}

func (m *Manager) startPlayers() error {
	m.procMutex.Lock()
	defer m.procMutex.Unlock()

	var eg errgroup.Group
	for _, _p := range append(m.newPlayers, m.currPlayers...) {
		p := _p
		eg.Go(func() error {
			if err := p.sunfish.StartCSA(); err != nil {
				err = errors.Wrap(err, fmt.Sprintf("failed to start sunfish %s", p.name))
				log.Println(err)
				return err
			}

			return nil
		})
	}
	return eg.Wait()
}

func stopPlayers(players []*player) {
	// Kill
	wg := &sync.WaitGroup{}
	for _, p := range players {
		wg.Add(1)
		go func(s *sunfish.Sunfish) {
			defer wg.Done()
			s.StopCSA()
		}(p.sunfish)
	}
	wg.Wait()

	// Clean
	for _, p := range players {
		p.sunfish.Clean()
	}
}

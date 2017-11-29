package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
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
	score   sunfish.Score
}

type Manager struct {
	Config Config

	server      *server.ShogiServer
	players     []*player
	normPlayers []*player
	scores      [][]scoreType
	eval        [][]float64
	procMutex   sync.Mutex
}

type scoreType struct {
	value int32
	win   float64
	loss  float64
}

func NewManager(config Config) *Manager {
	scores := make([][]scoreType, len(config.Params))
	for i := range scores {
		scores[i] = make([]scoreType, 0)
		param := config.Params[i]
		for val := param.MinimumValue; val <= param.MaximumValue; val += param.Step {
			scores[i] = append(scores[i], scoreType{
				value: val,
			})
		}
	}
	return &Manager{
		Config: config,
		scores: scores,
	}
}

func (m *Manager) Run() error {
	checkFiles()

	defer m.Destroy()

	err := m.start()
	if err != nil {
		return err
	}

	for gn := 1; ; gn++ {
		m.PrintGeneration(gn)

		time.Sleep(m.Config.Duration)

		err = m.next()
		if err != nil {
			log.Println(err)
		}
	}
}

func (m *Manager) start() error {
	m.procMutex.Lock()
	defer m.procMutex.Unlock()

	m.server = server.New()
	if err := m.server.Setup(); err != nil {
		return err
	}

	m.normPlayers = make([]*player, 0, m.Config.Concurrency)
	for i := 0; i < m.Config.Concurrency; i++ {
		name := fmt.Sprintf("n-%d", i)
		values := generateNormalValues(m.Config)
		m.normPlayers = append(m.normPlayers, m.generatePlayer(name, i, values))
	}

	m.players = make([]*player, 0, m.Config.Concurrency)
	for i := 0; i < m.Config.Concurrency; i++ {
		name := fmt.Sprintf("i-%d", i)
		values := generateRandomValues(m.Config)
		m.players = append(m.players, m.generatePlayer(name, i, values))
	}

	if err := setupPlayers(append(m.players, m.normPlayers...), m.Config); err != nil {
		return err
	}

	return startPlayers(append(m.players, m.normPlayers...))
}

func (m *Manager) next() error {
	m.procMutex.Lock()
	defer m.procMutex.Unlock()

	log.Println("Scores")
	var totalScore sunfish.Score
	for _, p := range m.players {
		if score, err := p.sunfish.GetScore(); err != nil {
			log.Println(err)
		} else {
			log.Printf("%s %d - %d\n", p.name, score.Win, score.Lose)
			p.score = score
			totalScore.Win += score.Win
			totalScore.Lose += score.Lose
		}
	}
	log.Printf("total %d - %d (%f)\n", totalScore.Win, totalScore.Lose, float64(totalScore.Win)/float64(totalScore.Win+totalScore.Lose))
	log.Println()

	// Update Scores
	m.updateScores()

	// Best Values
	bestValues, rates := m.getBestValues()
	log.Printf("best values %s\n", stringifyValues(bestValues))
	log.Printf("rates %s\n", stringifyRates(rates))
	log.Println()

	// New Generation
	players := make([]*player, 0, m.Config.Concurrency)
	var i int
	for ; i < m.Config.Concurrency*1/4; i++ {
		name := fmt.Sprintf("i-%d", i)
		players = append(players, m.generatePlayer(name, i, bestValues))
	}
	for ; i < m.Config.Concurrency*2/4; i++ {
		name := fmt.Sprintf("i-%d", i)
		values := m.mixSecondValue(bestValues)
		players = append(players, m.generatePlayer(name, i, values))
	}
	for ; i < m.Config.Concurrency; i++ {
		name := fmt.Sprintf("i-%d", i)
		values := m.mixRandomValue(bestValues)
		players = append(players, m.generatePlayer(name, i, values))
	}

	// Stop Previous Generation
	stopPlayers(append(m.players, m.normPlayers...))

	// Replace to New Generation
	m.players = players

	// Start Next Generation
	if err := setupPlayers(append(m.players, m.normPlayers...), m.Config); err != nil {
		log.Println(err)
	}

	if err := startPlayers(append(m.players, m.normPlayers...)); err != nil {
		log.Println(err)
	}

	return nil
}

func (m *Manager) generatePlayer(name string, gameNumber int, values []int32) *player {
	s := sunfish.New()
	s.Config.Directory = name
	s.CSAConfig.Pass = "test" + strconv.Itoa(gameNumber) + "-600-10,SunTest"
	s.CSAConfig.User = name
	return &player{
		name:    name,
		sunfish: s,
		values:  values,
	}
}

func (m *Manager) PrintGeneration(gn int) {
	log.Printf("Generation: %d\n", gn)
	for i := range m.players {
		log.Printf("%s %s\n", m.players[i].name, stringifyValues(m.players[i].values))
	}
	log.Println()
}

func (m *Manager) Destroy() {
	m.procMutex.Lock()
	defer m.procMutex.Unlock()

	stopPlayers(append(m.players, m.normPlayers...))
	m.server.Stop()
}

func (m *Manager) updateScores() {
	for i := range m.Config.Params {
		for j := range m.scores[i] {
			score := m.scores[i][j]
			score.win *= 0.99
			score.loss *= 0.99
			m.scores[i][j] = score
		}

		for _, p := range m.players {
			value := p.values[i]
			for j := range m.scores[i] {
				if m.scores[i][j].value == value {
					m.scores[i][j].win += float64(p.score.Win)
					m.scores[i][j].loss += float64(p.score.Lose)
				}
			}
		}
	}

	m.eval = make([][]float64, len(m.Config.Params))
	for i := range m.Config.Params {
		m.eval[i] = make([]float64, len(m.scores[i]))

		for j0 := range m.scores[i] {
			var win float64
			var loss float64
			for j1, score := range m.scores[i] {
				r := math.Pow(2, -math.Abs(float64(j0-j1)))
				win += score.win * r
				loss += score.loss * r
			}
			sum := win + loss

			if sum >= 1 {
				m.eval[i][j0] = win / sum
			} else {
				m.eval[i][j0] = rand.Float64() // [0.0,1.0)
			}
		}
	}
}

func (m *Manager) getBestValues() ([]int32, []float64) {
	values := make([]int32, len(m.Config.Params))
	rates := make([]float64, len(m.Config.Params))

	for i := range m.Config.Params {
		maxRate := float64(-1.0)
		for j, score := range m.scores[i] {
			if m.eval[i][j] > maxRate {
				values[i] = score.value
				maxRate = m.eval[i][j]
			}
		}
		rates[i] = maxRate
	}

	return values, rates
}

func (m *Manager) mixSecondValue(orgValues []int32) []int32 {
	values := make([]int32, len(m.Config.Params))

	target := int(rand.Int31n(int32(len(m.Config.Params))))
	for i := range m.Config.Params {
		if i == target {
			maxRate := float64(-1.0)
			for j, score := range m.scores[i] {
				if score.value != orgValues[i] && m.eval[i][j] > maxRate {
					values[i] = score.value
					maxRate = m.eval[i][j]
				}
			}
		} else {
			values[i] = orgValues[i]
		}
	}

	return values
}

func (m *Manager) mixRandomValue(orgValues []int32) []int32 {
	values := make([]int32, len(m.Config.Params))

	target := int(rand.Int31n(int32(len(m.Config.Params))))
	for i := range m.Config.Params {
		if i == target {
			r := rand.Int31n(int32(len(m.scores[i]) - 1))
			if m.scores[i][r].value >= orgValues[i] {
				r++
			}
			values[i] = m.scores[i][r].value
		} else {
			values[i] = orgValues[i]
		}
	}

	return values
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

func setupPlayers(players []*player, config Config) error {
	var eg errgroup.Group
	for _, _p := range players {
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

			if err := p.sunfish.WriteParamHpp(config.Params.toMap(p.values)); err != nil {
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

func startPlayers(players []*player) error {
	var eg errgroup.Group
	for _, _p := range players {
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

package main

import (
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	f, err := os.OpenFile("expt.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, f))

	params := []Param{
		{
			Name:   "EXT_SINGULAR",
			Normal: 2,
			Step:   1,
		},
		{
			Name:   "EXT_DEPTH_CHECK",
			Normal: 1,
			Step:   1,
		},
		{
			Name:   "EXT_DEPTH_ONE_REPLY",
			Normal: 3,
			Step:   1,
		},
		{
			Name:   "EXT_DEPTH_RECAP",
			Normal: 2,
			Step:   1,
		},
		{
			Name:   "NULL_DEPTH_RATE",
			Normal: 12,
			Step:   1,
		},
		{
			Name:   "NULL_DEPTH_REDUCE",
			Normal: 14,
			Step:   1,
		},
		{
			Name:   "NULL_DEPTH_VRATE",
			Normal: 475,
			Step:   5,
		},
		{
			Name:   "REDUCTION_RATE1",
			Normal: 8,
			Step:   1,
		},
		{
			Name:   "REDUCTION_RATE2",
			Normal: 12,
			Step:   1,
		},
		{
			Name:   "RAZOR_MARGIN1",
			Normal: 225,
			Step:   5,
		},
		{
			Name:   "RAZOR_MARGIN2",
			Normal: 675,
			Step:   5,
		},
		{
			Name:   "RAZOR_MARGIN3",
			Normal: 800,
			Step:   5,
		},
		{
			Name:   "RAZOR_MARGIN4",
			Normal: 670,
			Step:   5,
		},
		{
			Name:   "FUT_PRUN_MAX_DEPTH",
			Normal: 20,
			Step:   1,
		},
		{
			Name:   "FUT_PRUN_MARGIN_RATE",
			Normal: 62,
			Step:   2,
		},
		{
			Name:   "FUT_PRUN_MARGIN",
			Normal: 85,
			Step:   2,
		},
		{
			Name:   "PROBCUT_MARGIN",
			Normal: 219,
			Step:   2,
		},
		{
			Name:   "PROBCUT_REDUCTION",
			Normal: 4,
			Step:   1,
		},
		{
			Name:   "ASP_MIN_DEPTH",
			Normal: 7,
			Step:   1,
		},
		{
			Name:   "ASP_1ST_DELTA",
			Normal: 82,
			Step:   2,
		},
		{
			Name:   "ASP_DELTA_RATE",
			Normal: 23,
			Step:   2,
		},
		{
			Name:   "SINGULAR_DEPTH",
			Normal: 7,
			Step:   1,
		},
		{
			Name:   "SINGULAR_MARGIN",
			Normal: 22,
			Step:   1,
		},
	}
	config := Config{
		Params:      params,
		Concurrency: 16,
		Duration:    time.Hour * 1,
		Branch:      "new-feature",
		MoveLimit:   1,
	}

	m := NewManager(config)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-signalChan
		log.Println("signal:", s)
		m.Destroy()
		log.Fatal("exit")
	}()

	err = m.Run()
	if err != nil {
		log.Fatal(err)
	}
}

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
			Name:    "EXT_SINGULAR",
			Normal:  2,
			Minimum: 0,
			Maximum: 8,
			Step:    1,
		},
		{
			Name:    "EXT_DEPTH_CHECK",
			Normal:  3,
			Minimum: 0,
			Maximum: 8,
			Step:    1,
		},
		{
			Name:    "EXT_DEPTH_ONE_REPLY",
			Normal:  2,
			Minimum: 0,
			Maximum: 8,
			Step:    1,
		},
		{
			Name:    "EXT_DEPTH_RECAP",
			Normal:  2,
			Minimum: 0,
			Maximum: 8,
			Step:    1,
		},
		{
			Name:    "NULL_DEPTH_RATE",
			Normal:  12,
			Minimum: 0,
			Maximum: 16,
			Step:    2,
		},
		{
			Name:    "NULL_DEPTH_REDUCE",
			Normal:  14,
			Minimum: 0,
			Maximum: 24,
			Step:    2,
		},
		{
			Name:    "NULL_DEPTH_VRATE",
			Normal:  510,
			Minimum: 100,
			Maximum: 700,
			Step:    10,
		},
		{
			Name:    "REDUCTION_RATE1",
			Normal:  12,
			Minimum: 0,
			Maximum: 50,
			Step:    4,
		},
		{
			Name:    "REDUCTION_RATE2",
			Normal:  12,
			Minimum: 0,
			Maximum: 50,
			Step:    4,
		},
		{
			Name:    "RAZOR_MARGIN1",
			Normal:  230,
			Minimum: 100,
			Maximum: 1000,
			Step:    10,
		},
		{
			Name:    "RAZOR_MARGIN2",
			Normal:  700,
			Minimum: 100,
			Maximum: 1000,
			Step:    10,
		},
		{
			Name:    "RAZOR_MARGIN3",
			Normal:  820,
			Minimum: 100,
			Maximum: 1000,
			Step:    10,
		},
		{
			Name:    "RAZOR_MARGIN4",
			Normal:  660,
			Minimum: 100,
			Maximum: 1000,
			Step:    10,
		},
		{
			Name:    "FUT_PRUN_MAX_DEPTH",
			Normal:  24,
			Minimum: 0,
			Maximum: 60,
			Step:    4,
		},
		{
			Name:    "FUT_PRUN_MARGIN_RATE",
			Normal:  40,
			Minimum: 10,
			Maximum: 300,
			Step:    10,
		},
		{
			Name:    "FUT_PRUN_MARGIN",
			Normal:  70,
			Minimum: 0,
			Maximum: 300,
			Step:    10,
		},
		{
			Name:    "PROBCUT_MARGIN",
			Normal:  200,
			Minimum: 0,
			Maximum: 500,
			Step:    10,
		},
		{
			Name:    "PROBCUT_REDUCTION",
			Normal:  16,
			Minimum: 0,
			Maximum: 40,
			Step:    2,
		},
		{
			Name:    "ASP_MIN_DEPTH",
			Normal:  0,
			Minimum: 0,
			Maximum: 24,
			Step:    4,
		},
		{
			Name:    "ASP_1ST_DELTA",
			Normal:  70,
			Minimum: 40,
			Maximum: 240,
			Step:    10,
		},
		{
			Name:    "ASP_DELTA_RATE",
			Normal:  20,
			Minimum: 10,
			Maximum: 100,
			Step:    5,
		},
		{
			Name:    "SINGULAR_DEPTH",
			Normal:  36,
			Minimum: 28,
			Maximum: 60,
			Step:    4,
		},
		{
			Name:    "SINGULAR_MARGIN",
			Normal:  20,
			Minimum: 0,
			Maximum: 50,
			Step:    5,
		},
	}
	config := Config{
		Params:        params,
		Concurrency:   16,
		NumberOfGames: 2000,
		Branch:        "new-feature",
		MoveLimit:     1,
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

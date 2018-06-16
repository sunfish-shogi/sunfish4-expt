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
			Normal:  1,
			Minimum: 0,
			Maximum: 8,
			Step:    1,
		},
		{
			Name:    "EXT_DEPTH_ONE_REPLY",
			Normal:  3,
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
			Normal:  480,
			Minimum: 100,
			Maximum: 700,
			Step:    10,
		},
		{
			Name:    "REDUCTION_RATE1",
			Normal:  8,
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
			Normal:  220,
			Minimum: 100,
			Maximum: 1000,
			Step:    10,
		},
		{
			Name:    "RAZOR_MARGIN2",
			Normal:  670,
			Minimum: 100,
			Maximum: 1000,
			Step:    10,
		},
		{
			Name:    "RAZOR_MARGIN3",
			Normal:  800,
			Minimum: 100,
			Maximum: 1000,
			Step:    10,
		},
		{
			Name:    "RAZOR_MARGIN4",
			Normal:  670,
			Minimum: 100,
			Maximum: 1000,
			Step:    10,
		},
		{
			Name:    "FUT_PRUN_MAX_DEPTH",
			Normal:  20,
			Minimum: 0,
			Maximum: 60,
			Step:    4,
		},
		{
			Name:    "FUT_PRUN_MARGIN_RATE",
			Normal:  60,
			Minimum: 10,
			Maximum: 300,
			Step:    10,
		},
		{
			Name:    "FUT_PRUN_MARGIN",
			Normal:  80,
			Minimum: 0,
			Maximum: 300,
			Step:    10,
		},
		{
			Name:    "PROBCUT_MARGIN",
			Normal:  220,
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
			Normal:  8,
			Minimum: 0,
			Maximum: 24,
			Step:    4,
		},
		{
			Name:    "ASP_1ST_DELTA",
			Normal:  80,
			Minimum: 40,
			Maximum: 240,
			Step:    10,
		},
		{
			Name:    "ASP_DELTA_RATE",
			Normal:  25,
			Minimum: 10,
			Maximum: 100,
			Step:    5,
		},
		{
			Name:    "SINGULAR_DEPTH",
			Normal:  32,
			Minimum: 28,
			Maximum: 60,
			Step:    4,
		},
		{
			Name:    "SINGULAR_MARGIN",
			Normal:  15,
			Minimum: 0,
			Maximum: 50,
			Step:    5,
		},
	}
	config := Config{
		Params:        params,
		Concurrency:   16,
		NumberOfGames: 500,
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

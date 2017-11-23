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
			Name:         "EXT_SINGULAR",
			Normal:       4,
			MinimumValue: 0,
			MaximumValue: 4,
			Step:         1,
		},
		{
			Name:         "EXT_DEPTH_CHECK",
			Normal:       2,
			MinimumValue: 0,
			MaximumValue: 4,
			Step:         1,
		},
		{
			Name:         "EXT_DEPTH_ONE_REPLY",
			Normal:       3,
			MinimumValue: 0,
			MaximumValue: 4,
			Step:         1,
		},
		{
			Name:         "EXT_DEPTH_RECAP",
			Normal:       2,
			MinimumValue: 0,
			MaximumValue: 4,
			Step:         1,
		},
		{
			Name:         "NULL_DEPTH_RATE",
			Normal:       10,
			MinimumValue: 4,
			MaximumValue: 16,
			Step:         1,
		},
		{
			Name:         "NULL_DEPTH_REDUCE",
			Normal:       10,
			MinimumValue: 0,
			MaximumValue: 20,
			Step:         1,
		},
		{
			Name:         "NULL_DEPTH_VRATE",
			Normal:       681,
			MinimumValue: 50,
			MaximumValue: 800,
			Step:         50,
		},
		{
			Name:         "REDUCTION_RATE1",
			Normal:       9,
			MinimumValue: 5,
			MaximumValue: 30,
			Step:         5,
		},
		{
			Name:         "REDUCTION_RATE2",
			Normal:       11,
			MinimumValue: 5,
			MaximumValue: 30,
			Step:         5,
		},
		{
			Name:         "RAZOR_MARGIN1",
			Normal:       389,
			MinimumValue: 50,
			MaximumValue: 800,
			Step:         50,
		},
		{
			Name:         "RAZOR_MARGIN2",
			Normal:       705,
			MinimumValue: 50,
			MaximumValue: 800,
			Step:         50,
		},
		{
			Name:         "RAZOR_MARGIN3",
			Normal:       790,
			MinimumValue: 50,
			MaximumValue: 800,
			Step:         50,
		},
		{
			Name:         "RAZOR_MARGIN4",
			Normal:       796,
			MinimumValue: 50,
			MaximumValue: 800,
			Step:         50,
		},
		{
			Name:         "FUT_PRUN_MAX_DEPTH",
			Normal:       28,
			MinimumValue: 4,
			MaximumValue: 64,
			Step:         4,
		},
		{
			Name:         "FUT_PRUN_MARGIN_RATE",
			Normal:       57,
			MinimumValue: 20,
			MaximumValue: 200,
			Step:         20,
		},
		{
			Name:         "FUT_PRUN_MARGIN",
			Normal:       136,
			MinimumValue: 50,
			MaximumValue: 800,
			Step:         25,
		},
		{
			Name:         "PROBCUT_MARGIN",
			Normal:       423,
			MinimumValue: 50,
			MaximumValue: 500,
			Step:         25,
		},
		{
			Name:         "PROBCUT_REDUCTION",
			Normal:       1,
			MinimumValue: 1,
			MaximumValue: 10,
			Step:         1,
		},
		{
			Name:         "ASP_MIN_DEPTH",
			Normal:       6,
			MinimumValue: 2,
			MaximumValue: 10,
			Step:         1,
		},
		{
			Name:         "ASP_1ST_DELTA",
			Normal:       128,
			MinimumValue: 32,
			MaximumValue: 256,
			Step:         32,
		},
		{
			Name:         "ASP_DELTA_RATE",
			Normal:       50,
			MinimumValue: 25,
			MaximumValue: 200,
			Step:         25,
		},
		{
			Name:         "SINGULAR_DEPTH",
			Normal:       10,
			MinimumValue: 4,
			MaximumValue: 12,
			Step:         1,
		},
		{
			Name:         "SINGULAR_MARGIN",
			Normal:       32,
			MinimumValue: 1,
			MaximumValue: 32,
			Step:         1,
		},
	}
	config := Config{
		Params:      params,
		Concurrency: 14,
		Duration:    time.Minute * 10,
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

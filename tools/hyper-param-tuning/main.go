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
			Normal:       2,
			MinimumValue: 1,
			MaximumValue: 3,
			Step:         1,
		},
		{
			Name:         "EXT_DEPTH_CHECK",
			Normal:       2,
			MinimumValue: 1,
			MaximumValue: 3,
			Step:         1,
		},
		{
			Name:         "EXT_DEPTH_ONE_REPLY",
			Normal:       2,
			MinimumValue: 1,
			MaximumValue: 3,
			Step:         1,
		},
		{
			Name:         "EXT_DEPTH_RECAP",
			Normal:       3,
			MinimumValue: 2,
			MaximumValue: 4,
			Step:         1,
		},
		{
			Name:         "NULL_DEPTH_RATE",
			Normal:       12,
			MinimumValue: 10,
			MaximumValue: 14,
			Step:         1,
		},
		{
			Name:         "NULL_DEPTH_REDUCE",
			Normal:       12,
			MinimumValue: 10,
			MaximumValue: 14,
			Step:         1,
		},
		{
			Name:         "NULL_DEPTH_VRATE",
			Normal:       500,
			MinimumValue: 475,
			MaximumValue: 525,
			Step:         5,
		},
		{
			Name:         "REDUCTION_RATE1",
			Normal:       10,
			MinimumValue: 8,
			MaximumValue: 12,
			Step:         1,
		},
		{
			Name:         "REDUCTION_RATE2",
			Normal:       10,
			MinimumValue: 8,
			MaximumValue: 12,
			Step:         1,
		},
		{
			Name:         "RAZOR_MARGIN1",
			Normal:       250,
			MinimumValue: 225,
			MaximumValue: 275,
			Step:         5,
		},
		{
			Name:         "RAZOR_MARGIN2",
			Normal:       650,
			MinimumValue: 625,
			MaximumValue: 675,
			Step:         5,
		},
		{
			Name:         "RAZOR_MARGIN3",
			Normal:       800,
			MinimumValue: 775,
			MaximumValue: 825,
			Step:         5,
		},
		{
			Name:         "RAZOR_MARGIN4",
			Normal:       650,
			MinimumValue: 625,
			MaximumValue: 675,
			Step:         5,
		},
		{
			Name:         "FUT_PRUN_MAX_DEPTH",
			Normal:       20,
			MinimumValue: 15,
			MaximumValue: 25,
			Step:         1,
		},
		{
			Name:         "FUT_PRUN_MARGIN_RATE",
			Normal:       60,
			MinimumValue: 50,
			MaximumValue: 70,
			Step:         2,
		},
		{
			Name:         "FUT_PRUN_MARGIN",
			Normal:       75,
			MinimumValue: 65,
			MaximumValue: 85,
			Step:         2,
		},
		{
			Name:         "PROBCUT_MARGIN",
			Normal:       225,
			MinimumValue: 205,
			MaximumValue: 245,
			Step:         2,
		},
		{
			Name:         "PROBCUT_REDUCTION",
			Normal:       4,
			MinimumValue: 3,
			MaximumValue: 5,
			Step:         1,
		},
		{
			Name:         "ASP_MIN_DEPTH",
			Normal:       6,
			MinimumValue: 5,
			MaximumValue: 7,
			Step:         1,
		},
		{
			Name:         "ASP_1ST_DELTA",
			Normal:       64,
			MinimumValue: 44,
			MaximumValue: 84,
			Step:         2,
		},
		{
			Name:         "ASP_DELTA_RATE",
			Normal:       25,
			MinimumValue: 15,
			MaximumValue: 35,
			Step:         2,
		},
		{
			Name:         "SINGULAR_DEPTH",
			Normal:       6,
			MinimumValue: 5,
			MaximumValue: 7,
			Step:         1,
		},
		{
			Name:         "SINGULAR_MARGIN",
			Normal:       20,
			MinimumValue: 18,
			MaximumValue: 22,
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

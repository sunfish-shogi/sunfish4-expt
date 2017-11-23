package main

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Param struct {
	Name         string
	Normal       int32
	MinimumValue int32
	MaximumValue int32
	Step         int32
}

type Params []Param

type Config struct {
	Params      Params
	Concurrency int
	Duration    time.Duration
}

func generateNormalValues(config Config) []int32 {
	values := make([]int32, len(config.Params))
	for i := range config.Params {
		values[i] = config.Params[i].Normal
	}
	return values
}

func generateRandomValues(config Config) []int32 {
	values := make([]int32, len(config.Params))
	for i := range config.Params {
		min := config.Params[i].MinimumValue
		max := config.Params[i].MaximumValue
		step := config.Params[i].Step
		values[i] = min + rand.Int31n((max-min+1)/step)*step
	}
	return values
}

func stringifyValues(values []int32) string {
	ss := make([]string, len(values))
	for vi, v := range values {
		ss[vi] = strconv.Itoa(int(v))
	}
	return "[" + strings.Join(ss, ",") + "]"
}

func stringifyRates(rates []float64) string {
	ss := make([]string, len(rates))
	for vi, v := range rates {
		ss[vi] = strconv.FormatFloat(v, 'f', 2, 64)
	}
	return "[" + strings.Join(ss, ",") + "]"
}

func (p Params) toMap(values []int32) map[string]int32 {
	m := make(map[string]int32, len(p))
	for i := range p {
		m[p[i].Name] = values[i]
	}
	return m
}

package main

import (
	"strconv"
	"strings"
	"time"
)

type Param struct {
	Name   string
	Normal int32
	Step   int32
}

type Params []Param

type Config struct {
	Params      Params
	Concurrency int
	Duration    time.Duration
	Branch      string
	MoveLimit   int
}

func generateNormalValues(config Config) []int32 {
	values := make([]int32, len(config.Params))
	for i := range config.Params {
		values[i] = config.Params[i].Normal
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

func (p Params) toMap(values []int32) map[string]int32 {
	m := make(map[string]int32, len(p))
	for i := range p {
		m[p[i].Name] = values[i]
	}
	return m
}

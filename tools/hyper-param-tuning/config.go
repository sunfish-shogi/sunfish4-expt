package main

import (
	"strconv"
	"strings"

	"github.com/sunfish-shogi/sunfish4-expt/sunfish"
)

type Param struct {
	Name    string
	Normal  int32
	Minimum int32
	Maximum int32
	Step    int32
}

type Params []Param

type Config struct {
	Params        Params
	Concurrency   int
	NumberOfGames int
	Branch        string
	MoveLimit     int
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

func (p Params) list(values []int32) []sunfish.SearchParam {
	m := make([]sunfish.SearchParam, len(p))
	for i := range p {
		m[i].Name = p[i].Name
		m[i].Value = values[i]
	}
	return m
}

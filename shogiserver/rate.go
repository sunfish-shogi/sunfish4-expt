package shogiserver

import "gopkg.in/yaml.v2"

type Player struct {
	Name         string  `yaml:"name"`
	RatingGroup  int     `yaml:"rating_group"`
	Rate         float64 `yaml:"rate"`
	LastModified string  `yaml:"last_modified"`
	Win          float64 `yaml:"win"`
	Loss         float64 `yaml:"loss"`
}

type Rate struct {
	Players map[int]map[string]Player `yaml:"players"`
}

func UnmarshalRate(b []byte) (Rate, error) {
	var rate Rate
	err := yaml.Unmarshal(b, &rate)
	return rate, err
}

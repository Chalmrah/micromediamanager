package main

import (
	"encoding/json"
	"os"
)

type Show struct {
	FileName      string `json:"FileName"`
	MappingFolder string `json:"MappingFolder"`
	Season        int    `json:"Season"`
}

func ReadConfig(fileName string) ([]Show, error) {
	byteValue, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var shows []Show
	if err := json.Unmarshal(byteValue, &shows); err != nil {
		return nil, err
	}

	for i := range shows {
		shows[i].applyDefaults()
	}

	return shows, nil
}

func (s *Show) applyDefaults() {
	if 0 == s.Season {
		s.Season = 1
	}

}

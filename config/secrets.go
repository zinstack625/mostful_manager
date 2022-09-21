package config

import (
	"bufio"
	"encoding/json"
	"os"
)

type integrationTokens struct {
	AddMentor    string `json:"add_mentor"`
	RemoveMentor string `json:"remove_mentor"`
	CheckMe      string `json:"check_me"`
	Labs         string `json:"labs"`
	SetName      string `json:"set_name"`
}

func (i *integrationTokens) Init(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return nil
	}
	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)
	return decoder.Decode(i)
}

var IntegrationTokens integrationTokens

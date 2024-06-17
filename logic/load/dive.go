package load

import (
	"encoding/json"
	"os"
)

type DiveJSON struct {
	Layers []struct {
		Index     int    `json:"index"`
		ID        string `json:"id"`
		DigestID  string `json:"digestId"`
		SizeBytes int64  `json:"sizeBytes"` // to represent GB sizes
		Command   string `json:"command"`
	} `json:"layer"`
}

func LoadDiveFile(file string) (*DiveJSON, error) {
	diveData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var diveOutput DiveJSON
	if err = json.Unmarshal(diveData, &diveOutput); err != nil {
		return nil, err
	}

	return &diveOutput, nil
}

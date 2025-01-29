package storage

import (
	"encoding/base64"
	"encoding/json"

	"github.com/mediaprodcast/storage/pkg/storage/defs"
)

// Encode converts a StorageConfig struct into a Base64-s string.
func Encode(cfg *defs.StorageConfig) (string, error) {
	// Marshal the struct into a JSON byte slice.
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}

	// Encode the JSON byte slice into a Base64 string.
	s := base64.StdEncoding.EncodeToString(jsonData)
	return s, nil
}

// Decode converts a Base64-string string back into a StorageConfig struct.
func Decode(s string) (*defs.StorageConfig, error) {
	var config *defs.StorageConfig

	// Decode the Base64 string back into a JSON byte slice.
	jsonData, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return config, err
	}

	// Unmarshal the JSON byte slice back into the StorageConfig struct.
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

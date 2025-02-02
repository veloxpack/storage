package credentials

import (
	"encoding/base64"

	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Encode converts a StorageConfig struct into a Base64-s string.
func Encode(cfg *pb.StorageConfig) (string, error) {
	// Marshal the struct into a JSON byte slice.
	jsonData, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(cfg)
	if err != nil {
		return "", err
	}

	// Encode the JSON byte slice into a Base64 string.
	s := base64.StdEncoding.EncodeToString(jsonData)
	return s, nil
}

// Decode converts a Base64-string string back into a StorageConfig struct.
func Decode(s string) (*pb.StorageConfig, error) {
	var config pb.StorageConfig

	// Decode the Base64 string back into a JSON byte slice.
	jsonData, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err // Return nil on error
	}

	// Unmarshal the JSON byte slice back into the StorageConfig struct.
	err = protojson.Unmarshal(jsonData, &config)
	if err != nil {
		return nil, err // Return nil on error
	}

	return &config, nil // Return pointer to StorageConfig
}

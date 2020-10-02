package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/mctofu/homekit/client"
)

// ControllerConfig details a controller and its pairings
type ControllerConfig struct {
	Name              string
	DeviceID          string
	PublicKey         []byte
	PrivateKey        []byte
	AccessoryPairings []*AccessoryPairing
}

// AccessoryPairing details a paired accessory
type AccessoryPairing struct {
	Name             string
	DeviceID         string
	PublicKey        []byte
	IPConnectionInfo client.IPConnectionInfo
}

// ReadControllerConfig reads the ControllerConfig for a controller with the given name
// stored under the configPath directory
func ReadControllerConfig(configPath string, name string) (*ControllerConfig, error) {
	data, err := ioutil.ReadFile(path.Join(configPath, name+".json"))
	if err != nil {
		return nil, err
	}

	var cfg ControllerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %v", err)
	}

	return &cfg, nil
}

// SaveControllerConfig writes the ControllerConfig for a controller with the given name
// under the configPath directory. An error is returned if a config for the controller
// already exists unless the overwrite flag is set.
func SaveControllerConfig(configPath string, cfg *ControllerConfig, overwrite bool) error {
	filePath := path.Join(configPath, cfg.Name+".json")
	if !overwrite {
		if _, err := os.Stat(filePath); err == nil {
			return fmt.Errorf("%s already exists", filePath)
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	if err := os.MkdirAll(configPath, 0700); err != nil {
		return err
	}

	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	if err := WriteConfig(filePath, output.Bytes()); err != nil {
		return err
	}

	return nil
}

// WriteConfig writes data to path but first writes it to a temporary file
// to avoid errors where the write fails & the data from the original config
// is lost.
func WriteConfig(path string, data []byte) error {
	tempPath := path + ".new"
	if err := ioutil.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("write temp: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove existing: %v", err)
		}
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("rename: %v", err)
	}

	return nil
}

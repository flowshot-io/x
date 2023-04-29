package config

import (
	"fmt"
	"os"
	"path"

	"sigs.k8s.io/yaml"

	"github.com/go-playground/validator/v10"
)

const (
	defaultConfigDir  = "config"
	defaultConfigName = "settings.yaml"
)

func Load(configDir string, name string, config interface{}) error {
	if configDir == "" {
		configDir = defaultConfigDir
	}

	if name == "" {
		name = defaultConfigName
	}

	f, err := getConfigFile(configDir, name)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(f)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}

	validate := validator.New()
	return validate.Struct(config)
}

func getConfigFile(configDir string, name string) (string, error) {
	path := path.Join(configDir, name)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("config file %s does not exist", path)
	}

	return path, nil
}

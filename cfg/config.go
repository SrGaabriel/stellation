package cfg

import (
	"os"

	"github.com/BurntSushi/toml"
)

const configFile = "stellation.toml"

type Config struct {
	Name    string
	Path    string
	Website *string
}

func DefaultConfig() Config {
	return Config{
		Name:    "My Modpack",
		Path:    "./pack",
		Website: nil,
	}
}

func SaveConfig(conf Config) {
	var file, err = os.Create(configFile)
	if err != nil {
		panic("Failed to create config file ./stellation.toml: " + err.Error())
	}
	defer file.Close()
	encoder := toml.NewEncoder(file)
	err = encoder.Encode(conf)
	if err != nil {
		panic("Failed to encode config: " + err.Error())
	}
}

func GetConfig() {
	var file, err = os.ReadFile(configFile)
	if err != nil {
		panic("Failed to read config file ./stellation.toml. Initialize the project with 'stellation init': " + err.Error())
	}
	var conf Config
	_, tomlErr := toml.Decode(string(file), &conf)
	if tomlErr != nil {
		panic("Failed to decode config: " + tomlErr.Error())
	}
}

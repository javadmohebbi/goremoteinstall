package agent

import (
	"log"

	"github.com/spf13/viper"
)

type GriConfig struct {
	Time string

	DeployerAddress string
	DeployerPort    int

	Host   string
	HostID string

	TaskID     string
	ComputerID string

	IP           string
	ComputerName string

	// Command to run
	Bootstrap string

	// command line options
	Params []string

	Dir string
}

const (
	config_name = "griAgent"
	config_ext  = "json"
)

func NewConfig(dir_path string) (*GriConfig, error) {
	var conf GriConfig

	viper.SetConfigName(config_name)
	viper.SetConfigType(config_ext)
	// viper.SetConfigType(config_ext)
	viper.AddConfigPath(dir_path)

	log.Printf("Reading config from '%s\\%s.%s'", dir_path, config_name, config_ext)
	if err := viper.ReadInConfig(); err != nil {
		// if file not found set default fallback values
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Logfile %v.%v not found in path %v", dir_path, config_name, config_ext)
			return &GriConfig{}, nil
		}
		log.Fatalln("Could not read config file due to error: ", err)
		return &GriConfig{}, err
	}

	// Unmarshal configs
	err := viper.Unmarshal(&conf)

	// check foe unmarshal errors
	if err != nil {
		log.Println("Unable to unmarshal configs into struct due to error: ", err)
		return &GriConfig{}, err
	}

	return &conf, nil

}

package goremoteinstall

import (
	"log"

	"github.com/spf13/viper"
)

type ServerTaskConfig struct {

	// current server deployer
	DeployerAddress string
	DeployerPort    int

	Concurrent int
	Timeout    int64

	Username string
	Domain   string
	Password string

	Name      string
	desc      string
	Targets   string
	Files     string
	Bootstrap string
	Params    []string
}

const (
	config_server_task_ext = "yaml"
)

func NewServerTaskConfig(dir_path, filename string) (*ServerTaskConfig, error) {
	var conf ServerTaskConfig

	viper.SetConfigName(filename)
	viper.SetConfigType(config_server_task_ext)
	viper.AddConfigPath(dir_path)

	log.Printf("Reading config from '%s/%s'", dir_path, filename)
	if err := viper.ReadInConfig(); err != nil {
		// if file not found set default fallback values
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Logfile %v.%v not found in path %v", dir_path, filename, config_server_task_ext)
			return &ServerTaskConfig{}, nil
		}
		log.Fatalln("Could not read config file due to error: ", err)
		return &ServerTaskConfig{}, err
	}

	// Unmarshal configs
	err := viper.Unmarshal(&conf)

	// check foe unmarshal errors
	if err != nil {
		log.Println("Unable to unmarshal configs into struct due to error: ", err)
		return &ServerTaskConfig{}, err
	}

	return &conf, nil

}

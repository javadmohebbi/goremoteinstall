package goremoteinstall

import (
	"log"

	"github.com/spf13/viper"
)

type ServerGlobalConfig struct {

	// current server deployer
	DeployerListen  string
	DeployerAddress string
	DeployerPort    int

	AgentPath string

	// Concurrent int
	// Timeout    int64

	DeployerSocket string
}

const (
	config_global_name = "griServer"
	config_gloabl_ext  = "yaml"
)

func NewGloablConfig(dir_path string) (*ServerGlobalConfig, error) {
	var conf ServerGlobalConfig

	viper.SetConfigName(config_global_name)
	viper.SetConfigType(config_gloabl_ext)
	viper.AddConfigPath(dir_path)

	log.Printf("Reading config from '%s%s.%s'", dir_path, config_global_name, config_gloabl_ext)
	if err := viper.ReadInConfig(); err != nil {
		// if file not found set default fallback values
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Logfile %v.%v not found in path %v", dir_path, config_global_name, config_gloabl_ext)
			return &ServerGlobalConfig{}, nil
		}
		log.Fatalln("Could not read config file due to error: ", err)
		return &ServerGlobalConfig{}, err
	}

	// Unmarshal configs
	err := viper.Unmarshal(&conf)

	// check foe unmarshal errors
	if err != nil {
		log.Println("Unable to unmarshal configs into struct due to error: ", err)
		return &ServerGlobalConfig{}, err
	}

	return &conf, nil

}

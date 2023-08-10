package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

/*
This pacakge generates goroxyconfig.json template file.
and also reads and parses the file.
*/

type Config struct {
	Port   int    `json:"port,omitempty"`
	Target string `json:"target,omitempty"`
}

// ReadConfig takes command line arguments to reads and set the configuration.
// After parsing, it returns the configuration.
// If no arguments are passed, it reads the default config file.
// If there is no config file, it generates the config template file.
// "args" must not include the program name.
func ReadConfig(args []string) (*Config, error) {
	var err error
	conf := &Config{
		Port:   0,
		Target: "",
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		// port config, which specifies the port number to listen to.
		case "-p", "--port":
			if i+1 >= len(args) {
				break
			}
			conf.Port, err = strconv.Atoi(args[i+1])
			if err != nil {
				break
			}
			i++
			continue
		// target config, which specifies the target http server.
		case "-t", "--target":
			if i+1 >= len(args) {
				break
			}
			conf.Target = args[i+1]
			i++
			continue
		}
		return nil, fmt.Errorf("invalid argument %s", args[i])
	}
	// if no arguments are passsed, read the default config file.
	if conf.Port == 0 && conf.Target == "" {
		return ReadOrGenConfFile(conf)
	}
	return conf, nil
}

// ReadOrGenConfFile handles the case where required arguments are not passed.
// It first tries reading the default config file.
// If the file doesn't exist, it generates the config template file.
func ReadOrGenConfFile(cf *Config) (*Config, error) {
	// first read the default config file.
	confB, err := os.ReadFile("goroxyconfig.json")
	// if doesn't exist, generate the config template.
	if os.IsNotExist(err) {
		return nil, GenerateConfigTemplate()
	} else if err != nil {
		return nil, fmt.Errorf("failed to read config file:%v", err)
	}

	confI := make(map[string]interface{})
	err = json.Unmarshal(confB, &confI)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshalling the config file:%v", err)
	}

	readPort, ok1 := confI["port"].(float64)
	readTarget, ok2 := confI["target"].(string)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("invalid config file")
	}

	cf.Port = int(readPort)
	cf.Target = readTarget
	return cf, nil
}

// goroxyConfigTemplate is the template for the config file,
// which is generated by GenerateConfigTemplate().
var goroxyConfigTemplate = `{
	"port": 8888,
	"target": "https://www.example.com"
}`

// GenerateConfigTemplate generates the config template json file.
func GenerateConfigTemplate() error {
	err := os.WriteFile("goroxyconfig.json", []byte(goroxyConfigTemplate), 0644)
	if err != nil {
		return fmt.Errorf("failed to generate config file template:%v", err)
	}
	return fmt.Errorf("config file template generated")
}

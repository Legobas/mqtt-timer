package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	CONFIG_FILE = "mqtt-timer.yml"
	CONFIG_DIR  = ".config"
	CONFIG_ROOT = "/config"
)

type Mqtt struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Qos      int    `yaml:"qos"`
	Retain   bool   `yaml:"retain"`
}

type Timer struct {
	Id           string `yaml:"id"`
	Description  string `yaml:"description"`
	Cron         string `yaml:"cron"`
	Time         string `yaml:"time"`
	Days         string `yaml:"days"`
	Before       string `yaml:"before"`
	After        string `yaml:"after"`
	RandomBefore string `yaml:"randomBefore"`
	RandomAfter  string `yaml:"randomAfter"`
	Topic        string `yaml:"topic"`
	Message      string `yaml:"message"`
	Enabled      bool
}

type Config struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`

	Mqtt   Mqtt    `yaml:"mqtt"`
	Timers []Timer `yaml:"timers"`
}

func getConfig() Config {
	var config Config

	configFile := filepath.Join(CONFIG_ROOT, CONFIG_FILE)
	msg := configFile
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		homedir, _ := os.UserHomeDir()
		configFile := filepath.Join(homedir, CONFIG_DIR, CONFIG_FILE)
		msg += ", " + configFile
		data, err = ioutil.ReadFile(configFile)
	}
	if err != nil {
		workingdir, _ := os.Getwd()
		configFile := filepath.Join(workingdir, CONFIG_FILE)
		msg += ", " + configFile
		data, err = ioutil.ReadFile(configFile)
	}
	if err != nil {
		msg = "Configuration file could not be found: " + msg
		log.Fatal(msg)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(config.Timers); i++ {
		config.Timers[i].Enabled = true
	}

	err = validate(config)
	if err != nil {
		log.Fatal(err)
	}

	// log.Printf("%+v\n", config)
	return config
}

func validate(config Config) error {
	if config.Mqtt.Url == "" {
		return errors.New("Config error: MQTT Server URL is mandatory")
	}
	for _, timer := range config.Timers {
		if timer.Id == "" {
			return errors.New("Config error: timer.id is mandatory")
		}
		if timer.Cron == "" && timer.Time == "" {
			return fmt.Errorf("Config error: timer.cron or timer.time is mandatory (timer %s)", timer.Id)
		}
		if timer.Cron != "" && timer.Time != "" {
			return fmt.Errorf("Config error: use only timer.cron or timer.time (timer %s)", timer.Id)
		}
		if timer.Cron != "" && timer.Before != "" {
			return fmt.Errorf("Config error: timer.before cannot be used with cron (timer %s)", timer.Id)
		}
		if timer.Cron != "" && timer.RandomBefore != "" {
			return fmt.Errorf("Config error: timer.randomBefore cannot be used with cron (timer %s)", timer.Id)
		}
		if timer.Before != "" {
			if timer.RandomBefore != "" || timer.After != "" || timer.RandomAfter != "" {
				return fmt.Errorf("Config error: only one of before, randomBefore, after or randomAfter can be used (timer %s)", timer.Id)
			}
		}
		if timer.RandomBefore != "" {
			if timer.Before != "" || timer.After != "" || timer.RandomAfter != "" {
				return fmt.Errorf("Config error: only one of before,randomBefore, after or randomAfter can be used (timer %s)", timer.Id)
			}
		}
		if timer.After != "" {
			if timer.Before != "" || timer.RandomBefore != "" || timer.RandomAfter != "" {
				return fmt.Errorf("Config error: only one of before, randomBefore, after or randomAfter can be used (timer %s)", timer.Id)
			}
		}
		if timer.RandomAfter != "" {
			if timer.Before != "" || timer.RandomBefore != "" || timer.After != "" {
				return fmt.Errorf("Config error: only one of before, randomBefore, after or randomAfter can be used (timer %s)", timer.Id)
			}
		}
	}

	return nil
}

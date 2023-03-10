package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	TIMEOUT   time.Duration = time.Second * 10
	SUBSCRIBE               = APPNAME + "/set"
)

type SetTimer struct {
	Id          string      `json:"id"`
	Description string      `json:"description"`
	Start       string      `json:"start"`
	Interval    string      `json:"interval"`
	Until       string      `json:"until"`
	Topic       string      `json:"topic"`
	Message     interface{} `json:"message"`
	Enabled     *bool       `json:"enabled,omitempty"`
}

var mqttClient MQTT.Client

func sendToMtt(topic string, message string) {
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), config.Mqtt.Retain, message)
}

func sendToMttRetain(topic string, message string) {
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), true, message)
}

func receive(client MQTT.Client, msg MQTT.Message) {
	message := string(msg.Payload()[:])

	var setTimer SetTimer
	err := json.Unmarshal([]byte(message), &setTimer)
	if err != nil {
		log.Printf("JSON Error: %s", err.Error())
		return
	}

	err = validateMessage(setTimer)
	if err != nil {
		log.Printf("Message error: %s", err.Error())
		return
	}

	// check config
	inConfig := false
	for _, timer := range config.Timers {
		if timer.Id == setTimer.Id {
			inConfig = true
			if setTimer.Enabled != nil {
				timer.Enabled = *setTimer.Enabled
				// only enable/disable
				log.Printf("Set timer: '%s' enabled: %t", setTimer.Id, timer.Enabled)
				return
			}
		}
	}

	if inConfig {
		log.Printf("Error timer '%s' defined in config", setTimer.Id)
		return
	} else {
		scheduler.RemoveByTag(setTimer.Id)
	}

	var messages []string

	if setTimer.Message != nil {
		switch setTimer.Message.(type) {
		case string:
			messages = append(messages, setTimer.Message.(string))
		case []interface{}:
			msgArray := setTimer.Message.([]interface{})
			for _, message := range msgArray {
				messages = append(messages, message.(string))
			}
		default:
			log.Printf("Error: %s", fmt.Sprint(setTimer.Message))
		}
	}

	startTime := time.Now().Local().Add(time.Duration(int64(1000000000)))
	err = errors.New("")

	if setTimer.Start != "" {
		matchTime, _ := regexp.Match("^\\d{1,2}(:\\d{2}){1,2}$", []byte(setTimer.Start))
		if matchTime {
			startTime, err = time.Parse("15:04", setTimer.Start)
			if err != nil {
				startTime, err = time.Parse("15:04:05", setTimer.Start)
				if err != nil {
					log.Printf("Error: invalid time format: %s", setTimer.Start)
					return
				}
			}
		} else if strings.Contains(setTimer.Start, "sec") || strings.Contains(setTimer.Start, "min") {
			seconds := parseSeconds(setTimer.Start)
			if seconds > 0 {
				offset := time.Duration(int64(seconds) * int64(1000000000))
				startTime = time.Now().Local().Add(offset)
			} else {
				log.Printf("Invalid duration: %s", setTimer.Start)
				return
			}
		} else {
			log.Printf("Invalid start time: %s", setTimer.Start)
			return
		}
	}

	// default 30 sec.
	offset := time.Duration(int64(30) * int64(1000000000))
	if setTimer.Interval != "" {
		seconds := parseSeconds(setTimer.Interval)
		if seconds > 0 {
			offset = time.Duration(int64(seconds) * int64(1000000000))
		} else {
			log.Printf("Invalid interval duration: %s", setTimer.Interval)
			return
		}
	} else {
		if len(messages) > 1 {
			log.Println("Warning: no interval set, default interval is 30 seconds")
		}
	}

	for _, message := range messages {
		timer := Timer{}
		timer.Enabled = true
		timer.Id = setTimer.Id
		timer.Description = fmt.Sprintf("%s [%s]", setTimer.Description, message)
		timer.Time = startTime.Format("15:04:05")
		timer.Topic = setTimer.Topic
		timer.Message = message
		job, err := scheduler.Every(1).Day().At(startTime).Tag(timer.Id).Do(handleEvent, &timer)
		if err != nil {
			log.Printf("Scheduler Error: %s", err.Error())
			return
		}
		job.LimitRunsTo(1)
		startTime = startTime.Add(offset)
	}
}

func GetClientId() string {
	hostname, _ := os.Hostname()
	return APPNAME + "_" + hostname
}

func connLostHandler(c MQTT.Client, err error) {
	log.Fatal(err)
}

func validateMessage(msg SetTimer) error {
	if msg.Id == "" {
		return errors.New("Error: id is mandatory")
	}
	if msg.Until != "" && msg.Interval == "" {
		return errors.New("Error: interval must have a value if until is specified")
	}

	return nil
}

func startMqttClient() {
	opts := MQTT.NewClientOptions().AddBroker(config.Mqtt.Url)
	if config.Mqtt.Username != "" && config.Mqtt.Password != "" {
		opts.SetUsername(config.Mqtt.Username)
		opts.SetPassword(config.Mqtt.Password)
	}
	opts.SetClientID(GetClientId())
	opts.SetCleanSession(true)
	opts.SetBinaryWill(APPNAME+"/status", []byte("Offline"), 0, true)
	opts.SetConnectionLostHandler(connLostHandler)

	mqttClient = MQTT.NewClient(opts)
	token := mqttClient.Connect()
	if token.WaitTimeout(TIMEOUT) && token.Error() != nil {
		log.Fatal(token.Error())
	}

	token = mqttClient.Subscribe(SUBSCRIBE, 0, receive)
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	token = mqttClient.Publish(APPNAME+"/status", 2, true, "Online")
	token.Wait()
}

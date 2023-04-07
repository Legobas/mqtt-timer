package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
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
	Enable      *bool       `json:"enable,omitempty"`
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
		log.Printf("MQTT message error: %s", err.Error())
		return
	}

	if timerInConfig(setTimer) {
		return
	}

	if setTimer.Enable != nil && *setTimer.Enable {
		log.Printf("MQTT message error: programmable timers can only be disabled")
		return
	}

	removed := scheduler.RemoveByTag(setTimer.Id)
	if setTimer.Enable != nil {
		if removed == nil {
			if !*setTimer.Enable {
				log.Printf("Reset '%s'", setTimer.Id)
			}
		} else {
			log.Printf("Warning: timer '%s' not found", setTimer.Id)
		}
		return
	}

	var messages []string

	if setTimer.Topic == "" {
		setTimer.Topic = TIMERS_TOPIC + setTimer.Id + "/event"
	}

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
			log.Printf("Error: incorrect message type: %s", fmt.Sprint(setTimer.Message))
		}
	} else {
		messages = append(messages, setTimer.Id)
	}

	startTime, err := parseStart(setTimer.Start)
	if err != nil {
		log.Println(err)
		return
	}

	offset := parseInterval(setTimer.Interval, messages)

	until, untilTime := parseUntil(setTimer.Until, startTime)

	isEnd := true
	for isEnd {
		for _, message := range messages {
			timer := Timer{}
			timer.Active = true
			timer.Id = setTimer.Id
			timer.Description = strings.TrimPrefix(fmt.Sprintf("%s [%s]", setTimer.Description, message), " ")
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
		if until > 0 {
			until--
			isEnd = until > 0
		} else if until < 0 {
			t1 := startTime.Hour()*60*60 + startTime.Minute()*60 + startTime.Second()
			t2 := untilTime.Hour()*60*60 + untilTime.Minute()*60 + untilTime.Second()
			isEnd = t1 < t2
		} else {
			isEnd = false
		}
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
		return errors.New("id is mandatory")
	}
	if msg.Enable == nil {
		if msg.Start == "" && msg.Interval == "" {
			return errors.New("start or interval must be specified")
		}
		if msg.Until != "" && msg.Interval == "" {
			return errors.New("interval must have a value if until is specified")
		}
	} else {
		if msg.Start != "" || msg.Interval != "" || msg.Until != "" || msg.Topic != "" || msg.Message != nil {
			return errors.New("enable cannot be combined with other fields")
		}
	}

	return nil
}

func timerInConfig(setTimer SetTimer) bool {
	inConfig := false
	id, wildcard := strings.CutSuffix(setTimer.Id, "*")
	// check config
	for _, timer := range config.Timers {
		if timer.Id == id || wildcard && strings.HasPrefix(timer.Id, id) {
			if setTimer.Enable != nil {
				timer.Active = *setTimer.Enable
				if timer.Active {
					log.Printf("Enabled '%s'", timer.Id)
				} else {
					log.Printf("Disabled '%s'", timer.Id)
				}
				inConfig = true
			}
			if !inConfig {
				log.Printf("Error: timer '%s' defined in config", setTimer.Id)
				inConfig = true
			}
		}
	}
	return inConfig
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

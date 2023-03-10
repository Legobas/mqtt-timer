package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
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
	Repeat      string      `json:"repeat"`
	RepeatTimes int         `json:"repeatTimes"`
	RandomAfter string      `json:"randomAfter"`
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
	log.Println("SET: " + msg.Topic() + " = " + message)

	var setTimer SetTimer
	err := json.Unmarshal([]byte(message), &setTimer)
	if err != nil {
		log.Printf("JSON Error: %s", err.Error())
		return
	}

	log.Printf("%+v\n", setTimer)
	log.Printf("type: %s", reflect.TypeOf(setTimer.Message))

	var messages []string

	// check config
	inConfig := false
	for _, timer := range config.Timers {
		if timer.Id == setTimer.Id {
			inConfig = true
			if setTimer.Enabled != nil {
				timer.Enabled = *setTimer.Enabled
			}
			// only enable/disable
			return
		}
	}

	if inConfig {
		log.Printf("Set timer: %s defined in config", setTimer.Id)
		return
	} else {
		scheduler.RemoveByTag(setTimer.Id)
	}

	timer := Timer{}
	timer.Id = setTimer.Id
	timer.Description = setTimer.Description
	timer.Time = setTimer.Start
	timer.RandomAfter = setTimer.RandomAfter
	timer.Topic = setTimer.Topic
	switch setTimer.Message.(type) {
	case string:
		messages = append(messages, setTimer.Message.(string))
	case []interface{}:
		msgArray := setTimer.Message.([]interface{})
		for _, message := range msgArray {
			messages = append(messages, message.(string))
		}
	default:
		log.Fatal(fmt.Sprint(setTimer.Message))
	}

	log.Printf("%+v\n", messages)
	for _, message := range messages {
		log.Println(message)
	}

	matchTime, _ := regexp.Match("^\\d{1,2}(:\\d{2}){1,2}$", []byte(setTimer.Start))
	if matchTime {
		job, err := scheduler.Every(1).Day().At(setTimer.Start).Tag(timer.Id).Do(handleEvent, timer)
		if err != nil {
			log.Printf("Scheduler Error: %s", err.Error())
			return
		}
		job.LimitRunsTo(1)
	} else if strings.ToLower(setTimer.Start) == "now" {
		job, err := scheduler.Every(1).Day().StartImmediately().Tag(timer.Id).Do(handleEvent, timer)
		if err != nil {
			log.Printf("Scheduler Error: %s", err.Error())
			return
		}
		job.LimitRunsTo(1)
	} else if strings.Contains(setTimer.Start, "sec") || strings.Contains(setTimer.Start, "min") {
		seconds := parseSeconds(setTimer.Start)
		if seconds > 0 {
			offset := time.Duration(int64(seconds) * int64(1000000000))
			time := time.Now().Local().Add(offset)
			job, err := scheduler.Every(1).Day().StartAt(time).Tag(timer.Id).Do(handleEvent, timer)
			if err != nil {
				log.Printf("Scheduler Error: %s", err.Error())
				return
			}
			job.LimitRunsTo(1)
		} else {
			log.Printf("Invalid duration: %s", setTimer.Start)
		}
	} else {
		log.Printf("Invalid start time: %s", setTimer.Start)
	}
}

func GetClientId() string {
	hostname, _ := os.Hostname()
	return APPNAME + "_" + hostname
}

func connLostHandler(c MQTT.Client, err error) {
	log.Fatal(err)
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

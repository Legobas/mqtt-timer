package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	TOPIC   string        = "MQTT-Timer"
	TIMEOUT time.Duration = time.Second * 10
)

type SetTimer struct {
	Id          string   `json:"id"`
	Description string   `json:"description"`
	Start       string   `json:"start"`
	Repeat      string   `json:"repeat"`
	RepeatTimes int      `json:"repeatTimes"`
	RandomAfter string   `json:"randomAfter"`
	Topic       string   `json:"topic"`
	Message     string   `json:"message"`
	Messages    []string `json:"messages"`
}

var mqttClient MQTT.Client

func sendToMtt(topic string, message string) {
	if topic == "" {
		topic = TOPIC
	}
	log.Println("MQTT Out: " + topic + " = " + message)
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), config.Mqtt.Retain, message)
}

func sendToMttRetain(topic string, message string) {
	if topic == "" {
		topic = TOPIC
	}
	log.Println(topic + " = " + message)
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), true, message)
}

func receive(client MQTT.Client, msg MQTT.Message) {
	topic := msg.Topic()
	if topic[len(topic)-4:] == "/set" {
		message := string(msg.Payload()[:])
		log.Println("SET: " + topic + " = " + message)

		var setTimer SetTimer
		err := json.Unmarshal([]byte(message), &setTimer)
		if err != nil {
			log.Printf("JSON Error: %s", err.Error())
			return
		}

		log.Printf("%+v\n", setTimer)

		timer := Timer{}
		timer.Id = setTimer.Id
		timer.Description = setTimer.Description
		timer.Time = setTimer.Start
		timer.RandomAfter = setTimer.RandomAfter
		timer.Topic = setTimer.Topic
		timer.Message = setTimer.Message

		matchTime, _ := regexp.Match("^\\d{1,2}(:\\d{2}){1,2}$", []byte(setTimer.Start))
		if matchTime {
			job, err := scheduler.Every(1).Day().At(setTimer.Start).Do(handleEvent, timer)
			if err != nil {
				log.Printf("Scheduler Error: %s", err.Error())
				return
			}
			job.LimitRunsTo(1)
		} else if strings.ToLower(setTimer.Start) == "now" {
			job, err := scheduler.Every(1).Day().StartImmediately().Do(handleEvent, timer)
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
				job, err := scheduler.Every(1).Day().StartAt(time).Do(handleEvent, timer)
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
}

func GetClientId() string {
	hostname, _ := os.Hostname()
	return TOPIC + "_" + hostname
}

func connLostHandler(c MQTT.Client, err error) {
	log.Fatal(err)
}

func startMqttClient() {
	subscribe := TOPIC + "/#"

	opts := MQTT.NewClientOptions().AddBroker(config.Mqtt.Url)
	if config.Mqtt.Username != "" && config.Mqtt.Password != "" {
		opts.SetUsername(config.Mqtt.Username)
		opts.SetPassword(config.Mqtt.Password)
	}
	opts.SetClientID(GetClientId())
	opts.SetCleanSession(true)
	opts.SetBinaryWill(TOPIC+"/status", []byte("Offline"), 0, true)
	opts.SetConnectionLostHandler(connLostHandler)

	mqttClient = MQTT.NewClient(opts)
	token := mqttClient.Connect()
	if token.WaitTimeout(TIMEOUT) && token.Error() != nil {
		log.Fatal(token.Error())
	}

	token = mqttClient.Subscribe(subscribe, 0, receive)
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	token = mqttClient.Publish(TOPIC+"/status", 2, true, "Online")
	token.Wait()
}

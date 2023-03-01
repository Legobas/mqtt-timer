package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	TOPIC   string        = "MQTT-Timer"
	TIMEOUT time.Duration = time.Second * 10
)

type SetTimer struct {
	Cron         string `json:"cron"`
	Time         string `json:"time"`
	Before       string `json:"before"`
	After        string `json:"after"`
	RandomBefore string `json:"randomBefore"`
	RandomAfter  string `json:"randomAfter"`
	Topic        string `json:"topic"`
	Message      string `json:"message"`
}

type Set struct {
	Description string     `json:"description"`
	Timers      []SetTimer `json:"timers"`
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
			log.Println("JSON Error!")
		}

		log.Printf("%+v\n", setTimer)
	}
}

func GetClientId() string {
	hostname, _ := os.Hostname()
	return TOPIC + "_" + hostname
}

func connLostHandler(c MQTT.Client, err error) {
	log.Panic(err)
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

package main

import (
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	TOPIC   string        = "MQTT-Timer"
	TIMEOUT time.Duration = time.Second * 10
)

var mqttClient MQTT.Client

func sendToMtt(topic string, message string) {
	if topic == "" {
		topic = TOPIC
	}
	log.Println("MQTT Out: " + topic + " = " + message)
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), config.Mqtt.Retain, message)
}

func GetClientId() string {
	hostname, _ := os.Hostname()
	return TOPIC + "_" + hostname
}

func connLostHandler(c MQTT.Client, err error) {
	log.Panic(err)
}

func startMqttClient() {
	subscribe := TOPIC + "/set/#"

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

	token = mqttClient.Subscribe(subscribe, 0, func(client MQTT.Client, msg MQTT.Message) {
		topic := msg.Topic()
		message := string(msg.Payload()[:])
		log.Println("MQTT In: " + topic + " = " + message)
	})
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	token = mqttClient.Publish(TOPIC+"/status", 2, true, "Online")
	token.Wait()
}

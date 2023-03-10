﻿# MQTT-Timer

Programmable Timer for MQTT messaging.

[![mqtt-smarthome](https://img.shields.io/badge/mqtt-smarthome-blue.svg?style=flat-square)](https://github.com/mqtt-smarthome/mqtt-smarthome)
[![Build/Test](https://github.com/Legobas/mqtt-timer/actions/workflows/go.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/go.yml)
[![CI/CD](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Legobas/mqtt-timer)](https://goreportcard.com/report/github.com/Legobas/mqtt-timer)
[![Docker Pulls](https://badgen.net/docker/pulls/legobas/mqtt-timer?icon=docker&label=pulls)](https://hub.docker.com/r/legobas/mqtt-timer)
[![Docker Stars](https://badgen.net/docker/stars/legobas/mqtt-timer?icon=docker&label=stars)](https://hub.docker.com/r/legobas/mqtt-timer)
[![Docker Image Size](https://badgen.net/docker/size/legobas/mqtt-timer?icon=docker&label=image%20size)](https://hub.docker.com/r/legobas/mqtt-timer)

A timer is one of the most important parts of a home automation system.
It tells lights, thermostats and other devices to adjust at certain times of the day or at specific days of the week.
But in addition to fixed times, a timer must also know the sunrise and sunset times, as they are different each day. 
This is possible by providing the longitude and latitude coordinates. 

Because it is not always dark at the sunset time a timer needs to wait a certain period of time before or after sunrise/sunset.    
In some situations it is desirable to wait a random number of seconds or minutes before sending the event.
This can for example give the impression that someone is home switching the lights on every day at a different time.

Apart from these configurable timers, timers need to be programmable.
An example is the activation of a dimmable light if motion is detected, where the light percentages will go down in the minutes after the detected movement.

It must be possible to disable, re-enable or cancel a timer.    
MQTT-Timer aims to meet these requirements.

In a MQTT based home automation environment a timer independent from home control software like node-red or Home Assistant increases the stability of the whole system.
It follows the Unix/Linux philosophy: do one thing, and do it well.
If for example node-red crashes the timers will continue to send messages at the specified times.

## Installation

```bash
$ go get -u github.com/Legobas/mqtt-timer
```

# Configuration

MQTT-Timer can be configured with the `mqtt-timer.yml` yaml configuration file.
The `mqtt-timer.yml` file has to exist in one of the following locations:

 * A config directory in de filesystem root: `/config/mqtt-timer.yml`
 * A .config directory in the user home directory `~/.config/mqtt-timer.yml`
 * The current working directory

## Configuration options

| Config item               | Description                                                                  |
| ------------------------- | ---------------------------------------------------------------------------- |
| latitude/longitude        | GPS location used for Sunrise/Sunset                                         |
| **MQTT**                  |                                                                              |
| URL                       | MQTT Server URL                                                              |
| Username/Password         | MQTT Server Credentials                                                      |
| QOS                       | MQTT Server Quality Of Service                                               |
| Retain                    | MQTT Server Retain messages                                                  |
| **Timers**                |                                                                              |
| id                        | Unique ID for this message (mandatory)                                       |
| time                      | Time in `15:04` or `15:04:05` format                                         |
| cron                      | Cron expression in '`30 7 * * *`' or '`15 30 7 * * *`' (with seconds) format |
| description               | something useful                                                             |
| topic                     | MQTT Topic                                                                   |
| message                   | raw string or JSON                                                           |
| before, after             | offset: fixed number of seconds or minutes                                   |
| randomBefore, randomAfter | offset: random number of seconds or minutes                                  |

Example mqtt-timer.yml:

```yml
    latitude: 51.50722
    longitude: -0.1275
    
    mqtt:
      url: "tcp://<MQTT SERVER>:1883"
      username: <MQTT USERNAME>
      password: <MQTT PASSWORD>
      qos: 0
      retain: false
      
    timers:
    - id: 001
      time: 22:30
      description: Light outside on at 22:30
      topic: shellies/Shelly1/relay/0/command
      message: on
    - id: 002
      time: sunrise
      before: 20 minutes
      description: Light outside off 20 minutes before sunrise
      topic: shellies/Shelly1/relay/0/command
      message: off
```

See also: [Example mqtt-timer.yml](https://github.com/Legobas/mqtt-timer/blob/main/mqtt-timer.yml)

## Programmable timers

Timers can be set by sending a MQTT JSON messages to the topic:

    MQTT-Timer/set

The JSON message can use the following fields to set a timer:
 

| Field       | Description                                         |
| ----------- | --------------------------------------------------- |
| id          | unique ID for this message (mandatory)              |
| description | something useful                                    |
| start       | '`now`'                                             |
|             | duration in '`25 sec`' or '`12 min`' format         |
|             | time in '`15:04`' or '`15:04:05`' format            |
|             | if start is omitted the timer will fire immediately |
| repeat      | duration in '`25 sec`' or '`12 min`' format         |
| repeatUntil | number of times in digit(s)                         |
|             | duration in '`25 sec`' or '`12 min`' format         |
|             | time in '`15:04`' or '`15:04:05`' format            |
| topic       | MQTT Topic                                          |
| message     | raw string or JSON                                  |
| randomAfter | offset: random number of seconds or minutes         |


The JSON message to disable or cancel a timer:

| Field   | Description                                               |
| ------- | --------------------------------------------------------- |
| id      | unique ID for this message (mandatory)                    |
| enabled | true or false                                             |
|         | true (re-enable) can only be used for configurable timers |


examples:

```json
{
  "id": "light001",
  "description": "Light on after random max 10 min.",
  "start": "now,10 min, 10:15:00",
  "randomAfter": "10 min",
  "topic": "/homeassistant/light01",
  "message": "on",
}

{
  "id": "msg001",
  "description": "Dim light from now every 10 min.",
  "start": "now,10 min, 10:15:00",
  "repeat": "1 min",
  "repeatTimes": 10,
  "randomAfter": "1 min",
  "topic": "/homeassistant/light04/dimmer",
  "message": ["100%", "80%", "60%", "20%", "0%"]
}

{
  "id": "alarm01",
  "description": "Intruder detected",
  "start": "now",
  "repeat": "1 sec",
  "repeatTimes": 100,
  "topic": "/homeassistant/light02",
  "message": ["on", "off"]
}
```

## Docker

Docker run example:

```bash
$ docker run -d -v /home/legobas/mqtt-timer:/config legobas/mqtt-timer
```

Docker compose example:

```yml
version: "3.9"

services:
  MqttTimer:
    image: legobas/mqtt-timer
    container_name: mqtt-timer
    environment:
      - TZ=America/New_York
    volumes:
      - /home/legobas/mqtt-timer:/config
    restart: unless-stopped
```

## Timezone

By default all the times will be in the timezone of the server.
In a docker environment the timezone can be specified by the TZ environment variable.

For example: 

```bash
$ docker run -e TZ=America/New_York mqtt-timer
```

## Credits

Libraries:
* [GoCron](https://github.com/go-co-op/gocron)
* [Paho Mqtt Client](https://github.com/eclipse/paho.mqtt.golang)
* [GoSunrise](https://github.com/nathan-osman/go-sunrise)

# MQTT-Timer

Programmable Timer for MQTT messaging.

[![Build/Test](https://github.com/Legobas/mqtt-timer/actions/workflows/go.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/go.yml)
[![CI/CD](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Legobas/mqtt-timer)](https://goreportcard.com/report/github.com/Legobas/mqtt-timer)
[![Docker Pulls](https://badgen.net/docker/pulls/legobas/mqtt-timer?icon=docker&label=pulls)](https://hub.docker.com/r/legobas/mqtt-timer)
[![Docker Stars](https://badgen.net/docker/stars/legobas/mqtt-timer?icon=docker&label=stars)](https://hub.docker.com/r/legobas/mqtt-timer)
[![Docker Image Size](https://badgen.net/docker/size/legobas/mqtt-timer?icon=docker&label=image%20size)](https://hub.docker.com/r/legobas/mqtt-timer)

## Features

 * Standard timers are configurable 
 * Timers can fire daily or at specific days of the week
 * Timers can use daily sunrise & sunset times based on the latitude/longitude coordinates
 * Timers can have a fixed offset before or after a timestamp
 * Timers can have a random offset before or after a timestamp
 * Timers can be removed/reset
 * Programmable timers can be specified by MQTT JSON Messages
 * Programmable timers are repeatable
 * Programmable timers can wait for a random number of seconds/minutes before being fired
 * A range of programmable timers can be set with one MQTT message

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

| Config item               | Description                                                       |
| ------------------------- | ----------------------------------------------------------------- |
| latitude/longitude        | GPS location used for Sunrise/Sunset                              |
| **MQTT**                  |                                                                   |
| URL                       | MQTT Server URL                                                   |
| Username/Password         | MQTT Server Credentials                                           |
| QOS                       | MQTT Server Quality Of Service                                    |
| Retain                    | MQTT Server Retain messages                                       |
| **Timers**                |                                                                   |
| id                        | Unique ID for this message (mandatory)                            |
| time                      | Time in `15:04` or `15:04:05` format                              |
| cron                      | Cron in '`30 7 * * *`' or '`15 30 7 * * *`' (with seconds) format |
| topic                     | MQTT Topic                                                        |
| message                   | simple string or JSON                                             |
| before, after             | offset: fixed number of seconds or minutes                        |
| randomBefore, randomAfter | offset: random number of seconds or minutes                       |

Example config.yml:

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

See also: [example/config.yml](example/config.yml)

## MQTT JSON messages

    timer2mqtt/dimmer/set

```json
{
  "id": "msg001",
  "description": "Dim light",
  "start": "now",
  "randomAfter": "1 min",
  "topic": "/homeassistant/light04/dimmer",
  "message": "10%",
}
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

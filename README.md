# MQTT-Timer

Programmable Timer for MQTT messaging.

[![mqtt-smarthome](https://img.shields.io/badge/mqtt-smarthome-blue.svg?style=flat-square)](https://github.com/mqtt-smarthome/mqtt-smarthome)
[![Build/Test](https://github.com/Legobas/mqtt-timer/actions/workflows/release.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/release.yml)
[![CI/CD](https://github.com/Legobas/mqtt-timer/actions/workflows/deploy.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/deploy.yml)
[![CodeQL](https://github.com/Legobas/mqtt-timer/actions/workflows/codeql.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Legobas/mqtt-timer)](https://goreportcard.com/report/github.com/legobas/mqtt-timer)
[![Docker Pulls](https://badgen.net/docker/pulls/legobas/mqtt-timer?icon=docker&label=pulls)](https://hub.docker.com/r/legobas/mqtt-timer)
[![Docker Image Size](https://badgen.net/docker/size/legobas/mqtt-timer?icon=docker&label=image%20size)](https://hub.docker.com/r/legobas/mqtt-timer)

MQTT-Timer is a flexible timer service for scheduling and automating tasks.
It allows you to set up timers for any task, from turning on lights to running a script.
Because it is based on the MQTT protocol can it be easily be integrated with other home automation systems. 
It also provides a range of features such as customizable time intervals, random timers, sunrise/sunset timers and logging.

In a MQTT-based home automation environment, a timer independent from home control software like Node-Red or Home Assistant can significantly improve the stability of the system.
Adhering to the Unix/Linux philosophy of "do one thing, and do it well," this timer will continue to send messages at the specified times, 
even if node-red or other home control software crashes. 
This ensures that the system remains reliable and consistent, even in the event of an unexpected interruption.

## Installation

MQTT-Timer can be used in a [Go](https://go.dev) environment or as a [Docker container](#docker):

```bash
$ go get -u github.com/Legobas/mqtt-timer
```

## Environment variables

Supported environment variables:

```
LOGLEVEL = INFO/DEBUG/ERROR
```

# Configuration

MQTT-Timer can be configured with the `mqtt-timer.yml` yaml configuration file.
The `mqtt-timer.yml` file has to exist in one of the following locations:

 * A `config` directory in de filesystem root: `/config/mqtt-timer.yml`
 * A `.config` directory in the user home directory `~/.config/mqtt-timer.yml`
 * The current working directory

## Configuration options

| Config item               | Description                                                              |
| ------------------------- | ------------------------------------------------------------------------ |
| latitude/longitude        | GPS location used for Sunrise/Sunset                                     |
| **mqtt**                  |                                                                          |
| url                       | MQTT Server URL                                                          |
| username/password         | MQTT Server Credentials                                                  |
| qos                       | MQTT Server Quality Of Service                                           |
| retain                    | MQTT Server Retain messages                                              |
| **timers**                |                                                                          |
| id                        | Unique ID for this timer (mandatory)                                     |
| time                      | Time in `15:04` or `15:04:05` format                                     |
|                           | `sunrise` or `sunset`                                                    |
| cron                      | Cron expression in `30 7 * * *` or `15 30 7 * * *` (with seconds) format |
| description               | something useful                                                         |
| topic                     | MQTT Topic                                                               |
| message                   | string -->  message: `on`                                                |
|                           | JSON --> message: `'{"device"="light1", "command"="on"}'`                |
| before, after             | offset: fixed duration in `25 sec`,`12 min` or `1 hour` format           |
| randomBefore, randomAfter | offset: random duration in `25 sec`,`12 min` or `1 hour` format          |
| enabled                   | true (default), false                                                    |

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

Timers can be set by sending a MQTT JSON message to the topic:

    MQTT-Timer/set

The following fields can be part of the JSON message:
 
| Field       | Description                                                 | Default                        |
| ----------- | ----------------------------------------------------------- | ------------------------------ |
| id          | unique ID for this message (mandatory)                      |                                |
| description | something useful                                            |                                |
| start       | after: duration in `25 sec`,`12 min` or `1 hour` format     | immediately                    |
|             | at: time in `15:04` or `15:04:05` format                    |                                |
| interval    | duration in `25 sec`,`12 min` or `1 hour` format            | 30 seconds                     |
| until       | number of times in `10 times` or `10` format                | 1 time                         |
|             | duration in `25 sec`,`12 min` or `1 hour` format            |                                |
|             | time in `15:04` or `15:04:05` format                        |                                |
| topic       | MQTT Topic                                                  | `MQTT-Timer/timers/<id>/event` |
| message     | MQTT Message -->  "message": `"on"`                         | id                             |
|             | JSON --> "message": `"{'device'='light1', 'command'='on'}"` |                                |
|             | JSON Array --> "message": `["green", "red", "blue"]`        |                                |

examples:

```json
{
  "id": "alarm01",
  "description": "Intruder detected",
  "interval": "1 sec",
  "until": "100 times",
  "topic": "/homeassistant/light02",
  "message": ["on", "off"]
}
```

```json
{
  "id": "light01",
  "description": "Light on after 10 min.",
  "start": "10 min",
  "topic": "/homeassistant/light01",
  "message": "on",
}
```

```json
{
  "id": "pulsating_dimmer",
  "description": "Dim light from 100% to 0% and back to 100% with 10 second steps from 10:15 to 10:20",
  "start": "10:15:00",
  "interval": "10 sec",
  "until": "10:20:00",
  "topic": "/homeassistant/light04/dimmer",
  "message": ["100%", "80%", "60%", "20%", "0%", "20%", "60%", "80%", "100%"]
}
```

### Disable/Enable timers

Timers can be disabled or enabled by sending a JSON message with the `enable` field.

The MQTT topic for the disable/enable message is the same:

    MQTT-Timer/set

The behavior if a message with `enable: false` is received:
* Configurable timers will be paused.
* Programmable timers will be removed from the scheduler.

The behavior if a message with `enable: true` is received:
* Configurable timers will be activated.
* Programmable timers won't change, an error message will be logged.

Besides the `id` field the `enable` field has to be the only field in the message.

The JSON message to disable or cancel a timer:

| Field  | Description                                                              |
| ------ | ------------------------------------------------------------------------ |
| id     | unique ID for this message (mandatory)                                   |
|        | wildcard: `lamp_*` will enable/disable every timer starting with "lamp_" |
| enable | true or false                                                            |
|        | true (re-enable) can only be used for configurable timers                |

examples:

```json
{
  "id": "light01",
  "enable": false
}
```

```json
{
  "id": "light*",
  "enable": true
}
```

## Docker

Docker run example:

```bash
$ docker run -d -v /home/legobas/mqtt-timer:/config legobas/mqtt-timer
```

Docker compose example:

```yml
version: "3.0"

services:
  MqttTimer:
    image: legobas/mqtt-timer:latest
    container_name: mqtt-timer
    environment:
      - LOGLEVEL=debug
      - TZ=America/New_York
    volumes:
      - /home/legobas/mqtt-timer:/config:ro
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

* [GoCron](https://github.com/go-co-op/gocron)
* [Paho Mqtt Client](https://github.com/eclipse/paho.mqtt.golang)
* [GoSunrise](https://github.com/nathan-osman/go-sunrise)
* [ZeroLog](https://github.com/rs/zerolog)

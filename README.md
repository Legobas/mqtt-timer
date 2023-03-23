# MQTT-Timer

Programmable Timer for MQTT messaging.

[![mqtt-smarthome](https://img.shields.io/badge/mqtt-smarthome-blue.svg?style=flat-square)](https://github.com/mqtt-smarthome/mqtt-smarthome)
[![Build/Test](https://github.com/Legobas/mqtt-timer/actions/workflows/go.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/go.yml)
[![CI/CD](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Legobas/mqtt-timer)](https://goreportcard.com/report/github.com/Legobas/mqtt-timer)
[![Docker Pulls](https://badgen.net/docker/pulls/legobas/mqtt-timer?icon=docker&label=pulls)](https://hub.docker.com/r/legobas/mqtt-timer)
[![Docker Stars](https://badgen.net/docker/stars/legobas/mqtt-timer?icon=docker&label=stars)](https://hub.docker.com/r/legobas/mqtt-timer)
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

```bash
$ go get -u github.com/Legobas/mqtt-timer
```

# Configuration

MQTT-Timer can be configured with the `mqtt-timer.yml` yaml configuration file.
The `mqtt-timer.yml` file has to exist in one of the following locations:

 * A `config` directory in de filesystem root: `/config/mqtt-timer.yml`
 * A `.config` directory in the user home directory `~/.config/mqtt-timer.yml`
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
| message                   | string -->  message: on                                                      |
|                           | JSON --> message: '{"device"="light1", "command"="on"}'                      |
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
 

| Field       | Description                                               | Default                      |
| ----------- | --------------------------------------------------------- | ---------------------------- |
| id          | unique ID for this message (mandatory)                    |                              |
| description | something useful                                          |                              |
| start       | after: duration in '`25 sec`' or '`12 min`' format        | immediately                  |
|             | at: time in '`15:04`' or '`15:04:05`' format              |                              |
| interval    | duration in '`25 sec`' or '`12 min`' format               | 30 seconds                   |
| until       | number of times in '`10 times`' or '`10`' format          | 1 time                       |
|             | duration in '`25 sec`' or '`12 min`' format               |                              |
|             | time in '`15:04`' or '`15:04:05`' format                  |                              |
| topic       | MQTT Topic                                                | MQTT-Timer/timers/<id>/event |
|             | JSON Array --> "topic": ["device1/cmd", "device2/cmd"]    |                              |
| message     | MQTT Message -->  "message": "on"                         | id                           |
|             | JSON --> "message": "{'device'='light1', 'command'='on'}" |                              |
|             | JSON Array --> "message": ["green", "red", "blue"]        |                              |


### Disable/Enable timer

Timers can be disabled by sending a message with the enabled field set to false.

Behavior if a message with enabled=false is received:
* Configurable timers will be paused.
* Programmable timer will be removed from the scheduler.

If the enabled field is set no other fields will be applied.

The JSON message to disable or cancel a timer:

| Field   | Description                                               |
| ------- | --------------------------------------------------------- |
| id      | unique ID for this message (mandatory)                    |
| enabled | true or false                                             |
|         | true (re-enable) can only be used for configurable timers |

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
  "description": "Light on after random max 10 min.",
  "start": "10 min",
  "topic": "/homeassistant/light01",
  "message": "on",
}
```

```json
{
  "id": "dim01",
  "description": "Dim light from now every minute",
  "start": "10:15:00",
  "interval": "1 min",
  "until": "10:20:00",
  "topic": "/homeassistant/light04/dimmer",
  "message": ["100%", "80%", "60%", "20%", "0%"]
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

* [GoCron](https://github.com/go-co-op/gocron)
* [Paho Mqtt Client](https://github.com/eclipse/paho.mqtt.golang)
* [GoSunrise](https://github.com/nathan-osman/go-sunrise)

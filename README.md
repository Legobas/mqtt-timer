# mqtt-timer

Programmable Timer for MQTT

[![Build/Test](https://github.com/Legobas/mqtt-timer/actions/workflows/go.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/go.yml)
[![CI/CD](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Legobas/mqtt-timer)](https://goreportcard.com/report/github.com/Legobas/mqtt-timer)
[![Docker Pulls](https://badgen.net/docker/pulls/legobas/mqtt-timer?icon=docker&label=pulls)](https://hub.docker.com/r/legobas/mqtt-timer)
[![Docker Stars](https://badgen.net/docker/stars/legobas/mqtt-timer?icon=docker&label=stars)](https://hub.docker.com/r/legobas/mqtt-timer)
[![Docker Image Size](https://badgen.net/docker/size/legobas/mqtt-timer?icon=docker&label=image%20size)](https://hub.docker.com/r/legobas/mqtt-timer)


MQTT-Timer is a programmable timer for MQTT messaging.
Common daily or weekly timers can be specified in the configuration.
Any timers can be added by MQTT JSON messages.

## Installation

```bash
$ go get -u github.com/Legobas/mqtt-timer
```

## Credits

Libraries:
* [GoCron](https://github.com/go-co-op/gocron)
* [Paho Mqtt Client](https://github.com/eclipse/paho.mqtt.golang)
* [GoSunrise](https://github.com/nathan-osman/go-sunrise)

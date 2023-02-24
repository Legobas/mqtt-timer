# mqtt-timer
Programmable Timer for MQTT

[![mqtt-smarthome](https://img.shields.io/badge/mqtt-smarthome-blue.svg)](https://github.com/mqtt-smarthome/mqtt-smarthome)
[![Build](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml/badge.svg)](https://github.com/Legobas/mqtt-timer/actions/workflows/build.yml)


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

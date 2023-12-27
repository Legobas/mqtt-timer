package main

import (
	_ "embed"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/nathan-osman/go-sunrise"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	APPNAME      string = "MQTT-Timer"
	TIMERS_TOPIC string = APPNAME + "/timers/"
)

var (
	//go:embed version.txt
	VERSION     string
	config      Config
	dailyTimers []*Timer
	scheduler   *gocron.Scheduler
)

func init() {
	// Setup logging
	out := zerolog.NewConsoleWriter()
	out.NoColor = true
	out.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%-6s", i))
	}
	out.PartsExclude = []string{zerolog.TimestampFieldName, zerolog.CallerFieldName}
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(out)

	switch strings.ToLower(os.Getenv("LOGLEVEL")) {
	case "debug":
		log.Logger = log.Level(zerolog.DebugLevel)
	case "trace":
		log.Logger = log.Level(zerolog.TraceLevel)
	default:
		log.Logger = log.Level(zerolog.InfoLevel)
	}

	// Get Config
	config = getConfig()

	// Print Version
	log.Info().Msgf("%s %s", APPNAME, VERSION)
}

func handleEvent(timer *Timer) {
	if timer.Active {
		if timer.RandomBefore != "" || timer.After != "" || timer.RandomAfter != "" {
			time.Sleep(offsetDuration(timer))
		}
		descr := ""
		if timer.Description != "" {
			descr = " - " + timer.Description
		}
		log.Debug().Msgf("[%s] %s %s%s%s", timer.Id, offsetDescr(timer), timer.Time, timer.Cron, descr)

		timerTopic := TIMERS_TOPIC + timer.Id
		msg := time.Now().Format("2006-01-02 15:04:05")
		sendToMttRetain(timerTopic+"/event", msg)

		if timer.Topic != "" || timer.Message != "" {
			timerTopic = timerTopic + "/message"
			if timer.Topic != "" {
				timerTopic = timer.Topic
			}
			if timer.Message != "" {
				msg = timer.Message
			}
			sendToMtt(timerTopic, msg)
		}
	}
}

func setTimers() {
	for i := 0; i < len(config.Timers); i++ {
		timer := &config.Timers[i]
		disabled := ""
		if !timer.Active {
			disabled = " (disabled)"
		}
		if timer.Cron != "" {
			// Cron
			if len(strings.Split(timer.Cron, " ")) == 5 {
				log.Info().Msgf("Scheduled '%s'%s Cron [%s] '%s'", timer.Id, disabled, timer.Cron, timer.Description)
				scheduler.Cron(timer.Cron).Tag(timer.Id).Do(handleEvent, timer)
			} else if len(strings.Split(timer.Cron, " ")) == 6 {
				// Cron with Seconds
				log.Info().Msgf("Scheduled '%s'%s Cron [%s] '%s'", timer.Id, disabled, timer.Cron, timer.Description)
				scheduler.CronWithSeconds(timer.Cron).Tag(timer.Id).Do(handleEvent, timer)
			} else {
				log.Error().Msgf("Invalid Cron format: [%s]", timer.Cron)
			}
		} else if timer.Time != "" {
			// Time
			days := "daily"
			if timer.Days != "" {
				days = timer.Days
			}

			match, _ := regexp.Match("^\\d{1,2}(:\\d{2}){1,2}$", []byte(timer.Time))
			if match {
				schedule := scheduler.Every(1).Day()
				if timer.Days != "" {
					schedule = scheduler.Every(1).Week()
					if strings.Contains(timer.Days, "mon") {
						schedule = schedule.Monday()
					}
					if strings.Contains(timer.Days, "tue") {
						schedule = schedule.Tuesday()
					}
					if strings.Contains(timer.Days, "wed") {
						schedule = schedule.Wednesday()
					}
					if strings.Contains(timer.Days, "thu") {
						schedule = schedule.Thursday()
					}
					if strings.Contains(timer.Days, "fri") {
						schedule = schedule.Friday()
					}
					if strings.Contains(timer.Days, "sat") {
						schedule = schedule.Saturday()
					}
					if strings.Contains(timer.Days, "sun") {
						schedule = schedule.Sunday()
					}
				}
				schedTime := timeBefore(timer, timer.Time)
				schedule.At(schedTime).Tag(timer.Id).Do(handleEvent, timer)

				log.Info().Msgf("Scheduled '%s'%s %s %s %s '%s'", timer.Id, disabled, days, offsetDescr(timer), timer.Time, timer.Description)
			} else if timer.Time == "sunrise" || timer.Time == "sunset" {
				dailyTimers = append(dailyTimers, timer)
				log.Info().Msgf("Scheduled '%s'%s %s %s %s '%s'", timer.Id, disabled, days, offsetDescr(timer), timer.Time, timer.Description)
			} else {
				log.Error().Msgf("Invalid config [%v]", timer)
			}
		} else {
			log.Error().Msgf("Invalid config [%v]", timer)
		}
	}
}

func offsetDescr(timer *Timer) string {
	descr := "at"
	if timer.Before != "" {
		descr = timer.Before + " before"
	} else if timer.RandomBefore != "" {
		descr = "random max " + timer.RandomBefore + " before"
	} else if timer.After != "" {
		descr = timer.After + " after"
	} else if timer.RandomAfter != "" {
		descr = "random max " + timer.RandomAfter + " after"
	}
	return descr
}

func offsetDuration(timer *Timer) time.Duration {
	var offset int64

	offsetStr := ""
	random := false
	if timer.Before != "" {
		// nop
	} else if timer.RandomBefore != "" {
		offsetStr = timer.RandomBefore
		random = true
	} else if timer.After != "" {
		offsetStr = timer.After
	} else if timer.RandomAfter != "" {
		offsetStr = timer.RandomAfter
		random = true
	}

	seconds := parseDuration(offsetStr)
	if random {
		offset = int64(rand.Intn(seconds)) * int64(1000000000)
	} else {
		offset = int64(seconds) * int64(1000000000)
	}

	return time.Duration(offset)
}

func timeBefore(timer *Timer, timeStr string) time.Time {
	offsetStr := ""
	if timer.Before != "" {
		offsetStr = timer.Before
	} else if timer.RandomBefore != "" {
		offsetStr = timer.RandomBefore
	}

	offset := parseDuration(offsetStr)

	offsetTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		offsetTime, err = time.Parse("15:04:05", timeStr)
		if err != nil {
			log.Error().Msgf("Error: invalid time format: %s", timeStr)
		}
	}
	offsetTime = offsetTime.Add(time.Duration(-1*offset) * time.Second)

	return offsetTime
}

func setDailyTimes(midnight bool) {
	if midnight {
		timerTopic := TIMERS_TOPIC + "midnight"
		msg := time.Now().Format("2006-01-02 15:04:05")
		sendToMtt(timerTopic+"/event", msg)
	}

	sunrise, sunset := sunrise.SunriseSunset(config.Latitude, config.Longitude,
		time.Now().Year(), time.Now().Month(), time.Now().Day())

	// Sunrise
	sunriseTime := sunrise.Local()
	sunriseStr := sunriseTime.Format("15:04")
	if sunriseTime.After(time.Now().Local()) {
		timer := Timer{}
		timer.Id = "sunrise"
		timer.Time = sunriseStr
		timer.Active = true
		job, _ := scheduler.Every(1).Day().At(sunriseTime).Do(handleEvent, &timer)
		job.LimitRunsTo(1)
		log.Info().Msgf("Today: 'Sunrise' at %s", sunriseStr)
	}

	// Sunset
	sunsetTime := sunset.Local()
	sunsetStr := sunsetTime.Format("15:04")
	if sunsetTime.After(time.Now().Local()) {
		timer := Timer{}
		timer.Id = "sunset"
		timer.Time = sunsetStr
		timer.Active = true
		job, _ := scheduler.Every(1).Day().At(sunsetTime).Do(handleEvent, &timer)
		job.LimitRunsTo(1)
		log.Info().Msgf("Today: 'Sunset' at %s", sunsetStr)
	}

	// Daily timers
	for i := 0; i < len(dailyTimers); i++ {
		timer := dailyTimers[i]
		day := strings.ToLower(time.Now().Local().Weekday().String()[:3])
		if timer.Days == "" || strings.Contains(timer.Days, day) {
			timeStr := ""
			if timer.Time == "sunrise" {
				if time.Now().Local().After(sunriseTime) {
					continue
				}
				timeStr = sunriseStr
			} else if timer.Time == "sunset" {
				if time.Now().Local().After(sunsetTime) {
					continue
				}
				timeStr = sunsetStr
			}
			time := timeBefore(timer, timeStr)
			job, _ := scheduler.Every(1).Day().At(time).Tag(timer.Id).Do(handleEvent, timer)
			job.LimitRunsTo(1)
			log.Info().Msgf("Today: '%s' %s %s '%s'", timer.Id, offsetDescr(timer), timer.Time, timer.Description)
		}
	}

	// Refresh status
	sendToMttRetain(APPNAME+"/status", "Online")
}

func main() {
	zoneName, _ := time.Now().Zone()
	log.Debug().Msgf("%s start, Local Time=%s Timezone=%s", APPNAME, time.Now().Local().Format("15:04:05"), zoneName)

	scheduler = gocron.NewScheduler(time.Now().Location())

	startMqttClient()

	if config.Latitude != 0 && config.Longitude != 0 {
		scheduler.Every(1).Day().At("00:00").Do(setDailyTimes, true)

		// Startup: set timers for today once
		job, _ := scheduler.Every(1).Day().StartImmediately().Do(setDailyTimes, false)
		job.LimitRunsTo(1)
	} else {
		log.Warn().Msg("Warning: Latitude and Longitude not set, sunrise/sunset cannot be used")
	}

	setTimers()
	scheduler.StartAsync()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	log.Debug().Msgf("%s stop, Local Time=%s Timezone=%s", APPNAME, time.Now().Local().Format("15:04:05"), zoneName)
}

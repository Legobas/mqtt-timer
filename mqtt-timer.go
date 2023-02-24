package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/nathan-osman/go-sunrise"
)

var config = getConfig()
var sunriseTimer Timer
var sunsetTimer Timer

func init() {
	rand.Seed(time.Now().UnixNano())
}

func handleEvent(timer Timer) {
	log.Printf(timer.Id + " - " + timer.Description)
	msg := ""

	if timer.RandomBefore != "" || timer.After != "" || timer.RandomAfter != "" {
		offset := "after"
		if timer.RandomBefore != "" || timer.RandomAfter != "" {
			if timer.RandomBefore != "" {
				offset = "random (max " + timer.RandomBefore + ") before"
			} else {
				offset = "random (max " + timer.RandomAfter + ") after"
			}
		}
		msg += fmt.Sprintf("Timer: %s %s ", offset, timer.Time)
		time.Sleep(offsetDuration(timer))
		log.Printf("Timer Event Exec: " + timer.Description)
	}

	topic := TOPIC + "/" + timer.Id + "/time"
	msg = time.Now().Format("2006-01-02 15:04:05")
	sendToMtt(topic, msg)

	if timer.Topic != "" || timer.Message != "" {
		topic = TOPIC + "/" + timer.Id + "/message"
		if timer.Topic != "" {
			topic = timer.Topic
		}
		if timer.Message != "" {
			msg = timer.Message
		}
		sendToMtt(topic, msg)
	}
}

func setTimers(s *gocron.Scheduler) {
	for _, timer := range config.Timers {
		if timer.Cron != "" {
			// Cron
			if len(strings.Split(timer.Cron, " ")) == 5 {
				log.Printf("Scheduled '%s' Cron [%s] '%s'", timer.Id, timer.Cron, timer.Description)
				s.Cron(timer.Cron).Do(handleEvent, timer)
			} else if len(strings.Split(timer.Cron, " ")) == 6 {
				// Cron with Seconds
				log.Printf("Scheduled '%s' Cron [%s] '%s'", timer.Id, timer.Cron, timer.Description)
				s.CronWithSeconds(timer.Cron).Do(handleEvent, timer)
			} else {
				log.Printf("Invalid Cron format: [%s]", timer.Cron)
			}
		} else if timer.Time != "" {
			// Time
			days := "daily"
			if timer.Days != "" {
				days = timer.Days
			}
			offset := "at"
			if timer.Before != "" {
				offset = timer.Before + " before"
			} else if timer.RandomBefore != "" {
				offset = "random max " + timer.RandomBefore + " before"
			} else if timer.After != "" {
				offset = timer.After + " after"
			} else if timer.RandomAfter != "" {
				offset = "random max " + timer.RandomAfter + " after"
			}

			match, _ := regexp.Match("^\\d{1,2}(:\\d{2}){1,2}$", []byte(timer.Time))
			if match {
				schedule := s.Every(1).Day()
				if timer.Days != "" {
					schedule = s.Every(1).Week()
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
				schedTime := timeWithOffset(timer)
				schedule.At(schedTime).Do(handleEvent, timer)

				log.Printf("Scheduled '%s' %s %s %s '%s'", timer.Id, days, offset, timer.Time, timer.Description)
			} else if timer.Time == "sunrise" {
				sunriseTimer = timer
				log.Printf("Scheduled '%s' %s %s sunrise '%s'", timer.Id, days, offset, timer.Description)
			} else if timer.Time == "sunset" {
				sunsetTimer = timer
				log.Printf("Scheduled '%s' %s %s sunset: '%s'", timer.Id, days, offset, timer.Description)
			} else {
				log.Printf("Invalid config [%v]", timer)
			}
		} else {
			log.Printf("Invalid config [%v]", timer)
		}
	}
}

func offsetDuration(timer Timer) time.Duration {
	offset := int64(0)

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

	times := strings.Split(offsetStr, " ")
	if len(times) == 2 && times[1][:3] == "sec" {
		seconds, _ := strconv.Atoi(times[0])
		if random {
			offset = int64(rand.Intn(seconds) * 1000000000)
		} else {
			offset = int64(seconds * 1000000000)
		}
	} else if len(times) == 2 && times[1][:3] == "min" {
		minutes, _ := strconv.Atoi(times[0])
		if random {
			offset = int64(rand.Intn(minutes) * 60000000000)
		} else {
			offset = int64(minutes * 60000000000)
		}
	}
	return time.Duration(offset)
}

func timeWithOffset(timer Timer) time.Time {
	offset := 0

	offsetStr := ""
	if timer.Before != "" {
		offsetStr = timer.Before
	} else if timer.RandomBefore != "" {
		offsetStr = timer.RandomBefore
	}

	times := strings.Split(offsetStr, " ")
	if len(times) == 2 && times[1][:3] == "sec" {
		offset, _ = strconv.Atoi(times[0])
	} else if len(times) == 2 && times[1][:3] == "min" {
		minutes, _ := strconv.Atoi(times[0])
		offset = minutes * 60
	}

	offsetTime, err := time.Parse("15:04", timer.Time)
	if err != nil {
		offsetTime, err = time.Parse("15:04:05", timer.Time)
		if err != nil {
			log.Printf("Error: invalid time format: %s", timer.Time)
		}
	}
	offsetTime = offsetTime.Add(time.Duration(-1*offset) * time.Second)

	return offsetTime
}

func setDailyTimes(s *gocron.Scheduler) {
	sunrise, sunset := sunrise.SunriseSunset(config.Latitude, config.Longitude,
		time.Now().Year(), time.Now().Month(), time.Now().Day())

	// Sunrise
	sunriseTime := sunrise.Local()
	if sunriseTime.After(time.Now().Local()) {
		timer := Timer{}
		timer.Id = "sunrise"
		timer.Time = "sunrise"
		job, _ := s.Every(1).Day().At(sunriseTime).Do(handleEvent, timer)
		job.LimitRunsTo(1)
		log.Printf("Set Sunrise %s", sunriseTime.Format("15:04:05"))
		if sunriseTimer.Time != "" {
			sunriseTime = sunriseTime.Add(offsetDuration(sunriseTimer))
			job, _ := s.Every(1).Day().At(sunriseTime).Do(handleEvent, sunriseTimer)
			job.LimitRunsTo(1)
			log.Printf("Set Scheduled Sunrise %s", sunriseTime.Format("15:04:05"))
		}
	}

	// Sunset
	sunsetTime := sunset.Local()
	if sunsetTime.After(time.Now().Local()) {
		timer := Timer{}
		timer.Id = "sunset"
		timer.Time = "sunset"
		job, _ := s.Every(1).Day().At(sunsetTime).Do(handleEvent, timer)
		job.LimitRunsTo(1)
		log.Printf("Set Sunset %s", sunsetTime.Format("15:04:05"))
		if sunsetTimer.Time != "" {
			sunsetTime = sunsetTime.Add(offsetDuration(sunriseTimer))
			job, _ := s.Every(1).Day().At(sunsetTime).Do(handleEvent, sunsetTimer)
			job.LimitRunsTo(1)
			log.Printf("Set Scheduled Sunset %s", sunsetTime.Format("15:04:05"))
		}
	}
}

func main() {
	zone_name, _ := time.Now().Zone()
	log.Printf("%s starting, Local Time=%s Timezone=%s", TOPIC, time.Now().Local().Format("15:04:05"), zone_name)

	scheduler := gocron.NewScheduler(time.Now().Location())

	if config.Latitude != 0 && config.Longitude != 0 {
		scheduler.Every(1).Day().At("00:00").Do(setDailyTimes, scheduler)

		// startup: set timers for today once
		job, _ := scheduler.Every(1).Second().Do(setDailyTimes, scheduler)
		job.LimitRunsTo(1)
	} else {
		log.Println("Warning: Latitude and Longitude not set, sunrise/sunset cannot be used")
	}

	startMqttClient()

	setTimers(scheduler)
	scheduler.StartAsync()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}

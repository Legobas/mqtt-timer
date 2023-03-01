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
var dailyTimers []Timer

func init() {
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(0)
}

func handleEvent(timer Timer) {
	if timer.RandomBefore != "" || timer.After != "" || timer.RandomAfter != "" {
		time.Sleep(offsetDuration(timer))
	}
	log.Printf("%s: %s %s - %s", timer.Id, offsetDescr(timer), timer.Time, timer.Description)

	topic := TOPIC + "/" + timer.Id + "/time"
	msg := time.Now().Format("2006-01-02 15:04:05")
	sendToMttRetain(topic, msg)

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
				s.Cron(timer.Cron).Tag(timer.Id).Do(handleEvent, timer)
			} else if len(strings.Split(timer.Cron, " ")) == 6 {
				// Cron with Seconds
				log.Printf("Scheduled '%s' Cron [%s] '%s'", timer.Id, timer.Cron, timer.Description)
				s.CronWithSeconds(timer.Cron).Tag(timer.Id).Do(handleEvent, timer)
			} else {
				log.Printf("Invalid Cron format: [%s]", timer.Cron)
			}
		} else if timer.Time != "" {
			// Time
			days := "daily"
			if timer.Days != "" {
				days = timer.Days
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
				schedule.At(schedTime).Tag(timer.Id).Do(handleEvent, timer)

				log.Printf("Scheduled '%s' %s %s %s '%s'", timer.Id, days, offsetDescr(timer), timer.Time, timer.Description)
			} else if timer.Time == "sunrise" || timer.Time == "sunset" {
				dailyTimers = append(dailyTimers, timer)
				log.Printf("Scheduled '%s' %s %s %s '%s'", timer.Id, days, offsetDescr(timer), timer.Time, timer.Description)
			} else {
				log.Printf("Invalid config [%v]", timer)
			}
		} else {
			log.Printf("Invalid config [%v]", timer)
		}
	}
}

func offsetDescr(timer Timer) string {
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
			offset = int64(rand.Intn(seconds)) * int64(1000000000)
		} else {
			offset = int64(seconds) * int64(1000000000)
		}
	} else if len(times) == 2 && times[1][:3] == "min" {
		minutes, _ := strconv.Atoi(times[0])
		if random {
			offset = int64(rand.Intn(minutes)) * int64(60000000000)
		} else {
			offset = int64(minutes) * int64(60000000000)
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
	sunriseStr := sunriseTime.Format("15:04")
	if sunriseTime.After(time.Now().Local()) {
		timer := Timer{}
		timer.Id = "sunrise"
		timer.Description = fmt.Sprintf("at %s", sunriseStr)
		timer.Time = "sunrise"
		job, _ := s.Every(1).Day().At(sunriseTime).Do(handleEvent, timer)
		job.LimitRunsTo(1)
		log.Printf("Sunrise today %s", timer.Description)
	}

	// Sunset
	sunsetTime := sunset.Local()
	sunsetStr := sunsetTime.Format("15:04")
	if sunsetTime.After(time.Now().Local()) {
		timer := Timer{}
		timer.Id = "sunset"
		timer.Description = fmt.Sprintf("at %s", sunsetStr)
		timer.Time = "sunset"
		job, _ := s.Every(1).Day().At(sunsetTime).Do(handleEvent, timer)
		job.LimitRunsTo(1)
		log.Printf("Sunset today %s", timer.Description)
	}

	// Daily timers
	for _, timer := range dailyTimers {
		if timer.Time == "sunrise" {
			if time.Now().Local().After(sunriseTime) {
				continue
			}
			timer.Time = sunriseStr
		} else if timer.Time == "sunset" {
			if time.Now().Local().After(sunsetTime) {
				continue
			}
			timer.Time = sunsetStr
		}
		time := timeWithOffset(timer)
		job, _ := s.Every(1).Day().At(time).Tag(timer.Id).Do(handleEvent, timer)
		job.LimitRunsTo(1)
		log.Printf("Scheduled '%s' %s %s '%s'", timer.Id, offsetDescr(timer), timer.Time, timer.Description)
	}

	// refresh status
	sendToMttRetain(TOPIC+"/status", "Online")
}

func main() {
	zone_name, _ := time.Now().Zone()
	log.Printf("%s start, Local Time=%s Timezone=%s", TOPIC, time.Now().Local().Format("15:04:05"), zone_name)

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
	log.Printf("%s stop, Local Time=%s Timezone=%s", TOPIC, time.Now().Local().Format("15:04:05"), zone_name)
}

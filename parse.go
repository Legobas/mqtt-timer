package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func parseDuration(timeExpr string) int {
	seconds := 0
	times := strings.Split(timeExpr, " ")
	if len(times) == 2 && len(times[1]) > 2 && times[1][:3] == "sec" {
		seconds, _ = strconv.Atoi(times[0])
	} else if len(times) == 2 && len(times[1]) > 2 && times[1][:3] == "min" {
		minutes, _ := strconv.Atoi(times[0])
		seconds = minutes * 60
	} else if len(times) == 2 && len(times[1]) > 3 && times[1][:4] == "hour" {
		minutes, _ := strconv.Atoi(times[0])
		seconds = minutes * 3600
	}
	return seconds
}

func parseStart(startStr string) (time.Time, error) {
	startTime := time.Now().Local().Add(time.Duration(int64(1000000000)))
	var err error

	if startStr != "" {
		matchTime, _ := regexp.Match("^\\d{1,2}(:\\d{2}){1,2}$", []byte(startStr))
		if matchTime {
			startTime, err = time.Parse("15:04", startStr)
			if err != nil {
				startTime, err = time.Parse("15:04:05", startStr)
				if err != nil {
					return startTime, fmt.Errorf("Error: invalid time format: %s", startStr)
				}
			}
		} else if strings.Contains(startStr, "sec") || strings.Contains(startStr, "min") || strings.Contains(startStr, "hour") {
			seconds := parseDuration(startStr)
			if seconds > 0 {
				offset := time.Duration(int64(seconds) * int64(1000000000))
				startTime = time.Now().Local().Add(offset)
			} else {
				return startTime, fmt.Errorf("Invalid duration: %s", startStr)
			}
		} else {
			return startTime, fmt.Errorf("Invalid start time: %s", startStr)
		}
	}
	return startTime, err
}

func parseInterval(intervalStr string, messages []string) time.Duration {
	// default interval is 30 sec.
	interval := time.Duration(int64(30) * int64(1000000000))
	if intervalStr != "" {
		seconds := parseDuration(intervalStr)
		if seconds > 0 {
			interval = time.Duration(int64(seconds) * int64(1000000000))
		} else {
			log.Printf("Invalid time interval: %s", intervalStr)
		}
	} else {
		if len(messages) > 1 {
			log.Println("Warning: no interval set, default interval is 30 seconds")
		}
	}
	return interval
}

func parseUntil(untilStr string, startTime time.Time) (int, time.Time) {
	until := 1
	untilTime := time.Now().Local()
	var err error
	if untilStr != "" {
		matchTime, _ := regexp.Match("^\\d{1,2}(:\\d{2}){1,2}$", []byte(untilStr))
		matchTimes, _ := regexp.Match("^\\d*( time| times){0,1}$", []byte(untilStr))
		if matchTime {
			untilTime, err = time.Parse("15:04", untilStr)
			if err != nil {
				untilTime, err = time.Parse("15:04:05", untilStr)
				if err != nil {
					log.Printf("Error: 'until' invalid time format: %s", untilStr)
				} else {
					until = -1
				}
			} else {
				until = -1
			}
		} else if strings.Contains(untilStr, "sec") || strings.Contains(untilStr, "min") || strings.Contains(untilStr, "hour") {
			seconds := parseDuration(untilStr)
			if seconds > 0 {
				offset := time.Duration(int64(seconds) * int64(1000000000))
				untilTime = startTime.Add(offset)
				until = -1
			} else {
				log.Printf("Invalid 'until' duration: %s", untilStr)
			}
		} else if matchTimes {
			times := strings.Split(untilStr, " ")
			until, _ = strconv.Atoi(times[0])
		} else {
			log.Printf("Invalid 'until': %s", untilStr)
		}
	}
	return until, untilTime
}

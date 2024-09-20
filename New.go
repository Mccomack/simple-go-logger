package logger

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func getTimeStr() string {
	nowTime := time.Now().Local()

	hourStr := strconv.Itoa(nowTime.Hour())

	if len(hourStr) != 2 {
		hourStr = "0" + hourStr
	}

	minuteStr := strconv.Itoa(nowTime.Minute())

	if len(minuteStr) != 2 {
		minuteStr = "0" + minuteStr
	}

	secondStr := strconv.Itoa(nowTime.Second())

	if len(secondStr) != 2 {
		secondStr = "0" + secondStr
	}

	return fmt.Sprintf("%s:%s:%s", hourStr, minuteStr, secondStr)
}

type Logger struct {
	Log     func(logType string, log ...string)
	GetLogs func() map[int]string
}

func New(name string) Logger {
	var Logs map[int]string = make(map[int]string)

	var Logger Logger = Logger{
		Log: func(logType string, log ...string) {
			nowTime := getTimeStr()

			Log := fmt.Sprintf("[%s] [%s/%s]: %s", nowTime, name, logType, strings.Join(log, " "))

			Logs[len(Logs)] = Log

			fmt.Println(Log)
		},
		GetLogs: func() map[int]string {
			var cloned map[int]string = make(map[int]string)

			for k, v := range Logs {
				cloned[k] = v
			}

			return cloned
		},
	}

	return Logger
}

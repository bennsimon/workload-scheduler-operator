package util

import (
	workloadschedulerv1 "bennsimon.github.io/workload-scheduler-operator/api/v1"
	"fmt"
	"strings"
	"time"
)

const (
	DEPLOYMENT  = "deployment"
	STATEFULSET = "statefulset"
	ALL         = "*"
)

func ProcessScheduleTimeUnit(timeUnit workloadschedulerv1.TimeUnit, today time.Time) (time.Time, error) {
	_format := time.DateTime
	_time := timeUnit.Time
	_date := timeUnit.Date
	if len(strings.Trim(timeUnit.Date, "")) == 0 {
		_date = today.Format(time.DateOnly)
	} else {
		if strings.Contains(_date, "y") {
			_date = strings.Replace(_date, "y", fmt.Sprintf("%04d", today.Year()), 1)
		}

		if strings.Contains(_date, "m") {
			_date = strings.Replace(_date, "m", fmt.Sprintf("%02d", int(today.Month())), 1)
		}

		if strings.Contains(_date, "d") {
			_date = strings.Replace(_date, "d", fmt.Sprintf("%02d", today.Day()), 1)
		}
	}
	if len(strings.Trim(timeUnit.Time, "")) == 0 {
		_time = time.Time{}.Format(time.TimeOnly)
	}
	_time = fmt.Sprintf("%s %s", _date, _time)

	parsedTime, err := time.ParseInLocation(_format, _time, today.Location())
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}

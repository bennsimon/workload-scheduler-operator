package util

import (
	v1 "bennsimon.github.io/workload-scheduler-operator/api/v1"
	"reflect"
	"testing"
	"time"
)

func TestProcessScheduleTimeUnit(t *testing.T) {
	type args struct {
		timeUnit v1.TimeUnit
		today    time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{name: "should return todays' date and default time(00:00:00) if time and date are not specified.",
			args: args{today: time.Date(2023, 07, 21, 0, 0, 0, 0, time.UTC), timeUnit: v1.TimeUnit{}}, want: time.Date(2023, 07, 21, 0, 0, 0, 0, time.UTC), wantErr: false},
		{name: "should return error when date is not formatted correctly.",
			args: args{today: time.Date(2023, 07, 21, 0, 0, 0, 0, time.UTC), timeUnit: v1.TimeUnit{Date: "2023-13-12"}}, want: time.Time{}, wantErr: true},
		{name: "should return error when time is not formatted correctly.",
			args: args{today: time.Date(2023, 07, 21, 0, 0, 0, 0, time.UTC), timeUnit: v1.TimeUnit{Time: "20:23:20T234"}}, want: time.Time{}, wantErr: true},
		{name: "should return expected time when timeUnit parsed successfully.",
			args: args{today: time.Date(2023, 07, 21, 0, 0, 0, 0, time.UTC), timeUnit: v1.TimeUnit{Time: "20:23:20", Date: "2023-07-21"}}, want: time.Date(2023, 7, 21, 20, 23, 20, 0, time.UTC), wantErr: false},
		{name: "should return expected time when timeUnit parsed with placeholder successfully.",
			args: args{today: time.Date(2023, 07, 21, 0, 0, 0, 0, time.UTC), timeUnit: v1.TimeUnit{Time: "20:23:20", Date: "y-m-d"}}, want: time.Date(2023, 7, 21, 20, 23, 20, 0, time.UTC), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessScheduleTimeUnit(tt.args.timeUnit, tt.args.today)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessScheduleTimeUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessScheduleTimeUnit() got = %v, want %v", got, tt.want)
			}
		})
	}
}

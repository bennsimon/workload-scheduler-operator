package scheduleHandler

import (
	workloadschedulerv1 "bennsimon.github.io/workload-scheduler-operator/api/v1"
	"bennsimon.github.io/workload-scheduler-operator/util"
	"context"
	"fmt"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type IScheduleHandler interface {
	getScheduleByName(schedule string, r client.Reader, ctx context.Context) (*workloadschedulerv1.Schedule, error)
	FetchWorkloadSchedules(schedules []workloadschedulerv1.WorkloadScheduleUnit, r client.Reader, ctx context.Context) ([]workloadschedulerv1.Schedule, error)
	IsThisDayIncluded(days []string, now time.Time) bool
	ValidateSchedule(schedule *workloadschedulerv1.Schedule) error
}

type ScheduleHandler struct {
	IScheduleHandler
}

func New() *ScheduleHandler {
	return &ScheduleHandler{IScheduleHandler: &ScheduleHandler{}}
}

func (s *ScheduleHandler) ValidateSchedule(schedule *workloadschedulerv1.Schedule) error {
	if schedule.Spec.ScheduleUnits == nil || len(schedule.Spec.ScheduleUnits) == 0 {
		return fmt.Errorf("schedule(s) need to be defined")
	} else {
		now := time.Now().In(time.Local)
		longDayNames := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}

		for _, scheduleUnit := range schedule.Spec.ScheduleUnits {
			if scheduleUnit.Days != nil && len(scheduleUnit.Days) > 0 {
				for _, day := range scheduleUnit.Days {
					if !slices.Contains(longDayNames, strings.ToLower(day)) {
						return fmt.Errorf("day: %s, is not valid", day)
					}
				}
			}
			startTime, err := util.ProcessScheduleTimeUnit(scheduleUnit.Start, now)
			if err != nil {
				return err
			}
			endTime, err := util.ProcessScheduleTimeUnit(scheduleUnit.End, now)
			if err != nil {
				return err
			}

			if startTime.After(endTime) || startTime.Equal(endTime) {
				return fmt.Errorf("invalid timeunit; startTime: %s, endTime: %s", startTime, endTime)
			}
		}
	}
	return nil
}
func (s *ScheduleHandler) IsThisDayIncluded(days []string, now time.Time) bool {
	isThisDayIncluded := false
	for _, day := range days {
		if strings.ToLower(now.Weekday().String()) == strings.ToLower(day) {
			isThisDayIncluded = true
			break
		}
	}
	return isThisDayIncluded
}

func (s *ScheduleHandler) FetchWorkloadSchedules(_schedules []workloadschedulerv1.WorkloadScheduleUnit, r client.Reader, ctx context.Context) ([]workloadschedulerv1.Schedule, error) {
	var schedules []workloadschedulerv1.Schedule
	for _, _schedule := range _schedules {
		schedule, err := s.IScheduleHandler.getScheduleByName(_schedule.Schedule, r, ctx)
		if err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("error when fetching: %v schedule", _schedule), err)
		} else {
			schedules = append(schedules, *schedule)
		}
	}
	return schedules, nil
}

func (s *ScheduleHandler) getScheduleByName(_schedule string, r client.Reader, ctx context.Context) (*workloadschedulerv1.Schedule, error) {
	schedule := &workloadschedulerv1.Schedule{}
	err := r.Get(ctx, client.ObjectKey{Name: _schedule}, schedule)
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

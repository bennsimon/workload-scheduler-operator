package scheduleHandler

import (
	v1 "bennsimon.github.io/workload-scheduler-operator/api/v1"
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

func TestScheduleHandler_IsThisDayIncluded(t *testing.T) {
	type args struct {
		days []string
		now  time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "should return true since 2023/7/21 is a Friday", args: args{days: []string{"Friday"}, now: time.Date(2023, 07, 21, 0, 0, 0, 0, time.UTC)}, want: true},
		{name: "should return false since 2023/7/21 is a Friday", args: args{days: []string{"Monday"}, now: time.Date(2023, 07, 21, 0, 0, 0, 0, time.UTC)}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScheduleHandler{}
			if got := s.IsThisDayIncluded(tt.args.days, tt.args.now); got != tt.want {
				t.Errorf("IsThisDayIncluded() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testReader struct {
	mock.Mock
	client.Reader
}

func (t *testReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	args := t.Called(ctx, key, obj, opts)
	return args.Error(0)
}

func TestScheduleHandler_getSchedulesByName(t *testing.T) {
	var testreader *testReader

	type args struct {
		schedule string
		ctx      context.Context
	}
	//sd := v1.Schedule{
	//	Spec: v1.ScheduleSpec{
	//		ScheduleUnits: []v1.ScheduleUnit{v1.ScheduleUnit{Start: v1.TimeUnit{Date: "2023-07-23"}}}},
	//}
	//sch := &v1.ScheduleList{Items: []v1.Schedule{sd}}

	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		want        *v1.Schedule
		wantErr     bool
	}{
		{name: "should return nil when schedule is not found.", setupMocks: func() {
			testreader = &testReader{}
			testreader.On("Get", nil, mock.IsType(client.ObjectKey{Name: "test-schedule"}), mock.IsType(&v1.Schedule{}), []client.GetOption(nil)).Return(fmt.Errorf("some error"))
		}, verifyMocks: func() {
			testreader.AssertExpectations(t)
		}, args: args{schedule: "test-schedule", ctx: nil}, want: nil, wantErr: true},
		//{name: "should return schedule when found.", setupMocks: func() {
		//	testreader = &testReader{}
		//
		//	testreader.On("List", nil, sch, []client.ListOption{client.MatchingFields{config.IndexedField: "test-schedule"}}).Return(nil)
		//}, verifyMocks: func() {
		//	testreader.AssertExpectations(t)
		//}, args: args{schedule: "test-schedule", ctx: nil}, want: &v1.Schedule{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScheduleHandler{}
			tt.setupMocks()
			defer tt.verifyMocks()
			got, err := s.getScheduleByName(tt.args.schedule, testreader, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getScheduleByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getScheduleByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type testScheduleHandler struct {
	mock.Mock
	IScheduleHandler
}

func (s *testScheduleHandler) getScheduleByName(schedule string, r client.Reader, ctx context.Context) (*v1.Schedule, error) {
	args := s.Called(schedule, r, ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	} else {
		return args.Get(0).(*v1.Schedule), args.Error(1)
	}
}

func TestScheduleHandler_CheckIfSchedulesExist(t *testing.T) {
	var testschedulehandler *testScheduleHandler
	s := &ScheduleHandler{}
	type args struct {
		_schedules []v1.WorkloadScheduleUnit
		r          client.Reader
		ctx        context.Context
	}
	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		want        []v1.Schedule
		wantErr     bool
	}{
		{name: "should return nil schedules when error occurs when getting schedule.", setupMocks: func() {
			testschedulehandler = &testScheduleHandler{}
			testschedulehandler.On("getScheduleByName", mock.IsType("weekday"), mock.IsType(&testReader{}), mock.IsType(context.TODO())).Return(nil, fmt.Errorf("some error"))
			s.IScheduleHandler = testschedulehandler
		}, verifyMocks: func() {
			testschedulehandler.AssertExpectations(t)
		}, args: args{_schedules: []v1.WorkloadScheduleUnit{
			{
				Schedule: "weekday",
				Desired:  0,
			},
		}, r: &testReader{}, ctx: context.TODO()},
			want: nil, wantErr: true},
		{name: "should return schedules when schedule retrieved successfully.", setupMocks: func() {
			testschedulehandler = &testScheduleHandler{}
			testschedulehandler.On("getScheduleByName", mock.IsType("weekday"), mock.IsType(&testReader{}), mock.IsType(context.TODO())).Return(&v1.Schedule{}, nil)
			s.IScheduleHandler = testschedulehandler
		}, verifyMocks: func() {
			testschedulehandler.AssertExpectations(t)
		}, args: args{_schedules: []v1.WorkloadScheduleUnit{
			{
				Schedule: "weekday",
				Desired:  0,
			},
		}, r: &testReader{}, ctx: context.TODO()},
			want: []v1.Schedule{{}}, wantErr: false},
		{name: "should return empty schedules when scheduleUnits is empty.", setupMocks: func() {
			testschedulehandler = &testScheduleHandler{}
			s.IScheduleHandler = testschedulehandler
		}, verifyMocks: func() {
			testschedulehandler.AssertExpectations(t)
		}, args: args{_schedules: []v1.WorkloadScheduleUnit{}, r: &testReader{}, ctx: context.TODO()},
			want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			got, err := s.FetchWorkloadSchedules(tt.args._schedules, tt.args.r, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchWorkloadSchedules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchWorkloadSchedules() got = %v, want %v", got, tt.want)
			}
		})
	}
}

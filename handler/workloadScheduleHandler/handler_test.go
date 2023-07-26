package workloadScheduleHandler

import (
	v1 "bennsimon.github.io/workload-scheduler-operator/api/v1"
	"bennsimon.github.io/workload-scheduler-operator/handler/scheduleHandler"
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestWorkloadScheduleHandler_getDesired(t *testing.T) {
	type args struct {
		schedule  v1.Schedule
		schedules []v1.WorkloadScheduleUnit
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		{name: "should return error when matching schedule is not found.", args: args{schedule: v1.Schedule{ObjectMeta: metav1.ObjectMeta{Name: "test-schedule-1"}},
			schedules: []v1.WorkloadScheduleUnit{{Schedule: "test-schedule-2"}}}, want: 0, wantErr: true},
		{name: "should return desired value when matching schedule is found.", args: args{schedule: v1.Schedule{ObjectMeta: metav1.ObjectMeta{Name: "test-schedule-1"}},
			schedules: []v1.WorkloadScheduleUnit{{Schedule: "test-schedule-1", Desired: 3}}}, want: 3, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WorkloadScheduleHandler{}
			got, err := w.getDesired(tt.args.schedule, tt.args.schedules)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDesired() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getDesired() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkloadScheduleHandler_RankWorkloadScheduleBySelectors(t *testing.T) {
	type args struct {
		_specMap map[string]map[string][]v1.WorkloadScheduleData
	}
	tests := []struct {
		name string
		args args
		want []v1.WorkloadScheduleData
	}{
		{name: "should return nil if specMap is empty.", args: args{_specMap: map[string]map[string][]v1.WorkloadScheduleData{}}, want: nil},
		{name: "should return workSchedule sorted by selectors.", args: args{_specMap: map[string]map[string][]v1.WorkloadScheduleData{
			"1110": {
				"ns/deployment/test-deploy": []v1.WorkloadScheduleData{
					{
						WorkloadScheduler: "test-wscheduler-1", Name: "test-deploy", Namespace: "ns", Kind: "deployment", Desired: 0,
					},
				},
			},
			"0000": {
				"*/*/*": []v1.WorkloadScheduleData{
					{
						WorkloadScheduler: "test-wscheduler-3", Name: "*", Namespace: "*", Kind: "*", Desired: 0,
					},
				},
			},
			"0100": {
				"*/deployment/*": []v1.WorkloadScheduleData{
					{
						WorkloadScheduler: "test-wscheduler-4", Name: "*", Namespace: "*", Kind: "deployment", Desired: 0,
					},
				},
			},
			"0110": {
				"*/deployment/test-deploy": []v1.WorkloadScheduleData{
					{
						WorkloadScheduler: "test-wscheduler-5", Name: "test-deploy", Namespace: "*", Kind: "deployment", Desired: 0,
					},
				},
			},
		}}, want: []v1.WorkloadScheduleData{
			{
				WorkloadScheduler: "test-wscheduler-1", Name: "test-deploy", Namespace: "ns", Kind: "deployment", Desired: 0,
			},
			{
				WorkloadScheduler: "test-wscheduler-5", Name: "test-deploy", Namespace: "*", Kind: "deployment", Desired: 0,
			},
			{
				WorkloadScheduler: "test-wscheduler-4", Name: "*", Namespace: "*", Kind: "deployment", Desired: 0,
			},
			{
				WorkloadScheduler: "test-wscheduler-3", Name: "*", Namespace: "*", Kind: "*", Desired: 0,
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := New()
			got := w.RankWorkloadScheduleBySelectors(tt.args._specMap)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RankWorkloadScheduleBySelectors() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type testScheduleHandler struct {
	mock.Mock
	scheduleHandler.IScheduleHandler
}

func (t *testScheduleHandler) FetchWorkloadSchedules(schedules []v1.WorkloadScheduleUnit, r client.Reader, ctx context.Context) ([]v1.Schedule, error) {
	args := t.Called(schedules, r, ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	} else {
		return args.Get(0).([]v1.Schedule), args.Error(1)
	}
}

func TestWorkloadScheduleHandler_EvaluateWorkloadSchedulers(t *testing.T) {
	var testschedulehandler *testScheduleHandler
	w := &WorkloadScheduleHandler{}

	type args struct {
		workloadSchedulers *v1.WorkloadScheduleList
		r                  client.Reader
		ctx                context.Context
	}
	tests := []struct {
		name        string
		setupMocks  func()
		verifyMocks func()
		args        args
		want        map[string][]v1.Schedule
		want1       map[string]v1.WorkloadSchedule
	}{
		{name: "should return empty maps with no error if WorkloadScheduleList is empty.", setupMocks: func() {
		}, verifyMocks: func() {
		}, args: args{
			ctx:                nil,
			r:                  nil,
			workloadSchedulers: &v1.WorkloadScheduleList{},
		},
			want: map[string][]v1.Schedule{}, want1: map[string]v1.WorkloadSchedule{},
		},
		{name: "should return empty maps with an error if error found on a WorkloadSchedule.", setupMocks: func() {
			testschedulehandler = &testScheduleHandler{}
			testschedulehandler.On("FetchWorkloadSchedules", []v1.WorkloadScheduleUnit{}, nil, nil).Return(nil, fmt.Errorf("some error"))
			w.ScheduleHandler = testschedulehandler
		}, verifyMocks: func() {
		}, args: args{
			ctx:                nil,
			r:                  nil,
			workloadSchedulers: &v1.WorkloadScheduleList{Items: []v1.WorkloadSchedule{{Spec: v1.WorkloadScheduleSpec{Schedules: []v1.WorkloadScheduleUnit{}}}}},
		},
			want: map[string][]v1.Schedule{}, want1: map[string]v1.WorkloadSchedule{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			got, got1 := w.EvaluateWorkloadSchedulers(tt.args.workloadSchedulers, tt.args.r, tt.args.ctx)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvaluateWorkloadSchedulers() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("EvaluateWorkloadSchedulers() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestWorkloadScheduleHandler_BuildSpecMap(t *testing.T) {
	type args struct {
		_workloadSchedule v1.WorkloadSchedule
		specMap           map[string]map[string][]v1.WorkloadScheduleData
		schedule          v1.Schedule
	}

	tests := []struct {
		name string
		args args
		want map[string]map[string][]v1.WorkloadScheduleData
	}{
		{name: "should update specMap accordingly.", args: args{
			specMap: make(map[string]map[string][]v1.WorkloadScheduleData),
			_workloadSchedule: v1.WorkloadSchedule{
				ObjectMeta: metav1.ObjectMeta{Name: "test-workload-scheduler"},
				Spec: v1.WorkloadScheduleSpec{
					Schedules: []v1.WorkloadScheduleUnit{
						{
							Schedule: "test-schedule", Desired: 0},
					},
					Selector: v1.WorkloadSelector{
						Names: []string{"test-deploy"}},
				},
			},
			schedule: v1.Schedule{ObjectMeta: metav1.ObjectMeta{Name: "test-schedule"}},
		}, want: map[string]map[string][]v1.WorkloadScheduleData{
			"0010": {
				"*/deployment/test-deploy":  []v1.WorkloadScheduleData{{Name: "test-deploy", Namespace: "*", Kind: "deployment", Desired: 0, WorkloadScheduler: "test-workload-scheduler"}},
				"*/statefulset/test-deploy": []v1.WorkloadScheduleData{{Name: "test-deploy", Namespace: "*", Kind: "statefulset", Desired: 0, WorkloadScheduler: "test-workload-scheduler"}},
			},
		},
		},
		{name: "should update specMap with wildcards.", args: args{
			specMap: make(map[string]map[string][]v1.WorkloadScheduleData),
			_workloadSchedule: v1.WorkloadSchedule{
				ObjectMeta: metav1.ObjectMeta{Name: "test-workload-scheduler"},
				Spec: v1.WorkloadScheduleSpec{
					Schedules: []v1.WorkloadScheduleUnit{
						{
							Schedule: "test-schedule", Desired: 0},
					},
					Selector: v1.WorkloadSelector{}},
			},
			schedule: v1.Schedule{ObjectMeta: metav1.ObjectMeta{Name: "test-schedule"}},
		}, want: map[string]map[string][]v1.WorkloadScheduleData{
			"0000": {
				"*/deployment/*":  []v1.WorkloadScheduleData{{Name: "*", Namespace: "*", Kind: "deployment", Desired: 0, WorkloadScheduler: "test-workload-scheduler"}},
				"*/statefulset/*": []v1.WorkloadScheduleData{{Name: "*", Namespace: "*", Kind: "statefulset", Desired: 0, WorkloadScheduler: "test-workload-scheduler"}},
			},
		},
		},
		{name: "should update specMap with specified selectors.", args: args{
			specMap: make(map[string]map[string][]v1.WorkloadScheduleData),
			_workloadSchedule: v1.WorkloadSchedule{
				ObjectMeta: metav1.ObjectMeta{Name: "test-workload-scheduler"},
				Spec: v1.WorkloadScheduleSpec{
					Schedules: []v1.WorkloadScheduleUnit{
						{
							Schedule: "test-schedule", Desired: 0},
					},
					Selector: v1.WorkloadSelector{Names: []string{"test-deploy"}, Kinds: []string{"deployment"}, Namespaces: []string{"default"}}},
			},
			schedule: v1.Schedule{ObjectMeta: metav1.ObjectMeta{Name: "test-schedule"}},
		}, want: map[string]map[string][]v1.WorkloadScheduleData{
			"1110": {
				"default/deployment/test-deploy": []v1.WorkloadScheduleData{{Name: "test-deploy", Namespace: "default", Kind: "deployment", Desired: 0, WorkloadScheduler: "test-workload-scheduler"}},
			},
		},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WorkloadScheduleHandler{}
			w.BuildSpecMap(tt.args._workloadSchedule, tt.args.specMap, tt.args.schedule)
			if !reflect.DeepEqual(tt.args.specMap, tt.want) {
				t.Errorf("BuildSpecMap() got = %v want %v", tt.args.specMap, tt.want)
			}
		})
	}
}

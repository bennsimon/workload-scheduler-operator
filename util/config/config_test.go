package config

import (
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
	"time"
)

type testConfig struct {
	Provider
	mock.Mock
}

func (t *testConfig) LookUpEnv(env string) (string, bool) {
	args := t.Called(env)
	return args.String(0), args.Bool(1)
}

func TestConfig_InitializeEnvs(t *testing.T) {
	var testconfig *testConfig
	config := &Config{}
	tests := []struct {
		name        string
		want        map[string]string
		setupMocks  func()
		verifyMocks func()
	}{
		{name: "should initialize with 'default' namespaces.", setupMocks: func() {
			testconfig = &testConfig{}
			testconfig.On("LookUpEnv", NamespacesOffLimits).Return("", false)
			config.Provider = testconfig
		}, verifyMocks: func() {
			testconfig.AssertExpectations(t)
		}, want: map[string]string{"kube-system": "kube-system"}},
		{name: "should initialize with default and provided namespaces.", setupMocks: func() {
			testconfig = &testConfig{}
			testconfig.On("LookUpEnv", NamespacesOffLimits).Return("NS1", true)
			config.Provider = testconfig
		}, verifyMocks: func() {
			testconfig.AssertExpectations(t)
		}, want: map[string]string{"kube-system": "kube-system", "NS1": "NS1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			if got := config.GetIgnoredNamespacesMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeEnvs() = %v, want %v", ignoredNamespacesMap, tt.want)
			}
		})
	}
}

func TestConfig_GetTimeLocationConfig(t *testing.T) {
	var testconfig *testConfig
	config := &Config{}
	location, _ := time.LoadLocation("Africa/Nairobi")
	tests := []struct {
		name        string
		setupMocks  func()
		verifyMocks func()
		want        *time.Location
	}{
		{name: "should use local timezone", setupMocks: func() {
			testconfig = &testConfig{}
			config.Provider = testconfig
			testconfig.On("LookUpEnv", TIMEZONE).Return("", false)
		}, verifyMocks: func() {
			testconfig.AssertExpectations(t)
		}, want: time.Local},
		{name: "should return provided timezone", setupMocks: func() {
			testconfig = &testConfig{}
			config.Provider = testconfig
			testconfig.On("LookUpEnv", TIMEZONE).Return("Africa/Nairobi", true)
		}, verifyMocks: func() {
			testconfig.AssertExpectations(t)
		}, want: location},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			if got := config.GetTimeLocationConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTimeLocationConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

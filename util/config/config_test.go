package config

import (
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
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

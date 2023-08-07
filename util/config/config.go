package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	MapKeySeparator        = "/"
	IndexedField           = "metadata.name"
	NamespacesOffLimits    = "NAMESPACES_OFF_LIMITS"
	ReconciliationDuration = "RECONCILIATION_DURATION"
	Debug                  = "DEBUG"
)

type Provider interface {
	LookUpEnv(env string) (string, bool)
}

type Config struct {
	Provider
}

func New() *Config {
	return &Config{Provider: &Config{}}
}

func (c *Config) LookUpEnv(env string) (string, bool) {
	return os.LookupEnv(env)
}

func (c *Config) LookUpIntEnv(env string) (int, error) {
	val, exists := c.Provider.LookUpEnv(env)
	if exists {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return 0, err
		}
		return intVal, nil
	}
	return 0, fmt.Errorf("env %s does not exist", env)
}

func (c *Config) LookUpBooleanEnv(env string) bool {
	val, exists := c.Provider.LookUpEnv(env)
	if exists {
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			return false
		}
		return boolVal
	}
	return false
}

var ignoredNamespacesMap = make(map[string]string)

func (c *Config) InitializeEnvs() {
	if ignoredNamespaces, exists := c.Provider.LookUpEnv(NamespacesOffLimits); exists {
		ignoredNamespacesArr := strings.Split(ignoredNamespaces, ",")
		for _, ins := range ignoredNamespacesArr {
			ignoredNamespacesMap[ins] = ins
		}
	}
	//Ignore kube-system by default
	ignoredNamespacesMap["kube-system"] = "kube-system"
}

func (c *Config) GetIgnoredNamespacesMap() map[string]string {
	c.InitializeEnvs()
	return ignoredNamespacesMap
}

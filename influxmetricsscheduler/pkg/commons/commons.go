package commons

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"os"
	"strconv"
	"strings"
)

type MetricParams struct {
	MetricName        string
	StartDate         string
	EndDate           string
	Operation         string
	PriorityOrder     string
	FilterClause      string
	IsSecondLevel     string
	SecondLevelGroup  string
	SecondLevelSelect string
}

type SchedulerParams struct {
	MetricParams  MetricParams
	Influxdb      InfluxdbParams
	SchedulerName string
	Timeout       int
	LogLevel      string
	FilteredNodes string
}

type InfluxdbParams struct {
	Host         string
	Port         string
	Token        string
	Organization string
	Bucket       string
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func GetInternalIpsSlice(nodes []*v1.Node) []string {
	var ipSlice []string
	for _, node := range nodes {
		for _, address := range node.Status.Addresses {
			if string(address.Type) == "InternalIP" {
				ipSlice = append(ipSlice, address.Address)
			}
		}
	}
	return ipSlice
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrStringSlice(key string, defaultVal []string) []string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return strings.Split(val, ",")
	}
	return defaultVal
}

func lookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			log.Error(err, "LookupEnvOrBool", "key", key, "value", val)
		}
		return v
	}
	return defaultVal
}

func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Error(err, "LookupEnvOrBool", "key", key, "value", val)
		}
		return v
	}
	return defaultVal
}

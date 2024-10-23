package pkg

import v1 "k8s.io/api/core/v1"

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
	MetricParams      MetricParams
	TimescaleDbParams TimescaleDbParams
	SchedulerName     string
	Timeout           int
	LogLevel          string
	FilteredNodes     string
}

type TimescaleDbParams struct {
	Host               string
	Port               string
	User               string
	Password           string
	Database           string
	AuthenticationType string
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

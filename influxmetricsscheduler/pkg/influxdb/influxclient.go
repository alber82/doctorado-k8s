package influxdb

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"main/pkg/commons"
)

type DatabaseClient struct {
	Params commons.InfluxdbParams
}

// ConnectionParams ...
type ConnectionParams struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	Token        string `json:"token"`
	Organization string `json:"organization"`
	Bucket       string `json:"bucket"`
}

func (databaseClient *DatabaseClient) getConnectionParams() ConnectionParams {
	influxConnection := ConnectionParams{
		Host:         databaseClient.Params.Host,
		Port:         databaseClient.Params.Port,
		Token:        databaseClient.Params.Token,
		Organization: databaseClient.Params.Organization,
		Bucket:       databaseClient.Params.Bucket,
	}

	return influxConnection
}

func (databaseClient *DatabaseClient) GetMetrics(metricsParams commons.MetricParams) (map[string]int, error) {
	dbConnectionParams := databaseClient.getConnectionParams()

	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient(fmt.Sprintf("http://%s:%s", dbConnectionParams.Host, dbConnectionParams.Port), dbConnectionParams.Token)
	// Get query client
	queryAPI := client.QueryAPI(dbConnectionParams.Organization)
	log.Info("Connected to InfluxDB")
	result, err := queryAPI.Query(context.Background(), `import "math"

First = from(bucket: "doctorado")
  |> range(start: -20m, stop: -10m)
  |> filter(fn: (r) => r["_measurement"] == "prometheus_remote_write")
  |> filter(fn: (r) => r["_field"] == "node_network_receive_bytes_total")
  |> filter(fn: (r) => r["device"] == "eth0")
  |> group(columns: ["instance"], mode:"by")
  |> keep(columns: ["instance", "_value"])
  |> first()

Last = from(bucket: "doctorado")
  |> range(start: -1h, stop: now())
  |> filter(fn: (r) => r["_measurement"] == "prometheus_remote_write")
  |> filter(fn: (r) => r["_field"] == "node_network_receive_bytes_total")
  |> filter(fn: (r) => r["device"] == "eth0")
  |> group(columns: [ "instance"], mode:"by")
  |> keep(columns: ["instance", "_value"])
  |> last()

union(tables: [ First, Last])
|> difference()
|> map(fn: (r) => ({r with _value: math.abs(x: r._value)}))`)

	if err == nil {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				log.Info(fmt.Printf("table: %s\n", result.TableMetadata().String()))
			}
			// Access data
			log.Info(fmt.Printf("value: %v\n", result.Record().Value()))
		}
		// check for an error
		if result.Err() != nil {
			log.Info(fmt.Printf("query parsing error: %s\n", result.Err().Error()))
		}
	} else {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()

	var priorityMap = make(map[string]int)

	//for _, m := range rowsArray {
	//	fmt.Println("Node ", m.node, ", metric value ", m.value)
	//	priorityMap[m.node] = m.rowid
	//}

	return priorityMap, nil
}

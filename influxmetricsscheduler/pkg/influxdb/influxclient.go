package influxdb

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"main/pkg/commons"
	"reflect"
	"strings"
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

func (databaseClient *DatabaseClient) GetMetrics(metricsParams commons.MetricParams) (map[string]int64, error) {
	dbConnectionParams := databaseClient.getConnectionParams()

	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient(fmt.Sprintf("http://%s:%s", dbConnectionParams.Host, dbConnectionParams.Port), dbConnectionParams.Token)
	// Get query client
	queryAPI := client.QueryAPI(dbConnectionParams.Organization)
	log.Info("Connected to InfluxDB")

	var priorityMap = make(map[string]int64)

	query := fmt.Sprintf(`import "math"
`)

	if !metricsParams.IsSecondLevel {
		switch metricsParams.Operation {
		case "first", "last", "max", "min", "mean", "median", "sum", "spread":
			query += fmt.Sprintf(`from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r["_measurement"] == "%s")
`,
				dbConnectionParams.Bucket,
				metricsParams.StartDate,
				metricsParams.EndDate,
				metricsParams.MetricName)

			for _, filter := range strings.Split(metricsParams.FilterClause, ",") {
				query += fmt.Sprintf(`|> filter(%s)
`, strings.Replace(filter, "'", "\"", -1))
			}

			query += fmt.Sprintf(`|> group(columns: ["node_ip"], mode:"by")
	|> keep(columns: ["node_ip", "_value"])
	|> %s()
	|> yield(name: "%s")
`,
				metricsParams.Operation,
				metricsParams.Operation)

		case "difference":
			query += fmt.Sprintf(`First = from(bucket: "%s") 
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r["_measurement"] == "%s")
				`,
				dbConnectionParams.Bucket,
				metricsParams.StartDate,
				metricsParams.EndDate,
				metricsParams.MetricName)

			for _, filter := range strings.Split(metricsParams.FilterClause, ",") {
				query += fmt.Sprintf(`|> filter(%s)
`, strings.Replace(filter, "'", "\"", -1))
			}

			query += fmt.Sprintln(`	|> group(columns: ["node_ip"], mode:"by")
	|> keep(columns: ["node_ip", "_value"])
	|> first()
	|> yield(name: "first")
	
	`)

			query += fmt.Sprintf(`Last = from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r["_measurement"] == "%s")`,
				dbConnectionParams.Bucket,
				metricsParams.StartDate,
				metricsParams.EndDate,
				metricsParams.MetricName)

			for _, filter := range strings.Split(metricsParams.FilterClause, ",") {
				query += fmt.Sprintf(`	|> filter(%s)
`, strings.Replace(filter, "'", "\"", -1))
			}

			query += fmt.Sprintf(`	|> group(columns: ["node_ip"], mode:"by")
	|> keep(columns: ["node_ip", "_value"])
	|> last()
	|> yield(name: "last")
	
	union(tables: [First, Last])
	|> difference()
	|> map(fn: (r) => ({r with _value: math.abs(x: r._value)}))`)
		}
	} else {
		switch metricsParams.Operation {
		case "first", "last", "max", "min", "mean", "median", "sum", "spread":
			query += fmt.Sprintf(`%s = from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r["_measurement"] == "%s")
`,
				cases.Title(language.English, cases.Compact).String(metricsParams.SecondLevelOperation),
				dbConnectionParams.Bucket,
				metricsParams.StartDate,
				metricsParams.EndDate,
				metricsParams.MetricName)

			for _, filter := range strings.Split(metricsParams.FilterClause, ",") {
				query += fmt.Sprintf(`	|> filter(%s)
`, strings.Replace(filter, "'", "\"", -1))
			}

			query += fmt.Sprintf(`	|> group(columns: ["node_ip","%s"], mode:"by")
	|> keep(columns: ["node_ip", "%s","_value"])
	|> %s(column: "_value")
	|> yield(name: "%s")
`,
				metricsParams.SecondLevelGroup,
				metricsParams.SecondLevelGroup,
				metricsParams.SecondLevelOperation,
				metricsParams.SecondLevelOperation)

			query += fmt.Sprintf(`%s
	|> group(columns: [ "node_ip"], mode:"by")
	|> keep(columns: ["node_ip","_value"])
	|> map(fn: (r) => ({r with _value: math.abs(x: r._value)}))
	|> %s(column: "_value")`,
				cases.Title(language.English, cases.Compact).String(metricsParams.SecondLevelOperation),
				metricsParams.Operation)

		case "difference":
			query += fmt.Sprintf(`First = from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r["_measurement"] == "%s")
`,
				dbConnectionParams.Bucket,
				metricsParams.StartDate,
				metricsParams.EndDate,
				metricsParams.MetricName)

			for _, filter := range strings.Split(metricsParams.FilterClause, ",") {
				query += fmt.Sprintf(`	|> filter(%s)
`, strings.Replace(filter, "'", "\"", -1))
			}

			query += fmt.Sprintf(`	|> group(columns: ["node_ip","%s"], mode:"by")
	|> keep(columns: ["node_ip", "%s", "_value"])
	|> first()
`, metricsParams.SecondLevelGroup,
				metricsParams.SecondLevelGroup)

			query += fmt.Sprintf(`Last = from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r["_measurement"] == "%s")
`,
				dbConnectionParams.Bucket,
				metricsParams.StartDate,
				metricsParams.EndDate,
				metricsParams.MetricName)

			for _, filter := range strings.Split(metricsParams.FilterClause, ",") {
				query += fmt.Sprintf(`	|> filter(%s)
`, strings.Replace(filter, "'", "\"", -1))
			}

			query += fmt.Sprintf(`	|> group(columns: ["node_ip", "%s"], mode:"by")
	|> keep(columns: ["node_ip", "%s", "_value"])
	|> last()
`,
				metricsParams.SecondLevelGroup,
				metricsParams.SecondLevelGroup)

			query += fmt.Sprintf(`union(tables: [First, Last])
	|> difference()
	|> group(columns: [ "node_ip"], mode:"by")
	|> keep(columns: ["node_ip","_value"])
	|> map(fn: (r) => ({r with _value: math.abs(x: r._value)}))
	|> %s(column: "_value")`,
				metricsParams.SecondLevelOperation)
		}
	}

	log.Info("Query: ", query)

	result, err := queryAPI.Query(context.Background(), query)

	if err == nil {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				log.Info(fmt.Printf("table: %s\n", result.TableMetadata().String()))
			}

			float, err := getFloat(result.Record().Value())
			if err != nil {
				return nil, err
			}

			priorityMap[fmt.Sprintf("%s", result.Record().ValueByKey("node_ip"))] = int64(float)
			// Access data
			log.Info(fmt.Printf("url: %s  %f\n", result.Record().ValueByKey("node_ip"), float))
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

	log.Info(fmt.Sprintf("priorityMap %v", priorityMap))

	return priorityMap, nil
}

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

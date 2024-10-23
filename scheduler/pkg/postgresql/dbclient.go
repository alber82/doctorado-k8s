package postgresql

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"main/pkg"
	"strconv"
	"strings"
)

type DatabaseClient struct {
	Params pkg.TimescaleDbParams
}

func (databaseClient *DatabaseClient) getConnectionParams() ConnectionParams {
	dbConnection := ConnectionParams{
		AuthenticationType: AuthenticationType(databaseClient.Params.AuthenticationType),
		Host:               databaseClient.Params.Host,
		Port:               databaseClient.Params.Port,
		DbName:             databaseClient.Params.Database,
		User:               databaseClient.Params.User,
		Password:           databaseClient.Params.Password,
	}

	return dbConnection
}

func (databaseClient *DatabaseClient) GetMetrics(metricsParams pkg.MetricParams) (map[string]int, error) {
	dbConnectionParams := databaseClient.getConnectionParams()

	db, err := sql.Open("postgres", GetDatabaseConnectionString(dbConnectionParams))

	if err != nil {
		return map[string]int{}, err
	}
	var selectString string
	if ok, _ := strconv.ParseBool(metricsParams.IsSecondLevel); !ok {
		selectString = fmt.Sprint("select row_number() over() rowid, a.node, a.value",
			" from (select left(val(instance_id),-5) node, ",
			metricsParams.Operation, " value from ", metricsParams.MetricName,
			" where value <> 'Nan' ",
			" and time >=", metricsParams.StartDate,
			" and time <=", metricsParams.EndDate)

		if metricsParams.FilterClause != "" {
			filterClauseSlice := strings.Split(metricsParams.FilterClause, ",")
			filterClause := strings.Join(filterClauseSlice, " AND ")
			selectString = selectString + fmt.Sprint(" and  ", filterClause)
		}

		selectString = selectString + fmt.Sprint(" group by  val(instance_id)",
			" order by ", metricsParams.Operation, " ", metricsParams.PriorityOrder, ") as a;")

	} else {
		selectString = fmt.Sprint("select row_number() over() rowid, b.node, b.value",
			" from (select left(a.node,-5) node, ",
			metricsParams.Operation, "(a.value) value from (select ", metricsParams.SecondLevelSelect,
			" from ", metricsParams.MetricName,
			" where value <> 'Nan' ",
			" and time >=", metricsParams.StartDate,
			" and time <=", metricsParams.EndDate)

		if metricsParams.FilterClause != "" {
			filterClauseSlice := strings.Split(metricsParams.FilterClause, ",")
			filterClause := strings.Join(filterClauseSlice, " AND ")
			selectString = selectString + fmt.Sprint(" and  ", filterClause)
		}

		selectString = selectString + fmt.Sprint(" group by ", metricsParams.SecondLevelGroup, ") a",
			" group by node order by ", metricsParams.Operation, "(value) ", metricsParams.PriorityOrder, ") as b;")
	}

	fmt.Println(selectString)

	rows, err := db.Query(selectString)

	if err != nil {
		return map[string]int{}, err
	}

	rowsArray, err := rowsToMetricArray(rows)

	if err != nil {
		return map[string]int{}, err
	}

	fmt.Println("Values for metric", metricsParams.MetricName)

	var priorityMap = make(map[string]int)

	for _, m := range rowsArray {
		fmt.Println("Node ", m.node, ", metric value ", m.value)
		priorityMap[m.node] = m.rowid
	}

	return priorityMap, nil
}

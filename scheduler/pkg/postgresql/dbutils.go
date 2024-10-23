package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // PostgreSQL Drivers
	"time"
)

// AuthenticationType (Different types of security supported)
type AuthenticationType string

const (
	// AuthenticationTypeCert ()
	AuthenticationTypeCert AuthenticationType = "CERT"
	// AuthenticationTypeMD5 ()
	AuthenticationTypeMD5 AuthenticationType = "MD5"
	// AuthenticationTypeTrust ()
	AuthenticationTypeTrust AuthenticationType = "TRUST"
)

// ConnectionParams ...
type ConnectionParams struct {
	AuthenticationType AuthenticationType `json:"authenticationtype"`
	Host               string             `json:"host"`
	Port               string             `json:"port"`
	DbName             string             `json:"dbname"`
	User               string             `json:"user"`
	Password           string             `json:"password,omitempty"`
	SslRootCert        string             `json:"sslrootcert,omitempty"`
	SslKey             string             `json:"sslkey,omitempty"`
	SslCert            string             `json:"sslcert,omitempty"`
}

// GetDatabaseConnectionString ...
func GetDatabaseConnectionString(dbConnParam ConnectionParams) string {

	var connection string

	switch dbConnParam.AuthenticationType {
	case AuthenticationTypeTrust:
		connection = fmt.Sprint(
			" host=", dbConnParam.Host,
			" port=", dbConnParam.Port,
			" user=", dbConnParam.User,
			" dbname=", dbConnParam.DbName,
			" sslmode=require",
		)
	case AuthenticationTypeMD5:
		connection = fmt.Sprint(
			" host=", dbConnParam.Host,
			" port=", dbConnParam.Port,
			" user=", dbConnParam.User,
			" dbname=", dbConnParam.DbName,
			" password=", dbConnParam.Password,
			" sslmode=require",
		)
	case AuthenticationTypeCert:
		connection = fmt.Sprint(
			" host=", dbConnParam.Host,
			" port=", dbConnParam.Port,
			" user=", dbConnParam.User,
			" dbname=", dbConnParam.DbName,
			" sslmode=verify-full",
			" sslrootcert=", dbConnParam.SslRootCert,
			" sslkey=", dbConnParam.SslKey,
			" sslcert=", dbConnParam.SslCert,
		)
	default:
		fmt.Println("PostgreSQL authentication type: ", dbConnParam.AuthenticationType, ", not recognized")
	}
	fmt.Println("connection: ", connection)
	return connection
}

// Execute (Execute Command)
func Execute(connection string, timeout int, command string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", connection)
	if err != nil {
		fmt.Println("Error Openning PostgreSQL connection - ", err)
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, command)
	if err != nil {
		fmt.Println("Error executing sentence (", command, ") - ", err)
		return err
	}

	return nil
}

func ExecuteQuery(connection string, timeout int, query string) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", connection)
	if err != nil {
		fmt.Println("Error Openning PostgreSQL connection - ", err)
		return &sql.Rows{}, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query)

	if err != nil {

		return &sql.Rows{}, err
	}

	return rows, nil
}

func rowsToArray(rows *sql.Rows) ([]map[string]string, error) {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return []map[string]string{}, err
	}
	struc := make([]map[string]string, 0)
	results := make([]interface{}, len(cols))

	for i := range results {
		results[i] = new(interface{})
	}

	for rows.Next() {
		if err := rows.Scan(results[:]...); err != nil {
			return []map[string]string{}, err
		}
		row := make(map[string]string)
		for i := range results {
			val := *results[i].(*interface{})
			var str string
			if val == nil {
				str = "NULL"
			} else {
				switch v := val.(type) {
				case []byte:
					str = string(v)
				default:
					str = fmt.Sprintf("%v", v)
				}
			}
			row[cols[i]] = str
		}
		struc = append(struc, row)
	}
	return struc, nil
}

func rowsToMetricArray(rows *sql.Rows) ([]Metric, error) {
	defer rows.Close()

	struc := make([]Metric, 0)

	for rows.Next() {
		var m Metric
		if err := rows.Scan(&m.rowid, &m.node, &m.value); err != nil {
			return nil, err
		}
		struc = append(struc, m)
	}
	return struc, nil
}

type Metric struct {
	rowid int
	node  string
	value float64
}

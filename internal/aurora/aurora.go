package aurora

import (
	"database/sql"
	"fmt"

	"github.com/krotscheck/go-rds-driver"
)

type DataSource struct {
	*sql.DB
	schema string
}

type Config struct {
	ResourceArn string
	SecretArn   string
	Database    string
	AWSRegion   string
}

func New(ra, sa, dbname, r, s string) (*DataSource, error) {
	conf := &rds.Config{
		ResourceArn: ra,
		SecretArn:   sa,
		Database:    dbname,
		AWSRegion:   r,
		SplitMulti:  true,
		Custom: map[string][]string{
			"sslmode": {"disabled"},
		},
	}
	dsn := conf.ToDSN()
	db, err := sql.Open(rds.DRIVERNAME, dsn)
	if err != nil {
		return nil, err
	}
	return &DataSource{db, s}, nil
}

func (ds *DataSource) Tables() ([]string, error) {
	query := fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = '%s';", ds.schema)
	rows, err := ds.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (ds *DataSource) Preview(table string) ([][]string, error) {
	wildcard, err := ds.wildcard(table)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT 50;", wildcard, table)
	return ds.query(query)
}

func (ds *DataSource) query(query string) ([][]string, error) {
	rows, err := ds.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query: %s err: %w", query, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	table := [][]string{cols}
	for rows.Next() {
		cols := make([]string, len(cols))
		colsPtr := make([]interface{}, len(cols))
		for i := range cols {
			colsPtr[i] = &cols[i]
		}
		if err := rows.Scan(colsPtr...); err != nil {
			return nil, err
		}
		table = append(table, cols)
	}
	return table, nil
}

func (ds *DataSource) introspect(table string) ([][]string, error) {
	query := fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = '%s' AND table_name = '%s';", ds.schema, table)
	rows, err := ds.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query: %s err: %w", query, err)
	}
	defer rows.Close()

	cols := [][]string{}
	for rows.Next() {
		var col, dataType string
		if err := rows.Scan(&col, &dataType); err != nil {
			return nil, err
		}
		cols = append(cols, []string{col, dataType})
	}
	return cols, nil
}

func (ds *DataSource) wildcard(table string) (string, error) {
	cols, err := ds.introspect(table)
	if err != nil {
		return "", err
	}

	wildcard := ""

	for i, col := range cols {
		var colName string
		name := col[0]
		dtype := col[1]
		switch dtype {
		case "USER-DEFINED":
			colName = fmt.Sprintf("%s::text", name)
		case "timestamp with time zone":
			colName = fmt.Sprintf("to_char(%s, 'YYYY-MM-DD HH24:MI:SS') as %s", name, name)
		default:
			colName = name
		}

		if i == 0 {
			wildcard += colName
		} else {
			wildcard += ", " + colName
		}
	}
	return wildcard, nil
}

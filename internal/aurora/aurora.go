package aurora

import (
	"dapi/internal/tui"
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

func (ds *DataSource) Preview(table string) ([]string, [][]string, int, error) {
	wildcard, err := ds.wildcard(table)
	if err != nil {
		return nil, nil, 0, err
	}
	var count int
	err = ds.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", table)).Scan(&count)
	if err != nil {
		return nil, nil, 0, err
	}
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT 50;", wildcard, table)
	headers, data, err := ds.query(query)
	return headers, data, count, err
}

func (ds *DataSource) query(query string) ([]string, [][]string, error) {
	rows, err := ds.DB.Query(query)
	if err != nil {
		return nil, nil, fmt.Errorf("query: %s err: %w", query, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	table := [][]string{}
	for rows.Next() {
		cols := make([]string, len(cols))
		colsPtr := make([]interface{}, len(cols))
		for i := range cols {
			colsPtr[i] = &cols[i]
		}
		if err := rows.Scan(colsPtr...); err != nil {
			return nil, nil, err
		}
		table = append(table, cols)
	}
	return cols, table, nil
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

func (ds *DataSource) Filter(table string, condition string) ([]string, [][]string, error) {
	wildcard, err := ds.wildcard(table)
	if err != nil {
		return nil, nil, err
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s LIMIT 50;", wildcard, table, condition)
	return ds.query(query)
}

func (ds *DataSource) GetTable(name string, page int) (*tui.Table, error) {
	wildcard, err := ds.wildcard(name)
	if err != nil {
		return nil, err
	}
	var count int
	err = ds.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", name)).Scan(&count)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT 50 OFFSET %d;", wildcard, name, (page-1)*50)
	headers, data, err := ds.query(query)
	if err != nil {
		return nil, err
	}

	if count/50 < page {
		page = 1
	}

	return &tui.Table{
		Name:       name,
		Headers:    headers,
		Content:    data,
		TotalCount: count,
		CntPage:    page,
	}, nil
}

package tui

// Table is a struct for table data.
type Table struct {
	Name       string
	Headers    []string
	Content    [][]string
	TotalCount int
	CntPage    int
}

type DB interface {
	Tables() ([]string, error)
	GetTable(table string, page int) (*Table, error)
	Preview(table string) ([]string, [][]string, int, error)
	Filter(table string, condition string) ([]string, [][]string, error)
}

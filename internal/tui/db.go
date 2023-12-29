package tui

type DB interface {
	Tables() ([]string, error)
	Preview(table string) ([]string, [][]string, int, error)
	Filter(table string, condition string) ([]string, [][]string, error)
}

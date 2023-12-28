package tui

type DB interface {
	Tables() ([]string, error)
	Preview(table string) ([][]string, error)
}

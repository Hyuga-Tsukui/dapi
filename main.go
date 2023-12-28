package main

import (
	"dapi/internal/aurora"
	"dapi/internal/tui"
	"database/sql"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	flag "github.com/spf13/pflag"
)

var (
	resourceArn string
	secretArn   string
	database    string
	region      string
)

func main() {
	flag.StringVar(&resourceArn, "resource-arn", "", "RDS resource ARN")
	flag.StringVar(&secretArn, "secret-arn", "", "RDS secret ARN")
	flag.StringVar(&database, "database", "", "RDS database name")
	flag.StringVar(&region, "region", "ap-northeast-1", "AWS region")
	flag.Parse()

	startApp()
}

func startApp() {
	fmt.Print(resourceArn, secretArn, database, region)
	db, err := aurora.New(
		resourceArn,
		secretArn,
		database,
		region,
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()
	t := tui.New(db)
	if err := t.Run(); err != nil {
		panic(err)
	}
}

func Editor(db *sql.DB) *tview.TextArea {
	textArea := tview.NewTextArea().SetPlaceholder("Enter SQL here...")
	textArea.SetTitle("Editor").SetBorder(true)

	textArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlR {
			row, err := db.Query(textArea.GetText())
			if err != nil {
				panic(err)
			}
			defer row.Close()
		}
		return event
	})

	return textArea
}

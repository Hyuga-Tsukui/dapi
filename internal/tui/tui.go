package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUI struct {
	db         DB
	App        *tview.Application
	Tables     *tview.List
	Preview    *tview.TextView
	QueryInput *tview.InputField
	Footer     *tview.TextView
}

func (tui *TUI) queueUpdateDraw(f func()) {
	go func() {
		tui.App.QueueUpdateDraw(f)
	}()
}

func New(db DB) *TUI {
	t := &TUI{db: db}

	t.App = tview.NewApplication()

	t.Tables = tview.NewList()
	t.Tables.SetBorder(true).SetTitle("Tables")

	t.Preview = tview.NewTextView()
	t.Preview.SetTitle("Preview").SetBorder(true)

	t.QueryInput = tview.NewInputField()
	t.QueryInput.SetTitle("Query").SetBorder(true)

	t.Footer = tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Ctrl+R: Run Query")

	t.Tables.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'p':
			t.preview()
		case 'j':
			t.Tables.SetCurrentItem(t.Tables.GetCurrentItem() + 1)
		case 'k':
			t.Tables.SetCurrentItem(t.Tables.GetCurrentItem() - 1)
		}
		return event
	})

	t.Initialize()

	return t
}

func (t *TUI) Run() error {
	flex := tview.NewFlex().
		AddItem(t.Tables, 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(t.Preview, 0, 1, false).
			AddItem(t.QueryInput, 0, 1, true), 0, 4, false)

	return t.App.SetRoot(flex, true).EnableMouse(true).Run()
}

func (t *TUI) Initialize() {
	tables, err := t.db.Tables()
	if err != nil {
		panic(err)
	}

	t.queueUpdateDraw(func() {
		for _, table := range tables {
			t.Tables.AddItem(table, "", 0, nil)
		}
	})
}

func (t *TUI) preview() {
	defer func() {
		if r := recover(); r != nil {
			t.queueUpdateDraw(func() {
				t.Preview.SetText(r.(error).Error())
			})
		}
	}()

	table, _ := t.Tables.GetItemText(t.Tables.GetCurrentItem())
	preview, err := t.db.Preview(table)
	if err != nil {
		panic(err)
	}

	t.queueUpdateDraw(func() {
		t.Preview.Clear()
		if len(preview) == 1 {
			t.Preview.SetText("No rows")
			return
		}

		for _, row := range preview {
			for _, col := range row {
				t.Preview.Write([]byte(col + "\t"))
			}
			t.Preview.Write([]byte("\n"))
		}
	})
}
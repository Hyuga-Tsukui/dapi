package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUI struct {
	db         DB
	App        *tview.Application
	Tables     *tview.List
	Preview    *tview.Table
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

	t.Preview = tview.NewTable().SetBorders(true)
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

	t.initialize()

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

func (t *TUI) initialize() {
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
	table, _ := t.Tables.GetItemText(t.Tables.GetCurrentItem())
	headers, data, err := t.db.Preview(table)
	if err != nil {
		panic(err)
	}

	t.queueUpdateDraw(func() {
		t.Preview.Clear()

		for i, header := range headers {
			t.Preview.SetCell(
				0,
				i,
				tview.NewTableCell(header).
					SetTextColor(tcell.ColorYellow).
					SetAlign(tview.AlignLeft),
			)
		}

		for i, row := range data {
			for j, col := range row {
				t.Preview.SetCell(
					i+1,
					j,
					tview.NewTableCell(col).
						SetTextColor(tcell.ColorWhite).
						SetAlign(tview.AlignLeft),
				)
			}
		}
		t.App.SetFocus(t.Preview)
	})
}

package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// styles.
var (
	BgColor       = tview.Styles.PrimitiveBackgroundColor
	MenuBgColor   = tcell.ColorMediumPurple
	DialogBgColor = tcell.ColorDarkSlateGray
	DialogFgColor = tcell.ColorFloralWhite
)

type TUI struct {
	db     DB
	App    *tview.Application
	Tables *tview.List

	Preview            *tview.Table
	CurrentPreviewName string

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

	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			t.App.SetFocus(t.Tables)
		case tcell.KeyF2:
			t.App.SetFocus(t.Preview)
		}
		return event
	})

	t.Tables = tview.NewList()
	t.Tables.SetSelectedStyle(tcell.StyleDefault.Background(DialogBgColor).Foreground(DialogFgColor))
	t.Tables.SetBorder(true).SetTitle("Tables")

	t.Preview = tview.NewTable()
	t.Preview.SetBorder(true).SetTitleAlign(tview.AlignLeft).SetTitleColor(tcell.ColorYellow)
	t.Preview.SetSelectedStyle(tcell.StyleDefault.Background(DialogBgColor).Foreground(DialogFgColor))

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
			AddItem(t.Preview, 0, 1, false),
			0, 4, false)

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

		t.Preview.SetTitle(table)

		for i, header := range headers {
			t.Preview.SetCell(
				0,
				i,
				tview.NewTableCell(header).
					SetAlign(tview.AlignLeft).
					SetBackgroundColor(MenuBgColor).
					SetSelectable(false),
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
		t.Preview.SetSelectable(true, false)
		t.App.SetFocus(t.Preview)
	})
}

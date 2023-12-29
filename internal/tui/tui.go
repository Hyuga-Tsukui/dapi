package tui

import (
	"fmt"

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

	Preview            *Preview
	CurrentPreviewName string
	CurrentPreviewPage int

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
		case tcell.KeyCtrlF:
			t.App.SetFocus(t.QueryInput)
		}
		return event
	})

	t.Tables = tview.NewList()
	t.Tables.SetSelectedStyle(tcell.StyleDefault.Background(DialogBgColor).Foreground(DialogFgColor))
	t.Tables.SetBorder(true).SetTitle("Tables").SetTitleAlign(tview.AlignLeft)

	t.Preview = NewPreview()

	t.QueryInput = tview.NewInputField().SetPlaceholder("input condition here (e.g. id = 1)").SetPlaceholderStyle(tcell.StyleDefault.Foreground(tcell.ColorGray))
	t.QueryInput.SetTitle("Filter").SetBorder(true).SetTitleAlign(tview.AlignLeft)
	t.QueryInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			t.filter()
		case tcell.KeyCtrlL:
			t.QueryInput.SetText("")
			t.preview()
		}
		return event
	})

	t.Footer = tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Ctrl+R: Run Query")

	t.Tables.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'p':
			t.CurrentPreviewPage = 1
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
	outerFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	previewArea := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.QueryInput, 0, 1, false).
		AddItem(t.Preview, 0, 8, false)

	contentArea := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.Tables, 0, 1, true).
		AddItem(previewArea, 0, 4, false)

	outerFlex.
		AddItem(contentArea, 0, 1, true).
		AddItem(t.Footer, 1, 1, false)

	return t.App.SetRoot(outerFlex, true).EnableMouse(true).Run()
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
	headers, data, maxDataLen, err := t.db.Preview(table)
	if err != nil {
		panic(err)
	}

	t.queueUpdateDraw(func() {
		t.Preview.SetData(headers, data)
		// TODO: fix this.
		// Pagenationのためのカウント取得は切り出す.
		currentPage := t.CurrentPreviewPage
		t.Preview.SetTitle(fmt.Sprintf("%s %d/%d", table, currentPage, maxDataLen/50+1))
		t.Preview.SetSelectable(true)
		t.App.SetFocus(t.Preview)
	})
}

func (t *TUI) filter() {
	table, _ := t.Tables.GetItemText(t.Tables.GetCurrentItem())
	condition := t.QueryInput.GetText()

	if condition == "" {
		t.preview()
		return
	}

	headers, data, err := t.db.Filter(table, condition)
	if err != nil {
		panic(err)
	}

	t.queueUpdateDraw(func() {
		t.Preview.SetData(headers, data)
		t.Preview.SetSelectable(true)
		t.App.SetFocus(t.Preview)
	})
}

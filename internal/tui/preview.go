package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Preview struct {
	*tview.Box

	table *tview.Table
}

func NewPreview() *Preview {
	p := &Preview{
		Box: tview.NewBox().SetTitle("PREVIEW"),
	}
	p.table = tview.NewTable()
	p.table.SetBorder(true).SetTitleAlign(tview.AlignLeft).SetTitleColor(tcell.ColorYellow)
	p.table.SetSelectedStyle(tcell.StyleDefault.Background(DialogBgColor).Foreground(DialogFgColor))
	return p
}

func (p *Preview) SetData(header []string, data [][]string) {
	p.table.Clear()
	for i, h := range header {
		p.table.SetCell(0, i, tview.NewTableCell(h).SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft))
	}
	for i, row := range data {
		for j, col := range row {
			p.table.SetCell(i+1, j, tview.NewTableCell(col).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
		}
	}
}

func (p *Preview) SetSelectable(selectable bool) {
	p.table.SetSelectable(selectable, false)
}

// Draw draws this primitive onto the screen.
func (p *Preview) Draw(screen tcell.Screen) {
	p.Box.DrawForSubclass(screen, p)
	x, y, width, height := p.Box.GetInnerRect()
	p.table.SetRect(x, y, width, height)
	p.table.Draw(screen)
}

// InputHandler returns the handler for this primitive.
func (p *Preview) HasFocus() bool {
	if p.table.HasFocus() {
		return true
	}
	return p.Box.HasFocus()
}

// Focus is called by the application when the primitive receives focus.
func (p *Preview) Focus(delegate func(p tview.Primitive)) {
	p.table.SetSelectable(true, false)
	delegate(p.table)
}

// Blur is called by the application when the primitive loses focus.
func (p *Preview) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.table.InputHandler()
}

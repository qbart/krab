package main

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/rivo/tview"
)

type Pos struct {
	row int
	col int
}

type Selection struct {
	start Pos
	stop  Pos
}

// Document holds data related to a wrapped textview.
type Document struct {
	textView *tview.TextView

	cursor       Pos
	offset       Pos
	lines        []string
	maxOffsetRow int
	preferredCol int
	blinkingFlag bool
	insertMode   bool
}

// NewDocument creates a new document.
func NewDocument(textView *tview.TextView) *Document {
	return &Document{
		textView: textView,

		cursor:       Pos{1, 1},
		offset:       Pos{0, 0},
		lines:        make([]string, 0, 1000),
		maxOffsetRow: 0,

		blinkingFlag: true,
		insertMode:   false,
	}
}

// SetText sets text for textview.
func (doc *Document) SetText(text string) {
	doc.lines = strings.Split(text, "\n")
	doc.updateTextBuffer()
}

// CalculateOffset recalculates scroll offset after changes.
func (doc *Document) CalculateOffset() {
	doc.offset.row, doc.offset.col = doc.textView.GetScrollOffset()
	doc.offset.row = Max(0, doc.offset.row)
	doc.offset.col = Max(0, doc.offset.col)
}

// ClampCursor ensures cursor does not go out of possible area.
func (doc *Document) ClampCursor() {
	doc.maxOffsetRow = Max(0, len(doc.lines)-doc.visibleArea().row)
	doc.cursor.row = Clamp(doc.cursor.row, 1, len(doc.lines))
	doc.cursor.col = Clamp(doc.cursor.col, 1, len(doc.lines[doc.cursor.row-1]))
}

// VisibleLine returns selected line in textview relatievely to scroll offset.
func (doc *Document) VisibleLine() int {
	return doc.cursor.row - doc.offset.row - 1
}

// LineNumOffset returns first visible line number in textview.
func (doc *Document) LineNumOffset() int {
	return doc.offset.row + 1
}

// ShouldScrollDown returns true if document is allowed to perform scroll down.
func (doc *Document) ShouldScrollDown() bool {
	return doc.cursor.row-doc.offset.row > doc.visibleArea().row/2 && doc.offset.row < doc.maxOffsetRow
}

// ShouldScrollUp returns true if document is allowed to perform scroll up.
func (doc *Document) ShouldScrollUp() bool {
	return doc.cursor.row-doc.offset.row < doc.visibleArea().row/2 && doc.offset.row > 0
}

// ScrollDown textview.
func (doc *Document) ScrollDown() {
	doc.textView.ScrollTo(doc.offset.row+1, doc.offset.col)
	doc.CalculateOffset()
}

// ScrollUp textview.
func (doc *Document) ScrollUp() {
	doc.textView.ScrollTo(doc.offset.row-1, doc.offset.col)
	doc.CalculateOffset()
}

// GetLines returns text split by newline character.
func (doc *Document) GetLines() []string {
	return doc.lines
}

func (doc *Document) GetLineCount() int {
	return len(doc.lines)
}

// DeleteLine removes selected line.
func (doc *Document) DeleteLine() {
	if len(doc.lines) > 0 {
		i := doc.cursor.row - 1
		deletedLine := doc.lines[i]
		doc.lines = append(doc.lines[:i], doc.lines[i+1:]...)
		doc.updateTextBuffer()
		clipboard.WriteAll(deletedLine)
		doc.ClampCursor()
	}
}

// CopyLine copies selected line.
func (doc *Document) CopyLine() {
	if len(doc.lines) > 0 {
		i := doc.cursor.row - 1
		clipboard.WriteAll(doc.lines[i])
	}
}

// Paste pastes clipboard text below/above selected row.
func (doc *Document) Paste(above bool) {
	text, err := clipboard.ReadAll()

	if err == nil {
		i := doc.cursor.row - 1
		pastedLines := strings.Split(text, "\n")
		if above {
			before := doc.lines[:i]
			after := append(make([]string, 0), doc.lines[i:]...)
			doc.lines = append(before, pastedLines...)
			doc.lines = append(doc.lines, after...)
		} else {
			before := doc.lines[:i+1]
			after := append(make([]string, 0), doc.lines[i+1:]...)
			doc.lines = append(before, pastedLines...)
			doc.lines = append(doc.lines, after...)
		}
		doc.updateTextBuffer()
	}
}

func (doc *Document) StartPreciseSelection() {
}

func (doc *Document) MoveToBeginning() {
	doc.cursor.row = 1
	doc.cursor.col = 1
}

func (doc *Document) MoveToEnd() {
	doc.cursor.row = len(doc.lines)
	doc.cursor.col = 1
}

func (doc *Document) MoveDown() {
	doc.cursor.row++
	doc.cursor.col = doc.preferredCol

	doc.ClampCursor()
	doc.CalculateOffset()
}

func (doc *Document) MoveUp() {
	doc.cursor.row--
	doc.cursor.col = doc.preferredCol

	doc.ClampCursor()
	doc.CalculateOffset()
}

func (doc *Document) MoveLeft() {
	doc.cursor.col--
	doc.preferredCol = doc.cursor.col

	doc.ClampCursor()
	doc.CalculateOffset()
}

func (doc *Document) MoveRight() {
	doc.cursor.col++
	doc.preferredCol = doc.cursor.col

	doc.ClampCursor()
	doc.CalculateOffset()
}

// FindRunnableQueryRegions returns map of query groups to be run separately.
// Keys represent line numbers, values are query groups.
// Query groups start from 1. Zero means no query group.
func (doc *Document) FindRunnableQueryRegions() (result map[int]int) {
	result = make(map[int]int)
	region := 1
	findNext := false

	for t, line := range doc.lines {
		empty := len(strings.TrimSpace(line)) == 0
		if empty {
			if !findNext {
				result[t+1] = 0
				region++
				findNext = true
			}

		} else {
			result[t+1] = region
			findNext = false
		}
	}

	return
}

func (doc *Document) visibleArea() Pos {
	_, _, w, h := doc.textView.GetInnerRect()
	return Pos{w, h}
}

func (doc *Document) updateTextBuffer() {
	doc.textView.SetText(strings.Join(doc.lines, "\n"))
}

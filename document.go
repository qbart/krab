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
	lines        int
	offsetRow    int
	maxOffsetRow int
	offsetCol    int
	preferredCol int
	blinkingFlag bool
	insertMode   bool
}

// NewDocument creates a new document.
func NewDocument(textView *tview.TextView) *Document {
	return &Document{
		textView: textView,

		cursor:       Pos{1, 1},
		lines:        0,
		maxOffsetRow: 0,
		offsetRow:    0,
		offsetCol:    0,

		blinkingFlag: true,
		insertMode:   false,
	}
}

// CalculateOffset recalculates scroll offset after changes.
func (doc *Document) CalculateOffset() {
	doc.offsetRow, doc.offsetCol = doc.textView.GetScrollOffset()
	doc.offsetRow = Max(0, doc.offsetRow)
	doc.offsetCol = Max(0, doc.offsetCol)
}

// ClampCursor ensures cursor does not go out of possible area.
func (doc *Document) ClampCursor() {
	lines := doc.GetLines()
	doc.maxOffsetRow = Max(0, len(lines)-doc.lines)
	doc.cursor.row = Clamp(doc.cursor.row, 1, len(lines))
	doc.cursor.col = Clamp(doc.cursor.col, 1, len(lines[doc.cursor.row-1]))
}

// VisibleLine returns selected line in textview relatievely to scroll offset.
func (doc *Document) VisibleLine() int {
	return doc.cursor.row - doc.offsetRow - 1
}

// LineNumOffset returns first visible line number in textview.
func (doc *Document) LineNumOffset() int {
	return doc.offsetRow + 1
}

// ShouldScrollDown returns true if document is allowed to perform scroll down.
func (doc *Document) ShouldScrollDown() bool {
	return doc.cursor.row-doc.offsetRow > doc.lines/2 && doc.offsetRow < doc.maxOffsetRow
}

// ShouldScrollUp returns true if document is allowed to perform scroll up.
func (doc *Document) ShouldScrollUp() bool {
	return doc.cursor.row-doc.offsetRow < doc.lines/2 && doc.offsetRow > 0
}

// ScrollDown textview.
func (doc *Document) ScrollDown() {
	doc.textView.ScrollTo(doc.offsetRow+1, doc.offsetCol)
	doc.CalculateOffset()
}

// ScrollUp textview.
func (doc *Document) ScrollUp() {
	doc.textView.ScrollTo(doc.offsetRow-1, doc.offsetCol)
	doc.CalculateOffset()
}

// GetLines returns text split by newline character.
func (doc *Document) GetLines() []string {
	return strings.Split(doc.textView.GetText(true), "\n")
}

// DeleteLine removes line at given row.
func (doc *Document) DeleteLine(row int) {
	lines := doc.GetLines()
	if len(lines) > 0 {
		i := row - 1
		deletedLine := lines[i]
		lines = append(lines[:i], lines[i+1:]...)
		doc.textView.SetText(strings.Join(lines, "\n"))
		clipboard.WriteAll(deletedLine)
	}
}

// CopyLine copies selected line.
func (doc *Document) CopyLine() {
	lines := doc.GetLines()
	if len(lines) > 0 {
		i := doc.cursor.row - 1
		clipboard.WriteAll(lines[i])
	}
}

// Paste pastes clipboard text below/above selected row.
func (doc *Document) Paste(above bool) {
	text, err := clipboard.ReadAll()

	if err == nil {
		lines := doc.GetLines()
		i := doc.cursor.row - 1
		pastedLines := strings.Split(text, "\n")
		if above {
			before := lines[:i]
			after := append(make([]string, 0), lines[i:]...)
			lines = append(before, pastedLines...)
			lines = append(lines, after...)
		} else {
			before := lines[:i+1]
			after := append(make([]string, 0), lines[i+1:]...)
			lines = append(before, pastedLines...)
			lines = append(lines, after...)
		}
		doc.textView.SetText(strings.Join(lines, "\n"))
	}
}

func (doc *Document) StartPreciseSelection() {
}

func (doc *Document) MoveToBeginning() {
	doc.cursor.row = 1
	doc.cursor.col = 1
}

func (doc *Document) MoveToEnd() {
	doc.cursor.row = len(doc.GetLines())
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

	for t, line := range doc.GetLines() {
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

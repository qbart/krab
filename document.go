package main

import (
	"strings"

	"github.com/rivo/tview"
)

// Document holds data related to a wrapped textview.
type Document struct {
	textView *tview.TextView

	row          int
	col          int
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

		row:          1,
		col:          1,
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

// VisibleLine returns selected line in textview relatievely to scroll offset.
func (doc *Document) VisibleLine() int {
	return doc.row - doc.offsetRow - 1
}

// LineNumOffset returns first visible line number in textview.
func (doc *Document) LineNumOffset() int {
	return doc.offsetRow + 1
}

// ShouldScrollDown returns true if document is allowed to perform scroll down.
func (doc *Document) ShouldScrollDown() bool {
	return doc.row-doc.offsetRow > doc.lines/2 && doc.offsetRow < doc.maxOffsetRow
}

// ShouldScrollUp returns true if document is allowed to perform scroll up.
func (doc *Document) ShouldScrollUp() bool {
	return doc.row-doc.offsetRow < doc.lines/2 && doc.offsetRow > 0
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

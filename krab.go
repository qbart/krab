package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const corporate = `WITH regional_sales AS (
    SELECT region, SUM(amount) AS total_sales
    FROM orders
    GROUP BY region
), top_regions AS (
    SELECT region
    FROM regional_sales
    WHERE total_sales > (SELECT SUM(total_sales)/10 FROM regional_sales)
)
SELECT region,
       product,
       SUM(quantity) AS product_units,
       SUM(amount) AS product_sales
FROM orders
WHERE region IN (SELECT region FROM top_regions)
GROUP BY region, product;`

const (
	lineColSpan = 2
	colSpan     = 1
	rowSpan     = 1
)

var (
	cursorColor      = tcell.ColorRed
	bgColor          = tcell.NewRGBColor(38, 39, 47)
	fgColor          = tcell.NewRGBColor(70, 73, 90)
	highlightFgColor = tcell.NewRGBColor(255, 198, 58)
	highlightBgColor = tcell.NewRGBColor(70, 73, 90)
	footerBgColor    = fgColor //;tcell.NewRGBColor(92, 139, 154)
)

type Cursor struct {
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

type Context struct {
	db       []string
	duration time.Duration
}

const Version = "0.0.1"

func main() {
	context := Context{
		[]string{"PostgreSQL", "development"},
		time.Duration(16),
	}
	cursor := Cursor{
		row:       1,
		col:       1,
		lines:     0,
		offsetRow: 0,
		offsetCol: 0,

		blinkingFlag: true,
		insertMode:   false,
	}

	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(5, 0).
		SetBorders(false)

	editor := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false).
		SetChangedFunc(func() {
			app.Draw()
		})

	editor.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		cursor.lines = height
		visibleLine := cursor.row - cursor.offsetRow - 1

		for cx := x; cx < x+width; cx++ {
			screen.SetContent(cx, y+visibleLine, ' ', nil, tcell.StyleDefault.Background(highlightBgColor))
		}

		if cursor.blinkingFlag {
			screen.SetContent(x+cursor.col-1, y+visibleLine, ' ', nil, tcell.StyleDefault.Background(cursorColor))
		}

		return x, y, width, height
	})

	// setup cursor blinking
	go func() {
		for {
			cursor.blinkingFlag = !cursor.blinkingFlag
			duration := 300
			if cursor.blinkingFlag {
				duration = 1000
			}
			time.Sleep(time.Duration(duration) * time.Millisecond)
			app.Draw()
		}
	}()

	go func() {
		for _, line := range strings.Split(corporate, "\n") {
			words := strings.Split(line, " ")
			for i, word := range words {
				if word == ">" {
					word = "[palegreen]>[white]"
				}
				if word == ")" {
					word = "[palegreen])[white]"
				}
				if word == "(" {
					word = "[palegreen]([white]"
				}
				if word == "SELECT" {
					word = "[palegreen]SELECT[white]"
				}
				if word == "FROM" {
					word = "[palegreen]FROM[white]"
				}
				if word == "AS" {
					word = "[palegreen]AS[white]"
				}
				if word == "WHERE" {
					word = "[palegreen]WHERE[white]"
				}
				if word == "WITH" {
					word = "[palegreen]WITH[white]"
				}
				if word == "AND" {
					word = "[palegreen]AND[white]"
				}
				if word == "IN" {
					word = "[palegreen]IN[white]"
				}
				if word == "IN" {
					word = "[palegreen]IN[white]"
				}
				if word == "BY" {
					word = "[palegreen]BY[white]"
				}
				if word == "GROUP" {
					word = "[palegreen]GROUP[white]"
				}
				if i == len(words)-1 {
					fmt.Fprintf(editor, "%s", word)
				} else {
					fmt.Fprintf(editor, "%s ", word)
				}
			}
			fmt.Fprintf(editor, "\n")
		}
	}()

	editor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !cursor.insertMode {
			switch event.Key() {
			case tcell.KeyRune:
				switch event.Rune() {
				case 'i':
					cursor.insertMode = true
				}
			}
		}

		lines := strings.Split(editor.GetText(true), "\n")

		switch event.Key() {
		case tcell.KeyEscape:
			cursor.insertMode = false

		case tcell.KeyDown:
			cursor.row++
			cursor.col = cursor.preferredCol

		case tcell.KeyUp:
			cursor.row--
			cursor.col = cursor.preferredCol

		case tcell.KeyLeft:
			cursor.col--
			cursor.preferredCol = cursor.col

		case tcell.KeyRight:
			cursor.col++
			cursor.preferredCol = cursor.col
		}
		cursor.maxOffsetRow = len(lines) - cursor.lines
		cursor.row = Clamp(cursor.row, 1, len(lines))
		cursor.col = Clamp(cursor.col, 1, len(lines[cursor.row-1]))
		cursor.offsetRow, cursor.offsetCol = editor.GetScrollOffset()

		if cursor.row-cursor.offsetRow > cursor.lines/2 && cursor.offsetRow < cursor.maxOffsetRow {
			switch event.Key() {
			case tcell.KeyDown:
				editor.ScrollTo(cursor.offsetRow+1, cursor.offsetCol)
			}
		} else if cursor.row-cursor.offsetRow < cursor.lines/2 && cursor.offsetRow > 0 {
			switch event.Key() {
			case tcell.KeyUp:
				editor.ScrollTo(cursor.offsetRow-1, cursor.offsetCol)
			}
		}
		cursor.offsetRow, cursor.offsetCol = editor.GetScrollOffset()

		// c := lines[cursor.row-1][cursor.col-1]

		return nil
	})

	lineNumbers := tview.NewBox()

	lineNumbers.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		visibleLine := cursor.row - cursor.offsetRow - 1
		line := cursor.offsetRow + 1
		for cy := y; cy < y+height; cy++ {
			s := fmt.Sprintf("%3d  ", line)
			d := len(s)
			runes := []rune(s)
			selected := visibleLine == cy

			for i := 0; i < d; i++ {
				screen.SetContent(x+i, cy, runes[i], nil, tcell.StyleDefault.
					Foreground(ColorIf(selected, ColorIf(cursor.insertMode, cursorColor, highlightFgColor), fgColor)).
					Background(ColorIf(selected, highlightBgColor, bgColor)))
			}
			screen.SetContent(x+4, cy, '│', nil, tcell.StyleDefault.Foreground(fgColor).Background(bgColor))
			line++
		}

		return x, y, 5, height
	})

	footer := tview.NewBox().
		SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
			for cx := x; cx < x+width; cx++ {
				screen.SetContent(cx, y+height-1, ' ', nil, tcell.StyleDefault.
					Background(ColorIf(cursor.insertMode, cursorColor, footerBgColor)))
			}

			offset := 0
			for index, text := range context.db {
				db := []rune(text)
				for i := 0; i < len(text); i++ {
					screen.SetContent(offset+x+1, y+height-1, db[i], nil, tcell.StyleDefault.
						Background(ColorIf(cursor.insertMode, cursorColor, footerBgColor)))
					offset++
				}
				if index != len(context.db)-1 {
					offset++
					screen.SetContent(offset+x+1, y+height-1, '►', nil, tcell.StyleDefault.
						Background(ColorIf(cursor.insertMode, cursorColor, footerBgColor)))
					offset += 2
				}
			}

			time := []rune(fmt.Sprintf("%d ms | [%d,%d]", context.duration, cursor.row, cursor.col))
			timeX := x + width - len(time)

			for i := 0; i < len(time); i++ {
				screen.SetContent(timeX+i-1, y+height-1, time[i], nil, tcell.StyleDefault.
					Background(ColorIf(cursor.insertMode, cursorColor, footerBgColor)))
			}

			return x, y, width, 1
		})

	grid.
		AddItem(editor, 0, 1, rowSpan, colSpan, 0, 0, true).
		AddItem(lineNumbers, 0, 0, rowSpan, colSpan, 0, 0, false).
		AddItem(footer, 1, 0, rowSpan, lineColSpan, 0, 0, false)

	if err := app.SetRoot(grid, true).SetFocus(editor).Run(); err != nil {
		panic(err)
	}
}

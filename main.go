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

func colorIf(flag bool, a, b tcell.Color) tcell.Color {
	if flag {
		return a
	} else {
		return b
	}
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Clamp(v, a, b int) int {
	return Max(a, Min(v, b))
}

type Cursor struct {
	row          int
	col          int
	blinkingFlag bool
	insertMode   bool
}

func main() {
	cursor := Cursor{
		row:          1,
		col:          1,
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
		for cx := x; cx < x+width; cx++ {
			screen.SetContent(cx, y+cursor.row-1, ' ', nil, tcell.StyleDefault.Background(highlightBgColor))
		}

		if cursor.blinkingFlag {
			screen.SetContent(x+cursor.col-1, y+cursor.row-1, ' ', nil, tcell.StyleDefault.Background(cursorColor))
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
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'i':
				cursor.insertMode = true
			}
		case tcell.KeyEscape:
			cursor.insertMode = false
		case tcell.KeyDown:
			cursor.row++
		case tcell.KeyUp:
			cursor.row--
		case tcell.KeyLeft:
			cursor.col--
		case tcell.KeyRight:
			cursor.col++
		}
		s := editor.GetText(true)
		lines := strings.Split(s, "\n")
		cursor.row = Clamp(cursor.row, 1, len(lines))
		cursor.col = Clamp(cursor.col, 1, len(lines[cursor.row-1]))
		// c := lines[cursor.row-1][cursor.col-1]

		return event
	})

	lineNumbers := tview.NewBox().
		SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
			line := 1
			for cy := y; cy < y+height; cy++ {
				// lineStr := fmt.Sprintf(`%s`, line)
				s := fmt.Sprintf("%3d  ", line)
				d := len(s)
				runes := []rune(s)
				selected := cursor.row-1 == cy

				for i := 0; i < d; i++ {
					screen.SetContent(x+i, cy, runes[i], nil, tcell.StyleDefault.
						Foreground(colorIf(selected, colorIf(cursor.insertMode, cursorColor, highlightFgColor), fgColor)).
						Background(colorIf(selected, highlightBgColor, bgColor)))
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
					Background(colorIf(cursor.insertMode, cursorColor, footerBgColor)))
			}

			db := []rune(fmt.Sprintf("development"))
			for i := 0; i < len(db); i++ {
				screen.SetContent(x+i+1, y+height-1, db[i], nil, tcell.StyleDefault.
					Background(colorIf(cursor.insertMode, cursorColor, footerBgColor)))
			}

			time := []rune(fmt.Sprintf("16.6 ms"))
			timeX := x + width - len(time)

			for i := 0; i < len(time); i++ {
				screen.SetContent(timeX+i-1, y+height-1, time[i], nil, tcell.StyleDefault.
					Background(colorIf(cursor.insertMode, cursorColor, footerBgColor)))
			}

			return x, y, width, 1
		})

	// editor.Highlight(strconv.Itoa(cursor.row))

	grid.
		// SetBorders(true).
		AddItem(editor, 0, 1, rowSpan, colSpan, 0, 0, false).
		AddItem(lineNumbers, 0, 0, rowSpan, colSpan, 0, 0, false).
		AddItem(footer, 1, 0, rowSpan, lineColSpan, 0, 0, false)

	// AddItem(p Primitive, row, column, rowSpan, colSpan, minGridHeight, minGridWidth int, focus bool) *Grid

	if err := app.SetRoot(grid, true).SetFocus(editor).Run(); err != nil {
		panic(err)
	}
}

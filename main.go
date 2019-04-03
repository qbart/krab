package main

import (
	"fmt"
	"strconv"
	"strings"

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

const LINE_COL_SPAN = 2
const COL_SPAN = 1
const ROW_SPAN = 1

func main() {
	BG_COLOR := tcell.NewRGBColor(38, 39, 47)
	FG_COLOR := tcell.NewRGBColor(70, 73, 90)

	HIGHLIGHT_FG_COLOR := tcell.NewRGBColor(255, 198, 58)
	HIGHLIGHT_BG_COLOR := tcell.NewRGBColor(70, 73, 90)
	BG_FOOTER_COLOR := FG_COLOR //;tcell.NewRGBColor(92, 139, 154)

	currentLine := 15

	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(5, 0).
		SetBorders(false)

	editor := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	editor.SetBackgroundColor(BG_COLOR)
	numSelections := 0
	go func() {
		for _, line := range strings.Split(corporate, "\n") {
			for _, word := range strings.Split(line, " ") {
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
				// if word == "to" {
				// 	word = fmt.Sprintf(`["%d"]to[""]`, numSelections)
				// 	numSelections++
				// }
				fmt.Fprintf(editor, "%s ", word)
			}
			fmt.Fprintf(editor, "\n")
		}
	}()

	editor.SetDoneFunc(func(key tcell.Key) {
		currentSelection := editor.GetHighlights()
		if key == tcell.KeyEnter {
			if len(currentSelection) > 0 {
				editor.Highlight()
			} else {
				editor.Highlight("0").ScrollToHighlight()
			}
		} else if len(currentSelection) > 0 {
			index, _ := strconv.Atoi(currentSelection[0])
			if key == tcell.KeyTab {
				index = (index + 1) % numSelections
			} else if key == tcell.KeyBacktab {
				index = (index - 1 + numSelections) % numSelections
			} else {
				return
			}
			editor.Highlight(strconv.Itoa(index)).ScrollToHighlight()
		}
	})

	lineNumbers := tview.NewBox().
		SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
			line := 1
			for cy := y; cy < y+height; cy++ {
				// lineStr := fmt.Sprintf(`%s`, line)
				s := fmt.Sprintf("%3d  ", line)
				d := len(s)
				runes := []rune(s)
				selected_color := FG_COLOR
				selected_bg_color := BG_COLOR
				if currentLine == cy+1 {
					selected_color = HIGHLIGHT_FG_COLOR
					selected_bg_color = HIGHLIGHT_BG_COLOR
				}
				for i := 0; i < d; i++ {
					screen.SetContent(x+i, cy, runes[i], nil, tcell.StyleDefault.Foreground(selected_color).Background(selected_bg_color))
				}
				// screen.SetContent(x+4, cy, '│', nil, tcell.StyleDefault.Foreground(FG_COLOR).Background(BG_COLOR))
				line++
			}

			return x, y, 5, height
		})

	footer := tview.NewBox().
		SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
			for cx := x; cx < x+width; cx++ {
				screen.SetContent(cx, y+height-1, ' ', nil, tcell.StyleDefault.Background(BG_FOOTER_COLOR))
			}

			db := []rune(fmt.Sprintf("development"))
			for i := 0; i < len(db); i++ {
				screen.SetContent(x+i+1, y+height-1, db[i], nil, tcell.StyleDefault.Background(BG_FOOTER_COLOR))
			}

			time := []rune(fmt.Sprintf("16.6 ms"))
			timeX := x + width - len(time)

			for i := 0; i < len(time); i++ {
				screen.SetContent(timeX+i-1, y+height-1, time[i], nil, tcell.StyleDefault.Background(BG_FOOTER_COLOR))
			}

			return x, y, width, 1
		})

	grid.
		// SetBorders(true).
		AddItem(editor, 0, 1, ROW_SPAN, COL_SPAN, 0, 0, false).
		AddItem(lineNumbers, 0, 0, ROW_SPAN, COL_SPAN, 0, 0, false).
		AddItem(footer, 1, 0, ROW_SPAN, LINE_COL_SPAN, 0, 0, false)

	// AddItem(p Primitive, row, column, rowSpan, colSpan, minGridHeight, minGridWidth int, focus bool) *Grid

	if err := app.SetRoot(grid, true).SetFocus(editor).Run(); err != nil {
		panic(err)
	}
}

package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const example = `WITH regional_sales AS (
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

type Context struct {
	db       []string
	duration time.Duration
}

const Version = "0.0.1"

func main() {
	styles := NewTheme()
	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(5, 0).
		SetBorders(false)

	editor := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false).
		SetWrap(false).
		SetChangedFunc(func() {
			app.Draw()
		})

	lineNumbers := tview.NewBox()

	context := Context{
		[]string{"development"},
		time.Duration(16),
	}

	doc := NewDocument(editor)

	editor.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		doc.lines = height

		for cx := x; cx < x+width; cx++ {
			screen.SetContent(cx, y+doc.VisibleLine(), ' ', nil, tcell.StyleDefault.Background(styles.highlightBgColor))
		}

		if doc.blinkingFlag {
			screen.SetContent(x+doc.col-1, y+doc.VisibleLine(), ' ', nil, tcell.StyleDefault.Background(styles.cursorColor))
		}

		return x, y, width, height
	})

	// setup doc blinking
	go func() {
		for {
			doc.blinkingFlag = !doc.blinkingFlag
			duration := 300
			if doc.blinkingFlag {
				duration = 1000
			}
			time.Sleep(time.Duration(duration) * time.Millisecond)
			app.Draw()
		}
	}()

	editor.SetText(example)
	// editor.SetText("")

	pressedKeys := ""

	editor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if doc.insertMode {

		} else {
			if event.Key() == tcell.KeyRune {
				key := event.Rune()
				pressedKeys += string(key)

				switch pressedKeys {
				case "i":
					doc.insertMode = true
					pressedKeys = ""
				case "d":
				case "y":
				case "dd":
					doc.DeleteLine(doc.row)
					pressedKeys = ""
				case "yy":
					doc.CopyLine(doc.row)
					pressedKeys = ""
				default:
					pressedKeys = ""
				}
			}
		}

		switch event.Key() {
		case tcell.KeyEscape:
			doc.insertMode = false

		case tcell.KeyDown:
			doc.row++
			doc.col = doc.preferredCol

		case tcell.KeyUp:
			doc.row--
			doc.col = doc.preferredCol

		case tcell.KeyLeft:
			doc.col--
			doc.preferredCol = doc.col

		case tcell.KeyRight:
			doc.col++
			doc.preferredCol = doc.col
		}
		doc.ClampCursor()
		doc.CalculateOffset()

		if doc.ShouldScrollDown() {
			switch event.Key() {
			case tcell.KeyDown:
				doc.ScrollDown()
			}
		} else if doc.ShouldScrollUp() {
			switch event.Key() {
			case tcell.KeyUp:
				doc.ScrollUp()
			}
		}

		// c := lines[doc.row-1][doc.col-1]

		return nil
	})

	lineNumbers.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		runnableRegions := doc.FindRunnableQueryRegions()
		lineNum := doc.LineNumOffset()

		for cy := y; cy < y+height; cy++ {
			s := []rune(fmt.Sprintf("%3d  ", lineNum))
			selected := doc.VisibleLine() == cy

			for i := 0; i < len(s); i++ {
				screen.SetContent(x+i, cy, s[i], nil, tcell.StyleDefault.
					Foreground(ColorIf(
						doc.insertMode && selected,
						styles.cursorColor,
						ColorIf(
							runnableRegions[lineNum] != 0,
							styles.RunnableRegionColorByIndex(runnableRegions[lineNum]),
							ColorIf(selected,
								styles.highlightFgColor,
								styles.fgColor)))).
					Background(ColorIf(selected, styles.highlightBgColor, styles.bgColor)))
			}
			screen.SetContent(x+4, cy, '│', nil, tcell.StyleDefault.
				Foreground(styles.RunnableRegionColorByIndex(runnableRegions[lineNum])).
				Background(styles.bgColor))
			lineNum++
		}

		return x, y, 5, height
	})

	footer := tview.NewBox().
		SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
			for cx := x; cx < x+width; cx++ {
				screen.SetContent(cx, y+height-1, ' ', nil, tcell.StyleDefault.
					Background(ColorIf(doc.insertMode, styles.cursorColor, styles.footerBgColor)))
			}

			offset := 0
			for index, text := range context.db {
				db := []rune(text)
				for i := 0; i < len(text); i++ {
					screen.SetContent(offset+x+1, y+height-1, db[i], nil, tcell.StyleDefault.
						Background(ColorIf(doc.insertMode, styles.cursorColor, styles.footerBgColor)))
					offset++
				}
				if index != len(context.db)-1 {
					offset++
					screen.SetContent(offset+x+1, y+height-1, '►', nil, tcell.StyleDefault.
						Background(ColorIf(doc.insertMode, styles.cursorColor, styles.footerBgColor)))
					offset += 2
				}
			}
			runnableRegions := doc.FindRunnableQueryRegions()
			if runnableRegions[doc.row] != 0 {
				offset += 2
				screen.SetContent(offset, y+height-1, '■', nil, tcell.StyleDefault.
					Foreground(styles.RunnableRegionColorByIndex(runnableRegions[doc.row])).
					Background(ColorIf(doc.insertMode, styles.cursorColor, styles.footerBgColor)))
			}

			time := []rune(fmt.Sprintf("%d ms | [%d,%d]", context.duration, doc.row, doc.col))
			timeX := x + width - len(time)

			for i := 0; i < len(time); i++ {
				screen.SetContent(timeX+i-1, y+height-1, time[i], nil, tcell.StyleDefault.
					Background(ColorIf(doc.insertMode, styles.cursorColor, styles.footerBgColor)))
			}

			keysChars := []rune(pressedKeys)
			keysCharsX := x + width/2 - len(keysChars)

			for i := 0; i < len(keysChars); i++ {
				screen.SetContent(keysCharsX+i-1, y+height-1, keysChars[i], nil, tcell.StyleDefault.
					Background(ColorIf(doc.insertMode, styles.cursorColor, styles.footerBgColor)).
					Foreground(styles.highlightFgColor))
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

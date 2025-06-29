package ui

import (
    "fmt"
    "sort"

    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

// RunTimeline displays timeline groups using tview list + textview. keys: arrow, enter, q
func RunTimeline(groups map[string][]string) error {
    app := tview.NewApplication()

    // sort categories
    cats := make([]string, 0, len(groups))
    for k := range groups {
        cats = append(cats, k)
    }
    sort.Strings(cats)

    list := tview.NewList().ShowSecondaryText(false)
    for _, c := range cats {
        list.AddItem(fmt.Sprintf("%s (%d)", c, len(groups[c])), "", 0, nil)
    }
    list.AddItem("All", "", 0, nil)

    text := tview.NewTextView().SetDynamicColors(true).SetWrap(false)

    flex := tview.NewFlex().AddItem(list, 30, 1, true).AddItem(text, 0, 3, false)
    border := tview.NewFrame(flex).SetBorders(0, 0, 0, 0, 1, 1)

    update := func(idx int) {
        text.Clear()
        var lines []string
        if idx == len(cats) { // All
            for _, c := range cats {
                lines = append(lines, fmt.Sprintf("[yellow]=== %s ===[-]", c))
                lines = append(lines, groups[c]...)
                lines = append(lines, "")
            }
        } else {
            lines = groups[cats[idx]]
        }
        for _, l := range lines {
            fmt.Fprintln(text, l)
        }
        app.Draw()
    }
    update(0)

    list.SetChangedFunc(func(i int, _ string, _ string, _ rune) {
        update(i)
    })
    list.SetSelectedFunc(func(i int, _ string, _ string, _ rune) {
        update(i)
    })

    app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
        if e.Key() == tcell.KeyRune && (e.Rune() == 'q' || e.Rune() == 'Q') {
            app.Stop()
            return nil
        }
        return e
    })

    if err := app.SetRoot(border, true).Run(); err != nil {
        return err
    }
    return nil
}

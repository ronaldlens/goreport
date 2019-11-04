package main

import "github.com/rivo/tview"

var pages *tview.Pages

func RunGui() {
	app := tview.NewApplication()
	buildGui(app)
	err := app.Run()
	if err != nil {
		panic(err)
	}
}

func buildGui(app *tview.Application) {
	box := tview.NewBox().
		SetBorder(true).
		SetTitle("GoReport")
	app.SetRoot(box, true)
	buildMainMenu(box, app)
}

func buildMainMenu(box *tview.Box, app *tview.Application) {
	list := tview.NewList().ShowSecondaryText(false)
	list.AddItem("Import dataset", "", 'i', nil)
	list.AddItem("Run report", "", 'r', nil)

	stopFunc := func() {
		app.Stop()
	}
	list.AddItem("Quit", "", 'q', stopFunc)
	//box.
}

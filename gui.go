package main

import (
	ui "github.com/VladimirMarkelov/clui"
)

// RunGui starts the Gui
func RunGui(incidents Incidents) {
	ui.InitLibrary()
	defer ui.DeinitLibrary()

}

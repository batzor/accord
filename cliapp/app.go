package cliapp

import (
	"log"

	"github.com/jroimartin/gocui"
)

// Start launches client application
func Start() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatalln(err)
	}
	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, send); err != nil {
		log.Fatalln(err)
	}
	if err := g.SetKeybinding("password", gocui.KeyEnter, gocui.ModNone, login); err != nil {
		log.Fatalln(err)
	}
	if err := g.SetKeybinding("username", gocui.KeyEnter, gocui.ModNone, enterUsername); err != nil {
		log.Fatalln(err)
	}
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

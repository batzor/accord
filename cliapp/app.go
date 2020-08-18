package cliapp

import (
	"log"

	"github.com/jroimartin/gocui"
	"github.com/qvntm/Accord/client"
	"github.com/qvntm/Accord/pb"
)

type ClientApp struct {
	client       client.AccordClient
	streamClient pb.Chat_StreamClient
}

func NewClientApp() *ClientApp {
	return &ClientApp{
		client: *client.NewAccordClient(),
	}
}

// Start launches client application
func (app *ClientApp) Start() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatalln(err)
	}
	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, app.send); err != nil {
		log.Fatalln(err)
	}
	if err := g.SetKeybinding("password", gocui.KeyEnter, gocui.ModNone, app.login); err != nil {
		log.Fatalln(err)
	}
	if err := g.SetKeybinding("username", gocui.KeyEnter, gocui.ModNone, enterUsername); err != nil {
		log.Fatalln(err)
	}
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

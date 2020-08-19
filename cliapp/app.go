package cliapp

import (
	"log"

	"github.com/jroimartin/gocui"
	"github.com/qvntm/Accord/client"
)

type StreamCommunication struct {
	reqComm client.StreamRequestCommunication
	resComm client.StreamResponseCommunication
}

type ClientApp struct {
	client           client.AccordClient
	currentChannelID uint64
	// Currently, there will only be one channel-to-stream
	// mapping at a time, because cliapp will only stream
	// with one channel at a time, which is currentChannel.
	// However, if there is need to stream with multiple
	// channels simultaneously, then more keys can be added.
	channelsToStreams map[uint64]StreamCommunication
}

func NewClientApp() *ClientApp {
	serverID := uint64(67890) // dummy value so far
	return &ClientApp{
		client: *client.NewAccordClient(serverID),
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

package client

import "github.com/jroimartin/gocui"

// Layout creates the terminal interface of the client app
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	g.Cursor = true

	if messages, err := g.SetView("messages", 0, 0, maxX-30, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		messages.Title = " messages: "
		messages.Autoscroll = true
		messages.Wrap = true
	}

	if input, err := g.SetView("input", 0, maxY-5, maxX-30, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		input.Title = " send: "
		input.Autoscroll = false
		input.Wrap = true
		input.Editable = true
	}

	if chats, err := g.SetView("chats", maxX-30, 0, maxX-1, maxY-40); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		chats.Title = " chats: "
		chats.Autoscroll = false
		chats.Wrap = true
	}

	if users, err := g.SetView("users", maxX-30, maxY-40, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		users.Title = " users: "
		users.Autoscroll = false
		users.Wrap = true
	}

	if name, err := g.SetView("name", maxX/2-10, maxY/2-1, maxX/2+10, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetCurrentView("name")
		name.Title = " name: "
		name.Autoscroll = false
		name.Wrap = true
		name.Editable = true
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

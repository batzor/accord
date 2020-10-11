package cliapp

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

	if channels, err := g.SetView("channels", maxX-30, 0, maxX-1, maxY-15); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		channels.Title = " channels: "
		channels.Autoscroll = false
		channels.Wrap = true
	}

	if users, err := g.SetView("users", maxX-30, maxY-15, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		users.Title = " users: "
		users.Autoscroll = false
		users.Wrap = true
	}

	if password, err := g.SetView("password", maxX/2-15, maxY/2-1, maxX/2+15, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetViewOnBottom("password")
		password.Title = " Please, enter password: "
		password.Autoscroll = false
		password.Wrap = true
		password.Editable = true
	}

	if username, err := g.SetView("username", maxX/2-15, maxY/2-1, maxX/2+15, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetViewOnTop("username")
		g.SetCurrentView("username")
		username.Title = " Please, enter username: "
		username.Autoscroll = false
		username.Wrap = true
		username.Editable = true
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

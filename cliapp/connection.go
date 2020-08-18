package cliapp

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

var (
	// tempUsername stores the username before the user has logged in
	tempUsername string
)

func enterUsername(g *gocui.Gui, v *gocui.View) error {
	tempUsername = v.Buffer()
	if tempUsername == "" {
		// do nothing until user enters smth and presses "Enter" again
		return nil
	}
	v.Clear()
	g.SetViewOnTop("password")
	g.SetCurrentView("password")
	return nil
}

// Connect connects to the server and does some initialization
func (app *ClientApp) login(g *gocui.Gui, v *gocui.View) error {
	password := v.Buffer()
	if password == "" {
		// do nothing until user enters smth and presses "Enter" again
		return nil
	}
	v.Clear()

	if err := app.client.Connect("0.0.0.0:50051"); err != nil {
		return fmt.Errorf("couldn't connect to server: %v", err)
	}

	/*c := pb.NewChatClient(cc)
	loginRequest := &pb.LoginRequest{
		Password: password,
		Username: username,
	}
	loginResponse, err := pb.Login(context.Background(), loginRequest)
	if err != nil {
		status := status.FromError(err)
		if status.GetCode() == codes.InvalidArgument {
			// TODO: make some pop-up message saying "Incorrect login or password. Please, try again."
			g.SetViewOnTop("username")
			g.SetCurrentView("password")
			return nil
		} else {
			log.Fatalf("Login request has failed: %v", err)
		}
	}*/

	g.SetViewOnTop("messages")
	g.SetViewOnTop("input")
	g.SetViewOnTop("channels")
	g.SetViewOnTop("users")
	g.SetCurrentView("input")

	currentChannel = "baztmr" // ideally, we have to fetch an ID of the last channel user has used.

	channels := []string{"baztmr", "tmrbaz", "temirtau"} // ideally, we have to fetch a list of all channels where user belongs.
	channelsView, _ := g.View("channels")
	channelsView.Title = fmt.Sprintf(" %d channels: ", len(channels))
	for _, channel := range channels {
		fmt.Fprintln(channelsView, channel)
	}

	users := []string{"Temir", "Bazo"} // fetch it somehow
	usersView, _ := g.View("users")
	usersView.Title = fmt.Sprintf(" %d users in %s: ", len(users), currentChannel)
	for _, user := range users {
		fmt.Fprintln(usersView, user)
	}

	//messagesView, _ := g.View("messages")
	/*stream, err := c.Stream(context.Background())
	if err != nil {
		log.Fatalf("Couldn't obtain message stream: %v", err)
	}
	streamClient = stream*/

	// message receiving gorouitne
	/*go func() {
		for {
			data, _ := reader.ReadString('\n')
			msg := strings.TrimSpace(data)
			switch {
			case strings.HasPrefix(msg, "/clients>"):
				data := strings.SplitAfter(msg, ">")[1]
				clientsSlice := strings.Split(data, " ")
				clientsCount := len(clientsSlice)
				var clients string
				for _, client := range clientsSlice {
					clients += client + "\n"
				}
				g.Update(func(g *gocui.Gui) error {
					usersView.Title = fmt.Sprintf(" %d users: ", clientsCount)
					usersView.Clear()
					fmt.Fprintln(usersView, clients)
					return nil
				})
			default:
				g.Update(func(g *gocui.Gui) error {
					fmt.Fprintln(messagesView, msg)
					return nil
				})
			}
		}
	}()*/
	return nil
}

func (app *ClientApp) send(g *gocui.Gui, v *gocui.View) error {
	messagesView, _ := g.View("messages")
	v.Clear()
	fmt.Fprintln(messagesView, "Dummy message has been sent")
	return nil
}

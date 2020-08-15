package main

import (
	"github.com/qvntm/Accord/server"
)

func main() {
	server.NewAccordServer().Start("0.0.0.0:50051")
}

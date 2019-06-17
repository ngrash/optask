package main

import (
	"nicograshoff.de/x/optask/config"
	"nicograshoff.de/x/optask/server"
)

func main() {
	project := config.ReadConfig("config.json")
	server := server.NewServer(":8080", project)
	server.ListenAndServe()
}

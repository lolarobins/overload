package main

import (
	"math/rand"
	"os"
	"time"

	"lolarobins.ca/overload/fetch"
	"lolarobins.ca/overload/input"
	"lolarobins.ca/overload/log"
	"lolarobins.ca/overload/node"
	"lolarobins.ca/overload/settings"
	"lolarobins.ca/overload/webserver"
)

func shutdown() {
	log.Info("Stopping")

	if err := webserver.Stop(); err != nil {
		log.Error("Stopping web server: " + err.Error())
	}

	webserver.ShutdownLock.Lock()

	log.Info("Killing remaining active nodes")
	node.KillAll()
}

func mkdirReq(names ...string) bool {
	for _, name := range names {
		if _, err := os.ReadDir(name); os.IsNotExist(err) {
			log.Info("Creating directory '" + name + "'")

			if err := os.Mkdir(name, 0777); err != nil {
				log.Error("Failed to create required directory '" + name + "': " + err.Error())
				return false
			}
		}
	}

	return true
}

func main() {
	log.Info("Starting overload v0.1.1")

	// uwu
	rand.Seed(time.Now().UTC().UnixNano())

	if !mkdirReq("config", "jar", "nodes") {
		return
	}

	input.Init() // input

	if err := settings.Init(); err != nil { // settings
		log.Error("Intializing settings: " + err.Error())
	}

	if err := node.Init(); err != nil { // nodes
		log.Error("Intializing nodes: " + err.Error())
	}

	if err := webserver.Init(); err != nil { // webserver
		log.Error("Intializing web server: " + err.Error())
	}

	fetch.Init() // fetch jar util

	log.Info("Startup finished")

	input.AcceptInput()

	shutdown()
}

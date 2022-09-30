package webserver

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"lolarobins.ca/overload/log"
	"lolarobins.ca/overload/node"
	"lolarobins.ca/overload/settings"
)

var srv *http.Server
var ShutdownLock = new(sync.Mutex)

func Init() error {
	srv = &http.Server{Addr: settings.Settings.Hostname + ":" + settings.Settings.PanelPort}
	panel := http.FileServer(http.Dir("web"))

	// panel itself
	http.Handle("/", panel)

	// api
	// TODO

	// port forward
	port, _ := strconv.Atoi(settings.Settings.PanelPort)

	ip := settings.Settings.Hostname
	if settings.Settings.UPnP && settings.Settings.PanelPortForward {
		if inuse, _ := node.Router.IsForwardedTCP(uint16(port)); inuse {
			log.Info("Port " + settings.Settings.PanelPort + " is already forwared and may overlap with another public port")
		}

		node.Router.Clear(uint16(port))
		if err := node.Router.Forward(uint16(port), "overload web panel"); err != nil {
			return err
		} else {
			ip, _ = node.Router.ExternalIP()
		}
	}

	log.Info("Web panel and API visible on http://" + ip + ":" + settings.Settings.PanelPort + "/")

	ShutdownLock.Lock()

	// serve page and api
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("Error in webserver goroutine: " + err.Error())
		}

		log.Info("Web server has stopped")

		// clear port forward
		if settings.Settings.UPnP && settings.Settings.PanelPortForward {
			node.Router.Clear(uint16(port))
		}

		ShutdownLock.Unlock()
	}()

	return nil
}

func Stop() error {
	return srv.Shutdown(context.Background())
}

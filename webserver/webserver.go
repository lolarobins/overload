package webserver

import (
	"net/http"

	"lolarobins.ca/overload/log"
	"lolarobins.ca/overload/node"
	"lolarobins.ca/overload/settings"
)

func Init() error {
	// init api endpoint
	panel := http.FileServer(http.Dir("web"))

	// panel itself
	http.Handle("/", panel)

	// api
	// TODO

	// port forward
	ip := settings.Settings.Hostname
	if settings.Settings.UPnP && settings.Settings.PanelPortForward {
		if err := node.Router.Forward(8080, "overload web panel", "tcp"); err != nil {
			return err
		} else {
			ip, _ = node.Router.ExternalIP()
		}
	}

	log.Info("Web panel and API visible on http://" + ip + ":" + settings.Settings.PanelPort + "/")

	// serve page and api
	go func() {

		if err := http.ListenAndServe(settings.Settings.Hostname+":"+settings.Settings.PanelPort, nil); err != http.ErrServerClosed {
			log.Error("Error in webserver goroutine: " + err.Error())
		}

		// clear port forward
		if settings.Settings.UPnP && settings.Settings.PanelPortForward {
			node.Router.Clear(8080, "tcp")
		}
	}()

	return nil
}

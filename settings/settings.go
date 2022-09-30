package settings

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"

	"gitlab.com/NebulousLabs/go-upnp"
	"lolarobins.ca/overload/log"
)

type ServerSettings struct {
	UPnP             bool   `json:"upnp"`
	Hostname         string `json:"hostname"`
	PanelPort        string `json:"panelport"`
	PanelPortForward bool   `json:"panelportforward"`
	Router           string `json:"router"`
}

var Settings = ServerSettings{
	Hostname:         getOutboundIP().String(),
	PanelPort:        "8080",
	PanelPortForward: true,
}

// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func getOutboundIP() net.IP {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func Init() error {
	data, err := os.ReadFile("config/settings.json")

	if err != nil {
		router, err := upnp.DiscoverCtx(context.Background())

		// save upnp router to save startup speed
		if err == nil {
			Settings.Router = router.Location()
			Settings.UPnP = true
			log.Info("UPnP router found, enabling UPnP")
		} else {
			Settings.UPnP = false
			log.Info("No UPnP router found")
		}

		data, err := json.MarshalIndent(Settings, "", "    ")
		if err != nil {
			return errors.New("fatal: unexpected JSON error")
		}

		if err := os.WriteFile("config/settings.json", data, 0777); err != nil {
			return errors.New("fatal: unable to read/write in working directory")
		}
	} else if err := json.Unmarshal(data, &Settings); err != nil {
		return errors.New("fatal: 'config/settings.json' cannot be parsed")
	}

	return nil
}

func (s *ServerSettings) Save() error {
	data, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return errors.New("error marshalling JSON to output to file")
	}

	if err := os.WriteFile("config/settings.json", data, 0777); err != nil {
		return errors.New("could not write to file 'config/settings.json'")
	}

	return nil
}

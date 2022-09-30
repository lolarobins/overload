package settings

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
)

type ServerSettings struct {
	UPnP             bool   `json:"upnp"`
	Hostname         string `json:"hostname"`
	PanelPort        string `json:"panelport"`
	PanelPortForward bool   `json:"panelportforward"`
}

var Settings = ServerSettings{
	UPnP:             true,
	Hostname:         getOutboundIP().String(),
	PanelPort:        "8080",
	PanelPortForward: true,
}

// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func Init() error {
	data, err := os.ReadFile("config/settings.json")

	if err != nil {
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

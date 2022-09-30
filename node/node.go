package node

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"gitlab.com/NebulousLabs/go-upnp"
	"lolarobins.ca/overload/input"
	"lolarobins.ca/overload/log"
	"lolarobins.ca/overload/settings"
)

type NodeConfig struct {
	Name        string `json:"name"`
	Jar         string `json:"jar"`
	JVM         string `json:"jvm"`
	Port        string `json:"port"`
	Memory      uint16 `json:"memory"`
	Autostart   bool   `json:"autostart"`
	PortForward bool   `json:"portforward"`
}

type Node struct {
	Id      string
	Config  NodeConfig
	active  bool
	Monitor bool
	cmd     *exec.Cmd
	writer  *io.WriteCloser
}

var Nodes = make(map[string]*Node)
var Router *upnp.IGD
var WaitGroup = new(sync.WaitGroup)
var extIp string

var DefaultNode = NodeConfig{
	Name:        "Minecraft Server",
	JVM:         "java",
	Port:        "[N/A IN DEFAULT CONFIG]",
	Memory:      1024,
	Autostart:   false,
	PortForward: true,
}

func Init() error {
	if settings.Settings.UPnP {
		var err = error(nil)
		if Router, err = upnp.Load(settings.Settings.Router); err != nil {
			if Router, err = upnp.DiscoverCtx(context.Background()); err == nil {
				settings.Settings.Router = Router.Location()
				log.Info("Updated UPnP router")
			} else {
				settings.Settings.UPnP = false
				log.Info("Could not locate UPnP router, disabling UPnP")
			}
			settings.Settings.Save()
		}

		extIp, _ = Router.ExternalIP()

		if err == nil {
			log.Info("UPnP port forwarding on external host " + extIp)
		}
	} else {
		log.Info("UPnP disabled")
	}

	// constructs the node default configs
	data, err := os.ReadFile("config/default-node.json")
	if err != nil {

		data, err := json.MarshalIndent(DefaultNode, "", "    ")
		if err != nil {
			return errors.New("fatal: unexpected JSON error")
		}

		if err := os.WriteFile("config/default-node.json", data, 0777); err != nil {
			return errors.New("fatal: unable to read/write in working directory")
		}
	} else if err := json.Unmarshal(data, &DefaultNode); err != nil {
		return errors.New("fatal: 'config/default-node.json' cannot be parsed")
	}

	// loads the nodes
	if info, err := os.Stat("nodes"); !os.IsNotExist(err) && info.IsDir() {
		files, err := os.ReadDir("nodes")
		if err != nil {
			return errors.New("fatal: could not read 'nodes' directory")
		}

		for _, f := range files {
			// ignore dotfiles, non-directories
			if strings.HasPrefix(f.Name(), ".") || !f.IsDir() {
				continue
			}

			if _, err := Load(f.Name()); err != nil {
				log.Error("'" + f.Name() + "' could not be loaded: " + err.Error())
			}
		}
	} else if !os.IsNotExist(err) {
		return errors.New("fatal: 'nodes' exists as a file, preventing creation of directory")
	} else if err := os.Mkdir("nodes", 0777); err != nil {
		return errors.New("fatal: 'nodes' directory could not be created")
	}

	// node commands !
	input.Command{
		Function: func(s []string) {
			log.Info("Showing nodes:")

			for _, node := range Nodes {
				log.Info(node.Config.Name + " (" + node.Id + ") > Port: " + node.Config.Port + ", Memory: " + strconv.Itoa(int(node.Config.Memory)) + " Nodes: " + strconv.FormatBool(node.active))
			}
		},
		Command:     "nodes",
		Description: "View a list of nodes loaded in memory",
	}.Register()

	input.Command{
		Function: func(s []string) {
			if len(s) != 2 {
				log.Error("Invalid arguments")
				return
			}

			node, err := Create(s[1])

			if err != nil {
				log.Error("Error creating node: " + err.Error())
				return
			}

			log.Info("Created node " + node.Id + " (Port: " + node.Config.Port + ")")
		},
		Command:     "create",
		Args:        " <id>",
		Description: "Create a node given an ID",
	}.Register()

	input.Command{
		Function: func(s []string) {
			if len(s) != 2 {
				log.Error("Invalid arguments")
				return
			}

			if s[1] == "*" {
				log.Info("Starting all nodes")
				for _, n := range Nodes {
					n.Start()
				}
				return
			}

			node, err := Get(s[1])

			if err != nil {
				log.Error("Error starting node: " + err.Error())
				return
			}

			if err := node.Start(); err != nil {
				log.Error("Error starting node: " + err.Error())
			}
		},
		Command:     "start",
		Args:        " <id/*>",
		Description: "Start a node",
	}.Register()

	input.Command{
		Function: func(s []string) {
			if len(s) < 3 {
				log.Error("Invalid arguments")
				return
			}

			node, err := Get(s[1])

			if err != nil {
				log.Error("Error sending command: " + err.Error())
				return
			}

			msg := ""
			for i := 2; i < len(s); i++ {
				msg += s[i] + " "
			}
			msg = strings.TrimSpace(msg)

			if err := node.SendCommand(msg); err != nil {
				log.Error("Error sending command: " + err.Error())
			}
		},
		Command:     "send",
		Args:        " <id>",
		Description: "Send a command to a node",
	}.Register()

	input.Command{
		Function: func(s []string) {
			if len(s) != 2 {
				log.Error("Invalid arguments")
				return
			}

			if s[1] == "*" {
				log.Info("Sending stop command to all nodes")
				StopAll()
				return
			}

			node, err := Get(s[1])

			if err != nil {
				log.Error("Error stopping node: " + err.Error())
				return
			}

			if err := node.SendCommand("stop"); err != nil {
				log.Error("Error stopping node: " + err.Error())
				return
			}

			log.Info("Sending stop command to " + node.Config.Name + " (" + node.Id + ")")
		},
		Command:     "stop",
		Args:        " <id/*>",
		Description: "Send stop command to a node",
	}.Register()

	input.Command{
		Function: func(s []string) {
			if len(s) != 2 {
				log.Error("Invalid arguments")
				return
			}

			if s[1] == "*" {
				log.Info("Killing all nodes")
				KillAll()
				return
			}

			node, err := Get(s[1])

			if err != nil {
				log.Error("Error killing node: " + err.Error())
				return
			}

			if err := node.Kill(); err != nil {
				log.Error("Error killing node: " + err.Error())
				return
			}

			log.Info("Sending kill command to " + node.Config.Name + " (" + node.Id + ")")
		},
		Command:     "kill",
		Args:        " <id/*>",
		Description: "Send kill command to node",
	}.Register()

	input.Command{
		Function: func(s []string) {
			if len(s) != 2 {
				log.Error("Invalid arguments")
				return
			}

			if s[1] == "*" {
				log.Info("Monitoring all nodes")
				for _, n := range Nodes {
					n.Monitor = true
				}
				return
			}

			node, err := Get(s[1])

			if err != nil {
				log.Error("Error monitoring node: " + err.Error())
				return
			}

			if node.Monitor {
				log.Info("No longer monitoring " + node.Config.Name + " (" + node.Id + ")")
				node.Monitor = false
			} else {
				log.Info("Monitoring " + node.Config.Name + " (" + node.Id + ")")
				node.Monitor = true
			}
		},
		Command:     "monitor",
		Args:        " <id/*>",
		Description: "Monitor the output of a node while it is active",
	}.Register()

	input.Command{
		Function: func(s []string) {
			if len(s) != 2 {
				log.Error("Invalid arguments")
				return
			}

			if s[1] == "*" {
				log.Info("Accepted EULA for all nodes")
				for _, n := range Nodes {
					n.AcceptEULA()
				}
				return
			}

			node, err := Get(s[1])

			if err != nil {
				log.Error("Error accepting EULA: " + err.Error())
				return
			}

			if err := node.AcceptEULA(); err != nil {
				log.Error("Error accepting EULA: " + err.Error())
			}

			log.Info("Accepted EULA for " + node.Config.Name + " (" + node.Id + ")")
		},
		Command:     "eula",
		Args:        " <id/*>",
		Description: "Monitor the output of a node while it is active",
	}.Register()

	input.Command{
		Function: func(s []string) {
			if len(s) == 2 {
				node, err := Get(s[1])

				if err != nil {
					log.Error("Error getting node config: " + err.Error())
					return
				}

				log.Info("Showing configuration defined in nodes/" + node.Id + "/node.json:")
				log.Info("name: " + node.Config.Name)
				log.Info("port: " + node.Config.Port)
				log.Info("jar: " + node.Config.Jar)
				log.Info("jvm: " + node.Config.JVM)
				log.Info("memory (mb): " + strconv.Itoa(int(node.Config.Memory)))
				log.Info("autostart: " + strconv.FormatBool(node.Config.Autostart))
				log.Info("portforward: " + strconv.FormatBool(node.Config.PortForward))

				return
			}

			if len(s) < 4 {
				log.Error("Invalid arguments")
				return
			}

			node, err := Get(s[1])

			if err != nil && s[1] != "*" {
				log.Error("Error changing node config: " + err.Error())
				return
			}

			// build value as one str
			val := ""
			for i := 3; i < len(s); i++ {
				val += s[i] + " "
			}
			val = strings.TrimSpace(val)

			if s[1] == "*" {
				log.Info("Set " + strings.ToLower(s[2]) + " to " + val + " for all nodes")
				for _, n := range Nodes {
					n.SetConfig(strings.ToLower(s[2]), val)
				}
				return
			}

			if err := node.SetConfig(strings.ToLower(s[2]), val); err != nil {
				log.Error("Error changing node config: " + err.Error())
				return
			}

			log.Info("Set " + strings.ToLower(s[2]) + " to " + val + " for " + node.Config.Name + " (" + node.Id + ")")
		},
		Command:     "config",
		Args:        " <id/*> [key] [value]",
		Description: "View the current or change a value in the nodes configuration",
	}.Register()

	// returns the nils
	return nil
}

func Load(id string) (*Node, error) {
	if n, ok := Nodes[id]; ok {
		return n, nil
	}

	n := new(Node)

	n.Id = id
	n.active = false
	n.Config = DefaultNode

	data, err := os.ReadFile("nodes/" + id + "/node.json")
	if err != nil {
		return nil, errors.New("'nodes/" + id + "/node.json' does not exist")
	}

	if err := json.Unmarshal(data, &n.Config); err != nil {
		return nil, errors.New("'nodes/" + id + "/node.json' cannot be parsed")
	}

	n.SaveConfig()

	if n.Config.Autostart {
		n.Start()
	}

	Nodes[id] = n

	return n, nil
}

func Get(id string) (*Node, error) {
	n, ok := Nodes[id]
	if !ok {
		return nil, errors.New("node '" + id + "' does not exist or is not loaded into memory")
	}

	return n, nil
}

func Create(id string) (*Node, error) {
	if node, _ := Get(id); node != nil {
		return nil, errors.New("node '" + id + "' already exists")
	}

	node := Node{
		Id:     id,
		Config: DefaultNode,
		active: false,
	}

	node.Config.Port = strconv.Itoa(rand.Intn(45000-10000) + 10000)

	if err := node.SaveConfig(); err != nil {
		return nil, node.SaveConfig()
	}

	Nodes[id] = &node

	return &node, nil
}

func StopAll() {
	for _, n := range Nodes {
		n.SendCommand("stop")
	}
}

func KillAll() {
	for _, n := range Nodes {
		n.Kill()
	}
}

func (n *Node) SaveConfig() error {
	data, err := json.MarshalIndent(n.Config, "", "    ")
	if err != nil {
		return errors.New("error marshalling JSON to output to file")
	}

	if _, err := os.ReadDir("nodes/" + n.Id); os.IsNotExist(err) {
		os.Mkdir("nodes/"+n.Id, 0777)
	}

	if err := os.WriteFile("nodes/"+n.Id+"/node.json", data, 0777); err != nil {
		return errors.New("could not write to file 'nodes/" + n.Id + "/node.json'")
	}

	return nil
}

func (n *Node) Start() error {
	if n.active {
		return errors.New("node already started")
	}

	n.active = true

	ip := settings.Settings.Hostname

	port, _ := strconv.Atoi(n.Config.Port)
	if n.Config.PortForward && settings.Settings.UPnP {
		if inuse, _ := Router.IsForwardedTCP(uint16(port)); inuse {
			log.Info("Port " + n.Config.Port + " is already forwared and may overlap with another public port")
		}

		Router.Clear(uint16(port))
		if err := Router.Forward(uint16(port), "overload port forwarding"); err != nil {
			log.Error("Failed to port forward " + n.Config.Name + ": " + err.Error())
		} else {
			ip = extIp
		}
	}

	log.Info("Starting " + n.Config.Name + " (" + n.Id + ") on " + ip + ":" + n.Config.Port)

	n.cmd = exec.Command(n.Config.JVM, "-Xmx"+strconv.Itoa(int(n.Config.Memory))+"M", "-jar", "../../jar/"+n.Config.Jar, "--host", settings.Settings.Hostname, "--port", n.Config.Port, "--nogui")
	n.cmd.Dir = "nodes/" + n.Id

	reader, _ := n.cmd.StdoutPipe()
	writer, _ := n.cmd.StdinPipe()
	n.writer = &writer

	scanner := bufio.NewScanner(reader)

	WaitGroup.Add(1)

	go func() {
		for scanner.Scan() {
			if n.Monitor {
				println(n.Id + " > " + scanner.Text())
			}
		}

		n.active = false
		log.Info("Stopped " + n.Config.Name + " (" + n.Id + ")")

		if n.Config.PortForward && settings.Settings.UPnP {
			Router.Clear(uint16(port))
		}

		WaitGroup.Done()
	}()

	n.cmd.Start()

	return nil
}

func (n *Node) SendCommand(command string) error {
	if !n.active {
		return errors.New("node is not currently active")
	}

	io.WriteString(*n.writer, command+"\n")

	return nil
}

func (n *Node) Kill() error {
	if !n.active {
		return errors.New("node is not currently active")
	}

	n.active = false
	return n.cmd.Process.Kill()
}

func (n *Node) IsRunning() bool {
	return n.active
}

func (n *Node) SetConfig(key string, val string) error {
	switch key {
	case "name":
		n.Config.Name = val
	case "port":
		_, err := strconv.Atoi(val)

		if err != nil {
			return errors.New("invalid integer value")
		}

		n.Config.Port = val
	case "jar":
		n.Config.Jar = val
	case "jvm":
		n.Config.JVM = val
	case "memory":
		valint, err := strconv.Atoi(val)

		if err != nil {
			return errors.New("invalid integer value")
		}

		n.Config.Memory = uint16(valint)
	case "autostart":
		valbool := true

		switch strings.ToLower(val) {
		case "true", "on", "yes":
		case "false", "off", "no":
			valbool = false
		default:
			return errors.New("invalid boolean value")
		}

		n.Config.Autostart = valbool

	case "portforward":
		valbool := true

		switch strings.ToLower(val) {
		case "true", "on", "yes":
		case "false", "off", "no":
			valbool = false
		default:
			return errors.New("invalid boolean value")
		}

		n.Config.PortForward = valbool
	default:
		return errors.New("configuration key not found")
	}

	return n.SaveConfig()
}

func (n *Node) AcceptEULA() error {
	eula := []byte("eula=true")
	return os.WriteFile("nodes/"+n.Id+"/eula.txt", eula, 0777)
}

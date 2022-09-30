package fetch

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"lolarobins.ca/overload/input"
	"lolarobins.ca/overload/log"
)

func Init() {
	input.Command{
		Function: func(s []string) {
			if len(s) < 3 || len(s) > 5 {
				log.Error("Invalid arguments")
				return
			}

			switch strings.ToLower(s[1]) {
			case "paper":
				log.Info("Starting a goroutine to fetch server jar for PaperMC version '" + s[2] + "'")
				go func() {
					if err := FetchPaper(s[2]); err != nil {
						log.Error("Fetching paper: " + err.Error())
					}
				}()
			default:
				log.Error("Fetching server jar: implementation not found")
			}
		},
		Command:     "fetch",
		Args:        " <implementation> <version>",
		Description: "Fetch server jars for common server implementations such as Spigot, Paper, Bungee, etc",
	}.Register()
}

type paperErr struct {
	Err string `json:"error"`
}

type paperVersions struct {
	Versions []string `json:"versions"`
}

type paperBuilds struct {
	Builds []int `json:"builds"`
}

func FetchPaper(version string) error {
	// fetch paper script
	if strings.ToLower(version) == "latest" {
		resp, err := http.Get("https://api.papermc.io/v2/projects/paper")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		versions := &paperVersions{}
		json.Unmarshal(body, versions)

		version = versions.Versions[len(versions.Versions)-1]
	}

	resp, err := http.Get("https://api.papermc.io/v2/projects/paper/versions/" + version)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 404:
		err := &paperErr{}
		json.Unmarshal(body, err)

		return errors.New(err.Err)
	case 200:
	default:
		return errors.New("unhandled http response")
	}

	builds := &paperBuilds{}
	json.Unmarshal(body, builds)

	build := builds.Builds[len(builds.Builds)-1]

	out, err := os.Create("jar/paper-" + version + ".jar")
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err = http.Get("https://api.papermc.io/v2/projects/paper/versions/" + version + "/builds/" + strconv.Itoa(build) + "/downloads/paper-" + version + "-" + strconv.Itoa(build) + ".jar")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("could not fetch download file")
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Info("Done fetching PaperMC version '" + version + "' (Build: " + strconv.Itoa(build) + ")")

	return nil
}

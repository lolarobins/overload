# overload
## Introduction
overload is an open-source project looking to help server administrators manage multiple Minecraft servers in the most efficient manner possible through a command-line interface and web panel

overload is intended for hosting server networks on home & cloud servers with easy installation and setup: simply build the program, and edit the configuration files and you're ready to go!

## Features
Current Features:
- Automatic Paper & Waterfall latest version fetching
- UPnP Port-Forwarding for servers on networks that support it for easy port-forwarding
- Auto accept EULA
- Command-line interface for creating and managing nodes

**TODO:**
- Forge, Spigot, QuiltMC, Fabric, BungeeCord fetching/building
- Web API & Panel
- Integrations plugin to get stats about players, etc
- Plugin package manager

## Installation
Building overload is fairly simple, just open a terminal and enter each of the commands listed below. Make sure that you have a go compiler and git installed on your machine. In order to verify that they are installed, you can run `go version` and `git version` in your command-line.

1. Download the repository (`git clone https://github.com/lolarobins/overflow`)
2. Enter the repository's directory (macOS/Linux: `cd overload`, Windows: `chdir overload`)
3. Build overload (`go build`)
4. Launch overload (macOS/Linux: `./overload`, Windows: `overload.exe`)
5. Optionally, you may delete the source folders, or copy the executable to another directory to keep things cleaner.

## Setup
Setting up a node from a fresh install can get you a playable server with the following command examples:
- `create test` > Create a node named test
- `fetch paper latest` > Fetch latest version of PaperMC
- `config test jar paper-1.19.2.jar` > Configure test node to use downloaded jar
- `eula test` > Accept EULA for test
- `monitor test` > View console output for test
- `start test` > Start the server
And done! You now have a playable Minecraft server open to the world (if portforwarded using UPnP)!

## Contribution
Any contribution to overload would be greatly appreciated. If you have any features you'd like to see, or if you want to make changes and refactor code where it's beneficial, open a pull request :3

### Contributers:
- [Lola Robins](https://github.com/lolarobins)


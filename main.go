package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	cmdline "github.com/galdor/go-cmdline"
	ini "github.com/vaughan0/go-ini"
)

const VERSION = "0.0.0"

var (
	config   ini.File
	exitChan chan struct{}
)

func reqHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	// Handle command line options/arguments (-h/--help is implicit)
	cl := cmdline.New()
	cl.AddOption("c", "conf", "file", "path to the config file")
	cl.SetOptionDefault("conf", "/etc/DroneExternalConfig/config.ini")
	cl.Parse(os.Args)

	// Attempt to parse config file
	configPath := cl.OptionValue("conf")
	configFile, err := ini.LoadFile(configPath)
	if err != nil {
		log.Fatalf("I couldn't parse the config file because of an error: %s", err.Error())
	}
	config = configFile

	// Check and get the important sections of the config file
	serverConfig, serverConfigExists := config["server"]
	if !serverConfigExists {
		log.Fatalf("The config file seems to be missing the [server] section. Please add it and re-start DroneExternalConfig.")
	}
	_, mappingsConfigExists := config["config-map"]
	if !mappingsConfigExists {
		log.Fatalf("The config file seems to be missing the [config-map] section. Please add it and re-start DroneExternalConfig.")
	}

	// Log startup info
	log.Printf("Starting DroneExternalConfig version %s", VERSION)

	// Prepare server variables
	serverLAddr := "0.0.0.0"
	serverLPort := "80"

	if setting, exists := serverConfig["listen-addr"]; exists {
		serverLAddr = setting
	}
	if setting, exists := serverConfig["listen-port"]; exists {
		serverLPort = setting
	}

	// Set up and start the server
	http.HandleFunc("/", reqHandler)

	log.Printf("Starting HTTP listener on %s:%s", serverLAddr, serverLPort)
	serverErr := http.ListenAndServe(
		fmt.Sprintf("%s:%s",
			serverLAddr,
			serverLPort,
		),
		nil,
	)
	if serverErr != nil {
		log.Fatalf("The server failed to start because of the following error: %s", serverErr.Error())
	} else {
		<-exitChan
	}
}

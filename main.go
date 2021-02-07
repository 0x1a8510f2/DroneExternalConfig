package main

import (
	"log"
	"os"

	cmdline "github.com/galdor/go-cmdline"
	ini "github.com/vaughan0/go-ini"
)

func main() {
	// Handle command line options/arguments (-h/--help is implicit)
	cl := cmdline.New()
	cl.AddOption("c", "conf", "file", "path to the config file")
	cl.SetOptionDefault("conf", "/etc/DroneExternalConfig/config.ini")
	cl.Parse(os.Args)

	// Attempt to parse config file
	configPath := cl.OptionValue("conf")
	config, err := ini.LoadFile(configPath)
	if err != nil {
		log.Fatalf("I couldn't parse the config file because of an error: %s", err.Error())
	}

	// Get the important sections of the config file
	serverConfig, serverConfigExists := config["server"]
	if !serverConfigExists {
		log.Fatalf("The config file seems to be missing the [server] section. Please add it and re-start DroneExternalConfig.")
	}
	mappingsConfig, mappingsConfigExists := config["config-map"]
	if !mappingsConfigExists {
		log.Fatalf("The config file seems to be missing the [config-map] section. Please add it and re-start DroneExternalConfig.")
	}

	// Start the server
	// TODO
}

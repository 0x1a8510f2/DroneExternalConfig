/* Copyright (C) 2021 TR_SLimey - All Rights Reserved
 * You may use, distribute and modify this code under the
 * terms of the Apache 2.0 license, which can be found in
 * the LICENSE file.
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	cmdline "github.com/galdor/go-cmdline"
	ini "github.com/vaughan0/go-ini"
)

const VERSION = "1.0.2"

var (
	config   ini.File
	exitChan chan struct{}
)

type Repo struct {
	Id             int64
	Uid            string
	User_id        int64
	Namespace      string
	Name           string
	Slug           string
	Scm            string
	Git_http_url   string
	Git_ssh_url    string
	Link           string
	Default_branch string
	Private        bool
	Visibility     string
	Active         bool
	Config         string
	Trusted        bool
	Protected      bool
	Ignore_forks   bool
	Ignore_pulls   bool
	Cancel_pulls   bool
	Timeout        int64
	Counter        int64
	Synced         int64
	Created        int64
	Updated        int64
	Version        int64
}

type Build struct {
	Id            int64
	Repo_id       int64
	Number        int64
	Parent        int64
	Status        string
	Error         string
	Event         string
	Action        string
	Link          string
	Timestamp     int64
	Title         string
	Message       string
	Before        string
	After         string
	Ref           string
	Source_repo   string
	Source        string
	Target        string
	Author_login  string
	Author_name   string
	Author_email  string
	Author_avatar string
	Sender        string
	Params        map[string]string
	Cron          string
	Deploy_to     string
	Deploy_id     int64
	Started       int64
	Finished      int64
	Created       int64
	Updated       int64
	Version       int64
}

type droneRequest struct {
	Repo  Repo
	Build Build
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
	// Keep track of status code and log message
	respStatusCode := http.StatusNoContent
	respMsg := ""

	// Log request when done processing
	defer func() {
		log.Printf("[%d] HTTP %s %s from %s (%s)", respStatusCode, r.Method, r.RequestURI, r.RemoteAddr, respMsg)
	}()

	// Only accept POST requests
	if r.Method != http.MethodPost {
		// Unless...
		if r.Method == "BREW" {
			respMsg = "I am a teapot"
			respStatusCode = http.StatusTeapot
			w.WriteHeader(respStatusCode)
			return
		}
		respMsg = fmt.Sprintf("rejected request due to invalid method: %s", r.Method)
		respStatusCode = http.StatusMethodNotAllowed
		w.WriteHeader(respStatusCode)
		return
	}

	// Read request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respMsg = fmt.Sprintf("error while reading body of request: %s", err.Error())
		respStatusCode = http.StatusInternalServerError
		w.WriteHeader(respStatusCode)
		return
	}

	// Try to unmarshal request body
	data := droneRequest{}
	err = json.Unmarshal(reqBody, &data)
	if err != nil {
		respMsg = fmt.Sprintf("error while unmarshalling request data: %s", err.Error())
		respStatusCode = http.StatusBadRequest
		w.WriteHeader(respStatusCode)
		return
	}

	// Get the full name of the repo and check against our map
	repoName := data.Repo.Slug
	if configLocation, exists := config["config-map"][repoName]; exists {
		// Parse the URL pointing at the config
		configLocationParsed, err := url.Parse(configLocation)
		if err != nil {
			// URL is invalid so fallback on standard config but log the error
			respMsg = fmt.Sprintf("unable parse URL for repo %s (%s) due to error (%s) so falling back to config in repo", repoName, configLocation, err.Error())
			respStatusCode = http.StatusNoContent
			w.WriteHeader(respStatusCode)
			return
		}

		// Attempt to fetch the config from its location
		var response []byte
		if configLocationParsed.Scheme == "http" || configLocationParsed.Scheme == "https" {
			result, err := http.Get(configLocation)
			if err != nil {
				// We couldn't fetch the config so fall back to file in repo
				respMsg = fmt.Sprintf("unable to retrieve file at URL for repo %s (%s) due to error (%s) so falling back to config in repo", repoName, configLocation, err.Error())
				respStatusCode = http.StatusNoContent
				w.WriteHeader(respStatusCode)
				return
			}
			body, err := ioutil.ReadAll(result.Body)
			result.Body.Close()
			if err != nil {
				// We couldn't fetch the config so fall back to file in repo
				respMsg = fmt.Sprintf("unable to retrieve file at URL for repo %s (%s) due to error (%s) so falling back to config in repo", repoName, configLocation, err.Error())
				respStatusCode = http.StatusNoContent
				w.WriteHeader(respStatusCode)
				return
			}
			response = body
		} else if configLocationParsed.Scheme == "file" {
			data, err := ioutil.ReadFile(configLocationParsed.Path)
			if err != nil {
				// We couldn't fetch the config so fall back to file in repo
				respMsg = fmt.Sprintf("unable to retrieve file at URL for repo %s (%s) due to error (%s) so falling back to config in repo", repoName, configLocation, err.Error())
				respStatusCode = http.StatusNoContent
				w.WriteHeader(respStatusCode)
				return
			}
			response = data
		} else {
			// The scheme is unsupported so fallback
			respMsg = fmt.Sprintf("the URL scheme for repo %s (%s) in unsupported so falling back to config in repo", repoName, configLocationParsed.Scheme)
			respStatusCode = http.StatusNoContent
			w.WriteHeader(respStatusCode)
			return
		}

		// Wrap the config in JSON
		responseDict := map[string]string{"data": string(response)}
		finaldata, err := json.Marshal(responseDict)
		if err != nil {
			respMsg = fmt.Sprintf("failed to wrap config in JSON due to error (%s)", err.Error())
			respStatusCode = http.StatusInternalServerError
			w.WriteHeader(respStatusCode)
			return
		}

		// Return the resulting config to requester
		respStatusCode = http.StatusOK
		w.WriteHeader(respStatusCode)
		_, err = w.Write(finaldata)
		if err != nil {
			respMsg = fmt.Sprintf("failed to send config back to requester due to error (%s)", err.Error())
		} else {
			respMsg = fmt.Sprintf("successfully sent config for repo %s (%s)", repoName, configLocation)
		}
	} else {
		// Otherwise, return a 204 to fallback to a local config for the repo
		respMsg = fmt.Sprintf("no matching config found for repo %s", repoName)
		respStatusCode = http.StatusNoContent
		w.WriteHeader(respStatusCode)
		return
	}
}

func main() {
	// Handle command line options/arguments (-h/--help is implicit)
	// TODO: Allow overriding config within commandline
	// TODO: --version
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
	serverTLSCert := ""
	serverTLSKey := ""

	if setting, exists := serverConfig["listen-addr"]; exists {
		serverLAddr = setting
	}
	if setting, exists := serverConfig["listen-port"]; exists {
		serverLPort = setting
	}
	if settingTLSCert, exists := serverConfig["tls-cert"]; exists {
		if settingTLSKey, exists := serverConfig["tls-key"]; exists {
			serverTLSCert, serverTLSKey = settingTLSCert, settingTLSKey
		}
	}

	// Set up and start the server
	http.HandleFunc("/", reqHandler)

	log.Printf("Starting HTTP listener on %s:%s", serverLAddr, serverLPort)
	var serverErr error
	if serverTLSCert == "" {
		serverErr = http.ListenAndServe(
			fmt.Sprintf("%s:%s",
				serverLAddr,
				serverLPort,
			),
			nil,
		)
	} else {
		serverErr = http.ListenAndServeTLS(
			fmt.Sprintf("%s:%s",
				serverLAddr,
				serverLPort,
			),
			serverTLSCert,
			serverTLSKey,
			nil,
		)
	}
	if serverErr != nil {
		log.Fatalf("The server failed to start because of the following error: %s", serverErr.Error())
	} else {
		<-exitChan
	}
}

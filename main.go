package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type Repo struct {
	Id             int64
	Uid            int64
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
	// Keep track of status code
	respStatusCode := http.StatusNoContent
	respMsg := ""

	// Log request when done processing
	defer func() {
		log.Printf("[%d] HTTP %s %s from %s (%s)", respStatusCode, r.Method, r.RequestURI, r.RemoteAddr, respMsg)
	}()

	// Only accept POST requests
	if r.Method != http.MethodPost {
		respMsg = fmt.Sprintf("Rejected request due to invalid method: %s", r.Method)
		respStatusCode = http.StatusMethodNotAllowed
		w.WriteHeader(respStatusCode)
		return
	}

	// Read request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respMsg = fmt.Sprintf("Error while reading body of request: %s", err.Error())
		respStatusCode = http.StatusInternalServerError
		w.WriteHeader(respStatusCode)
		return
	}

	// Try to unmarshal reqeust body
	data := droneRequest{}
	err = json.Unmarshal(reqBody, &data)
	if err != nil {
		respMsg = fmt.Sprintf("Error while unmarshalling request data: %s", err.Error())
		respStatusCode = http.StatusBadRequest
		w.WriteHeader(respStatusCode)
		return
	}

	fmt.Printf("%v\n", data)
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

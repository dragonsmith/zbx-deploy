package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func helloHandler(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "http://evl.ms", http.StatusFound)
}

func startHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	req.ParseForm()
	projectName := vars["project"]

	if projects[projectName] == nil {
		return
	}

	token := req.Header.Get("X-Auth-Token")
	if token != config.Projects[projectName].Token {
		http.Error(w, "invalid token", 401)
		return
	}

	if projects[projectName].MaintenanceId > 0 {
		fmt.Fprintf(w, "already deploying: %d", projects[projectName].MaintenanceId)
		return
	}

	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	deployer := req.FormValue("deployer")
	time := time.Now().Format(layout)
	inserted, err := CreateMaintenance(fmt.Sprintf("deploy: %s @ %s", projectName, time), fmt.Sprintf("deployed by %s", deployer), config.DeployDuration, projects[projectName].GroupIds)
	if err != nil {
		log.Fatalf("Failed to create: %v", err)
	}

	fmt.Printf("[%s] creating maintenance #%d\n", projectName, inserted)
	projects[projectName].MaintenanceId = inserted

	fmt.Fprintf(w, "maintenance %d created\n", inserted)
}

func finishHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	projectName := vars["project"]

	if projects[projectName] == nil {
		return
	}

	token := req.Header.Get("X-Auth-Token")
	if token != config.Projects[projectName].Token {
		http.Error(w, "invalid token", 401)
		return
	}

	remove := projects[projectName].MaintenanceId

	if remove > 0 {
		fmt.Fprintf(w, "removing maintenance %d\n", remove)
		fmt.Printf("[%s] deleting maintenance #%d\n", projectName, remove)
		DeleteMaintenance(remove)

		projects[projectName].MaintenanceId = 0
	}
}

type projectStatus struct {
	GroupIds      []int
	MaintenanceId int64
}

var bindAddress string

var projects map[string]*projectStatus

func init() {
	flag.StringVar(&bindAddress, "bind", "127.0.0.1:3001", "address to bind")
}

func CleanupMaintenances() {
	for key, value := range projects {
		if value.MaintenanceId > 0 {
			fmt.Printf("[%s] cleaning up maintenance #%d\n", key, value.MaintenanceId)
			DeleteMaintenance(value.MaintenanceId)
		}
	}
}

func prepareProjects() map[string]*projectStatus {
	projects = make(map[string]*projectStatus, 0)

	for k, v := range config.Projects {
		projects[k] = &projectStatus{v.Hosts, 0}
	}

	return projects
}

func main() {
	parseConfig()
	projects = prepareProjects()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("captured %v, cleaning up and exiting..", sig)
			CleanupMaintenances()
			os.Exit(1)
		}
	}()

	LoginZabbix(config.Zabbix.Endpoint, config.Zabbix.Username, config.Zabbix.Password)

	r := mux.NewRouter()
	r.HandleFunc("/", helloHandler)
	r.HandleFunc("/start/{project}", startHandler)
	r.HandleFunc("/finish/{project}", finishHandler)

	http.Handle("/", r)

	fmt.Printf("Starting API on %s\n", bindAddress)

	err := http.ListenAndServe(bindAddress, r)
	if err != nil {
		log.Fatalf("Server start error: %v", err)
	}
}

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

	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	deployer := req.FormValue("deployer")
	time := time.Now().Format(layout)

	duration := config.DeployDuration
	if duration == 0 {
		// can't be zero
		duration = 600
	}

	// We need to prolongate existing maintenance period
	// instead of ignoring it.
  if projects[projectName].MaintenanceId > 0 && IsExpired(projects[projectName].MaintenanceId) {
		err := DeleteMaintenance(projects[projectName].MaintenanceId)
		if err != nil {
			fmt.Fprintf(w, "Error occured while deleting expired #%d\n", projects[projectName].MaintenanceId)
			fmt.Printf("[%s] Error occured while deleting expired #%d\n", projectName, projects[projectName].MaintenanceId)
			return
		}
		fmt.Fprintf(w, "Expired old #%d was deleted.\n", projects[projectName].MaintenanceId)
		fmt.Printf("[%s] Expired old #%d was deleted.\n", projectName, projects[projectName].MaintenanceId)
		projects[projectName].MaintenanceId = 0
  }

  if projects[projectName].MaintenanceId > 0 {
		fmt.Fprintf(w,
			"Already deploying #%d, adding %d more seconds, counting from now.\n",
			projects[projectName].MaintenanceId,
			duration)
		fmt.Printf("[%s] already deploying #%d, adding %d more seconds, counting from now.\n",
			projectName,
			projects[projectName].MaintenanceId,
			duration)

		err := UpdateMaintenance(projects[projectName].MaintenanceId, duration, projects[projectName].GroupIds)

		if err != nil {
			fmt.Fprintf(w, "Error occured while updating duration for: #%d\n", projects[projectName].MaintenanceId)
			fmt.Printf("[%s] Error occured while updating duration for: #%d\n", projectName, projects[projectName].MaintenanceId)
		}
		return
  }

	inserted, err := CreateMaintenance(fmt.Sprintf("deploy: %s @ %s", projectName, time), fmt.Sprintf("deployed by %s", deployer), duration, projects[projectName].GroupIds)
	if err != nil {
		log.Fatalf("Failed to create: %v", err)
	}

	fmt.Printf("[%s] creating maintenance #%d\n", projectName, inserted)
	projects[projectName].MaintenanceId = inserted

	fmt.Fprintf(w, "maintenance %d created\n", inserted)
}

func delayedDeleteMaintenance(delay time.Duration, projectName string, id int64) {
	delay = delay / time.Second
	fmt.Printf("[%s] Expiring current maintenance #%d in %d seconds.\n", projectName, id, delay)
	err := UpdateMaintenance(projects[projectName].MaintenanceId, int(delay), projects[projectName].GroupIds)
	if err != nil {
		fmt.Printf("[%s] Error occured while expiring #%d\n", projectName, projects[projectName].MaintenanceId)
	}
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
		fmt.Fprintf(w, "removing maintenance %d after 2 minutes\n", remove)
		delayedDeleteMaintenance(2*time.Minute, projectName, remove)
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

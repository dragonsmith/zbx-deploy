package main

import (
	"flag"
	"fmt"
	"github.com/codegangsta/negroni"
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
	project := vars["project"]

	if projects[project] == nil {
		return
	}

	if projects[project].MaintenanceId > 0 {
		fmt.Fprintf(w, "already deploying: %d", projects[project].MaintenanceId)
		return
	}

	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	deployer := req.FormValue("deployer")
	time := time.Now().Format(layout)
	inserted, err := CreateMaintenance(fmt.Sprintf("deploy: %s @ %s", project, time), fmt.Sprintf("deployed by %s", deployer), 600, projects[project].GroupIds)
	if err != nil {
		log.Fatalf("Failed to create: %v", err)
	}

	projects[project].MaintenanceId = inserted

	fmt.Fprintf(w, "maintenance %d created\n", inserted)
}

func finishHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	project := vars["project"]

	if projects[project] == nil {
		return
	}

	remove := projects[project].MaintenanceId

	fmt.Fprintf(w, "removing maintenance %d\n", remove)
	if remove > 0 {
		fmt.Printf("deleting %d\n", remove)
		DeleteMaintenance(remove)

		projects[project].MaintenanceId = 0
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
			fmt.Printf("cleaning up %s %d", key, value.MaintenanceId)
			DeleteMaintenance(value.MaintenanceId)
		}
	}
}

func prepareProjects() map[string]*projectStatus {
	projects = make(map[string]*projectStatus, 0)

	for k, v := range config.Projects {
		projects[k] = &projectStatus{v, 0}
	}

	return projects
}

func TokenMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	token := r.Header.Get("X-Auth-Token")
	if token != config.Token {
		http.Error(rw, "invalid token", 401)
	} else {
		next(rw, r)
	}
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

	n := negroni.New()
	n.Use(negroni.HandlerFunc(TokenMiddleware))
	n.UseHandler(r)

	n.Run(bindAddress)
}

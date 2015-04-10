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
)

func helloHandler(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "http://evl.ms", http.StatusFound)
}

func startHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	project := vars["project"]

	if projects[project] == nil {
		return
	}

	inserted, err := CreateMaintenance("demo", "demo", 600, projects[project].GroupIds)
	if err != nil {
		log.Fatalf("Failed to create: %v", err)
	}

	projects[project].MaintenanceId = inserted

	fmt.Fprintf(w, "maintenance %d created\n", remove)
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

func prepareProjects(config Config) map[string]*projectStatus {
	projects = make(map[string]*projectStatus, 0)

	for k, v := range config.Projects {
		projects[k] = &projectStatus{[]int{v}, 0}
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

	fmt.Printf("config: %v\n", config)
	projects = prepareProjects(config)
	fmt.Printf("projects: %v\n", projects)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("captured %v, cleaning up and exiting..", sig)
			CleanupMaintenances()
			os.Exit(1)
		}
	}()

	endpoint := os.Getenv("ZBX_ENDPOINT")
	username := os.Getenv("ZBX_USER")
	password := os.Getenv("ZBX_PASSWORD")

	fmt.Printf("endpoint: %s\n", endpoint)
	LoginZabbix(endpoint, username, password)

	r := mux.NewRouter()
	r.HandleFunc("/", helloHandler)
	r.HandleFunc("/start/{project}", startHandler)
	r.HandleFunc("/finish/{project}", finishHandler)

	n := negroni.New()
	n.Use(negroni.HandlerFunc(TokenMiddleware))
	n.UseHandler(r)

	n.Run(bindAddress)
}

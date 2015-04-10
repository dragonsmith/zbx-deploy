package main

import (
	"fmt"
	"github.com/kirs/zabbix"
	"os"
	"time"
)

func main() {
	endpoint := os.Getenv("ZBX_ENDPOINT")
	username := os.Getenv("ZBX_USER")
	password := os.Getenv("ZBX_PASSWORD")
	api, err := zabbix.NewAPI(endpoint, username, password)
	if err != nil {
		fmt.Println(err)
		return
	}

	versionresult, err := api.Version()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("version: %s\n", versionresult)

	_, err = api.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Connected to API!")

	fmt.Printf("api.auth: %s", api.GetAuth())

	since := time.Now().Unix()
	params := make(map[string]interface{}, 0)
	params["groupids"] = []int{12}
	params["name"] = "from_go_with_love"
	params["maintenance_type"] = 0
	params["description"] = "created from go"
	params["active_since"] = since
	params["active_till"] = since + 600

	timeperiods := make(map[string]string, 0)
	timeperiods["timeperiod_type"] = "0"
	timeperiods["start_date"] = fmt.Sprintf("%d", since)
	timeperiods["period"] = "600"
	params["timeperiods"] = timeperiods

	response, err := api.ZabbixRequest("maintenance.create", params)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		panic(err)
	}

	if response.Error.Code != 0 {
		panic(&response.Error)
	}

	fmt.Printf("result: %v\n", response.Result)
}

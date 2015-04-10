package main

import (
	"fmt"
	"github.com/kirs/zabbix"
	"strconv"
	"time"
)

var api *zabbix.API

func LoginZabbix(endpoint, username, password string) {
	var err error
	api, err = zabbix.NewAPI(endpoint, username, password)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	versionresult, err := api.Version()
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	fmt.Printf("version: %s\n", versionresult)

	_, err = api.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Connected to API!")

	fmt.Printf("api.auth: %s", api.GetAuth())
}

func CreateMaintenance(name string, description string, duration int, groupids []int) (int64, error) {
	since := time.Now().Unix()
	params := make(map[string]interface{})
	params["groupids"] = groupids
	params["name"] = name
	params["maintenance_type"] = 0
	params["description"] = description
	params["active_since"] = since
	params["active_till"] = since + int64(duration)

	period := make(map[string]string)
	period["timeperiod_type"] = "0"
	period["start_date"] = fmt.Sprintf("%d", since)
	period["period"] = fmt.Sprintf("%d", duration)

	timeperiods := []map[string]string{}
	timeperiods = append(timeperiods, period)
	params["timeperiods"] = timeperiods

	response, err := api.ZabbixRequest("maintenance.create", params)
	if err != nil {
		return 0, err
	}

	if response.Error.Code != 0 {
		return 0, err
	}

	ids := response.Result.(map[string]interface{})
	created := ids["maintenanceids"].([]interface{})
	inserted := created[0].(string)

	id, _ := strconv.ParseInt(inserted, 10, 0)
	return id, nil
}

func DeleteMaintenance(id int64) error {
	params := make([]int64, 1)
	params[0] = id

	response, err := api.ZabbixRequest("maintenance.delete", params)
	if err != nil {
		return err
	}

	if response.Error.Code != 0 {
		return err
	}

	return nil
}
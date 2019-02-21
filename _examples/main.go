package main

import (
	"fmt"

	lwApi "github.com/liquidweb/go-lwApi"
)

type ZoneDetails struct {
	lwApi.LWAPIError
	AvlZone     string   `json:"availability_zone"`
	Desc        string   `json:"description"`
	GatewayDevs []string `json:"gateway_devices"`
	HvType      string   `json:"hv_type"`
	ID          int      `json:"id"`
	Legacy      int      `json:"legacy"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	SourceHVs   []string `json:"valid_source_hvs"`
}

func main() {
	config := lwApi.LWAPIConfig{
		Username: "ExampleUsername",
		Password: "ExamplePassword",
		Url:      "api.liquidweb.com",
	}
	apiClient, iErr := lwApi.New(&config)
	if iErr != nil {
		panic(iErr)
	}
	args := map[string]interface{}{
		"uniq_id": "SJ9NG6",
	}
	got, gotErr := apiClient.Call("bleed/asset/details", args)
	if gotErr != nil {
		panic(gotErr)
	}

	fmt.Printf("RETURNED:\n\n%+v\n\n", got)

	var zone ZoneDetails
	zArgs := map[string]interface{}{
		"id": 1,
	}
	err := apiClient.CallInto("network/zone/details", zArgs, &zone)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Got struct %#v\n", zone)
}

package main

import (
	"fmt"
	lwInternalApi "github.com/liquidweb/go-lwInternalApi"
)

func main() {
	apiClient, iErr := lwInternalApi.New()
	if iErr != nil {
		panic(iErr)
	}
	args := map[string]interface{}{
		"uniq_id": "2UPHPL",
	}
	got, gotErr := apiClient.Call("bleed/asset/details", args)
	if gotErr != nil {
		panic(gotErr)
	}
	fmt.Printf("RETURNED:\n\n%+v\n\n", got)
}

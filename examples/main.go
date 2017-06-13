package main

import (
	"github.com/spf13/viper"
	"fmt"
	lwInternalApi "github.com/liquidweb/go-lwInternalApi"
)

func main() {
	config := viper.New()
	config.SetConfigName(".go-lwInternalApi")
	config.AddConfigPath("/usr/local/lp/etc")
	// Match environment variables as well
	config.AutomaticEnv()

	config.ReadInConfig()

	apiClient, iErr := lwInternalApi.New(config)
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

# go-lwApi
LiquidWeb  API Golang client
## Setting up Authentication
When creating an api client, it expects to be configured with a viper config. Here is an example of how to get an api client.

```
package main

import (
	"fmt"

	lwApi "github.com/liquidweb/go-lwApi"
	"github.com/spf13/viper"
)

func main() {
	config := viper.New()
	config.SetConfigName("lwApi")
	config.AddConfigPath(".")
	// Match environment variables as well
	config.AutomaticEnv()

	viperErr := config.ReadInConfig()
	if viperErr != nil {
		panic(viperErr)
	}

	config.Debug()

	apiClient, iErr := lwApi.New(config)
}
```

In this scenario, we rely on a configuration file in the current working directory named lwApi.{toml,yaml,json} to set up a viper client.
This file might look like the following example:
``` toml
[lwApi]
username = "SUPERDUPERUSER"
password = "SUPERDUPERPW"
url = "https://api.stormondemand.com"
timeout = 15
```
## Importing
``` go
import (
        lwApi "github.com/liquidweb/go-lwApi"
)
```
## Calling a method
``` go
apiClient, iErr := lwApi.New(config)
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
```

As you can see, you don't need to prefix the `params` key, as that is handled in the `Call()` function for you.

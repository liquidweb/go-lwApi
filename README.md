# go-lwInternalApi
LiquidWeb Internal API Golang client
## Setting up Authentication
This package reads `/usr/local/lp/etc/.go-lwInternalApi.{toml,yaml,json}` for authentication details. Example:
```sh
ssullivan@data ~/golang/src/github.com/liquidweb/go-lwInternalApi $ cat /usr/local/lp/etc/.go-lwInternalApi.toml 
user = "SUPERDUPERUSER"
pw = "SUPERDUPERPW"
url = "https://api-internal.ssullivan.dev.liquidweb.com:20800"
timeout = 15
ssullivan@data ~/golang/src/github.com/liquidweb/go-lwInternalApi $ 
```
## Importing
``` go
import (
        lwInternalApi "github.com/liquidweb/go-lwInternalApi"
)
```
## Calling a method
``` go
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
```

As you can see, you don't need to prefix the `params` key, as that is handled in the `Call()` function for you.

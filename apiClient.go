package lwInternalApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
}

/* public */

func New() (*Client, error) {
	if err := initConfig(); err != nil {
		return nil, err
	}
	if err := santizeConfig(); err != nil {
		return nil, err
	}
	timeout := viper.GetInt("timeout")
	if timeout == 0 {
		timeout = 20
	}
	httpClient := &http.Client{Timeout: time.Duration(time.Duration(timeout) * time.Second)}
	client := Client{httpClient}
	return &client, nil
}

func (client *Client) Call(method string, params interface{}) (interface{}, error) {
	// internal api wants the "params" prefix key. Do it here so consumers dont have
	// to do this everytime.
	args := map[string]interface{}{
		"params": params,
	}
	encodedArgs, encodeErr := json.Marshal(args)
	if encodeErr != nil {
		return nil, encodeErr
	}
	// formulate the HTTP POST request
	url := fmt.Sprintf("%s/%s", viper.GetString("url"), method)
	req, reqErr := http.NewRequest("POST", url, bytes.NewReader(encodedArgs))
	if reqErr != nil {
		return nil, reqErr
	}
	// HTTP basic auth
	req.SetBasicAuth(viper.GetString("user"), viper.GetString("pw"))
	// make the POST request
	resp, doErr := client.httpClient.Do(req)
	if doErr != nil {
		return nil, doErr
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad HTTP response code [%d] from [%s]", resp.StatusCode, viper.GetString("url"))
	}
	// read the response body into a byte slice
	bsRb, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	// json decode into interface
	var decodedResp interface{}
	if jsonDecodeErr := json.Unmarshal(bsRb, &decodedResp); jsonDecodeErr != nil {
		return nil, jsonDecodeErr
	}
	mapDecodedResp := decodedResp.(map[string]interface{})
	// handle LW specific error responses
	errConditions := map[string]interface{}{
		"error_class": mapDecodedResp["error_class"],
		"error":       mapDecodedResp["error"],
	}
	for _, value := range errConditions {
		switch t := value.(type) {
		case string:
			if t != "" {
				return nil, fmt.Errorf("%s: %+v", value, mapDecodedResp)
			}
		}
	}
	// no LW errors so return the decoded response
	return mapDecodedResp, nil
}

/* private */

func initConfig() error {
	viper.SetConfigName(".go-lwInternalApi") // name of config file (without extension)
	viper.AddConfigPath("/usr/local/lp/etc") // adding home directory as first search path
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

func santizeConfig() error {
	if viper.GetString("user") == "" {
		return fmt.Errorf("user is missing from config file: [%s]", viper.ConfigFileUsed())
	}
	if viper.GetString("pw") == "" {
		return fmt.Errorf("pw is missing from config file: [%s]", viper.ConfigFileUsed())
	}
	if viper.GetString("url") == "" {
		return fmt.Errorf("url is missing from config file: [%s]", viper.ConfigFileUsed())
	}
	return nil
}

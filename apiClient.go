package lwInternalApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"time"
	"crypto/tls"
)

type Client struct {
	config     *viper.Viper
	httpClient *http.Client
}

/* public */

func New(config *viper.Viper) (*Client, error) {
	if err := santizeConfig(config); err != nil {
		return nil, err
	}
	timeout := config.GetInt("lwInternalApi.timeout")
	if timeout == 0 {
		timeout = 20
	}

	httpClient := &http.Client{Timeout: time.Duration(time.Duration(timeout) * time.Second)}

	if config.GetBool("lwInternalApi.secure") != true {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient.Transport = tr
	}
	client := Client{config, httpClient}
	return &client, nil
}

func (client *Client) Call(method string, params interface{}) (interface{}, error) {
	thisViper := client.config
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
	url := fmt.Sprintf("%s/%s", thisViper.GetString("lwInternalApi.url"), method)
	req, reqErr := http.NewRequest("POST", url, bytes.NewReader(encodedArgs))
	if reqErr != nil {
		return nil, reqErr
	}
	// HTTP basic auth
	req.SetBasicAuth(thisViper.GetString("lwInternalApi.username"), thisViper.GetString("lwInternalApi.password"))
	// make the POST request
	resp, doErr := client.httpClient.Do(req)
	if doErr != nil {
		return nil, doErr
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad HTTP response code [%d] from [%s]", resp.StatusCode, thisViper.GetString("lwInternalApi.url"))
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
	return decodedResp, nil
}

/* private */

func santizeConfig(config *viper.Viper) error {
	if config.GetString("lwInternalApi.username") == "" {
		return fmt.Errorf("lwInternalApi.username is missing from config file: [%s]", config.ConfigFileUsed())
	}
	if config.GetString("lwInternalApi.password") == "" {
		return fmt.Errorf("lwInternalApi.password is missing from config file: [%s]", config.ConfigFileUsed())
	}
	if config.GetString("lwInternalApi.url") == "" {
		return fmt.Errorf("lwInternalApi.url is missing from config file: [%s]", config.ConfigFileUsed())
	}
	return nil
}

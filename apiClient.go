package lwInternalApi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type Client struct {
	config     *viper.Viper
	httpClient *http.Client
}

type LWAPIError struct {
	ErrorMsg     string `json:"error,omitempty"`
	ErrorClass   string `json:"error_class,omitempty"`
	ErrorFullMsg string `json:"full_message,omitempty"`
}

func (e LWAPIError) Error() string {
	return fmt.Sprintf("%v: %v", e.ErrorClass, e.ErrorFullMsg)
}

func (e LWAPIError) HadError() bool {
	return e.ErrorClass != ""
}

type LWAPIRes interface {
	Error() string
	HadError() bool
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

// Call takes a path, such as "network/zone/details" and a params structure.
// It is recommended that the params be a map[string]interface{}, but you can use
// anything that serializes to the right json structure.
// A `interface{}` and an error are returned, in typical go fasion.
func (client *Client) Call(method string, params interface{}) (interface{}, error) {
	bsRb, err := client.CallRaw(method, params)
	if err != nil {
		return nil, err
	}

	// json decode into interface
	var decodedResp interface{}
	if jsonDecodeErr := json.Unmarshal(bsRb, &decodedResp); jsonDecodeErr != nil {
		return nil, jsonDecodeErr
	}
	mapDecodedResp := decodedResp.(map[string]interface{})
	errorClass, ok := mapDecodedResp["error_class"]
	if ok {
		errorClassStr := errorClass.(string)
		if errorClassStr != "" {
			return nil, LWAPIError{
				ErrorClass:   errorClassStr,
				ErrorFullMsg: mapDecodedResp["full_message"].(string),
				ErrorMsg:     mapDecodedResp["error"].(string),
			}
		}
	}
	// no LW errors so return the decoded response
	return decodedResp, nil
}

// CallInto is like call, but instead of returning an interface you pass it a
// struct which is filled, much like the json.Unmarshal function.  The struct
// you pass must satisfy the LWAPIRes interface.  If you embed the LWAPIError
// struct from this package into your struct, this will be taken care of for you.
//
// Example:
//	type ZoneDetails struct {
//		lwInternalApi.LWAPIError
//		AvlZone     string   `json:"availability_zone"`
//		Desc        string   `json:"description"`
//		GatewayDevs []string `json:"gateway_devices"`
//		HvType      string   `json:"hv_type"`
//		ID          int      `json:"id"`
//		Legacy      int      `json:"legacy"`
//		Name        string   `json:"name"`
//		Status      string   `json:"status"`
//		SourceHVs   []string `json:"valid_source_hvs"`
//	}
//	var zone ZoneDetails
//	err = apiClient.CallInto("network/zone/details", paramers, &zone)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Got struct %#v\n", zone)
//
func (client *Client) CallInto(method string, params interface{}, into LWAPIRes) error {
	bsRb, err := client.CallRaw(method, params)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bsRb, into)
	if err != nil {
		return err
	}

	if into.HadError() {
		// the LWAPIRes satisfies the Error interface, so we can just return it on
		// error.
		return into
	}

	return nil
}

// CallRaw is just like Call, except it returns the raw json as a byte slice.
func (client *Client) CallRaw(method string, params interface{}) ([]byte, error) {
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
		return nil, fmt.Errorf("Bad HTTP response code [%d] from [%s]", resp.StatusCode, url)
	}
	// read the response body into a byte slice
	bsRb, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	return bsRb, nil
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

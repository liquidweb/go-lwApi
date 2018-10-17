// Package lwApi is a minimalist API client to LiquidWeb's (https://www.liquidweb.com) API:
//
// https://cart.liquidweb.com/storm/api/docs/v1
//
// https://cart.liquidweb.com/storm/api/docs/bleed
//
// As you might have guessed from the above API documentation links, there are API versions:
// "v1" and "bleed". As the name suggests, if you always want the latest features and abilities,
// use "bleed". If you want long term compatibility (at the cost of being a little further behind
// sometimes), use "v1".
package lwApi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// A Client holds the packages *viper.Viper and *http.Client. To get a *Client, call New.
type Client struct {
	config        *viper.Viper
	httpClient    *http.Client
	BeforeRequest func(*http.Request)
	AfterRequest  func(*http.Request, *http.Response, time.Duration)
}

// A LWAPIError is used to identify error responses when JSON unmarshalling json from a
// byte slice.
type LWAPIError struct {
	ErrorMsg     string `json:"error,omitempty"`
	ErrorClass   string `json:"error_class,omitempty"`
	ErrorFullMsg string `json:"full_message,omitempty"`
}

// Given a LWAPIError, returns a string containing the ErrorClass and ErrorFullMsg.
func (e LWAPIError) Error() string {
	return fmt.Sprintf("%v: %v", e.ErrorClass, e.ErrorFullMsg)
}

// Given a LWAPIError, returns boolean if ErrorClass was present or not. You can
// use this function to determine if a LWAPIRes response indicates an error or not.
func (e LWAPIError) HadError() bool {
	return e.ErrorClass != ""
}

// LWAPIRes is a convenient interface used (for example) by CallInto to ensure a passed
// struct knows how to indicate whether or not it had an error.
type LWAPIRes interface {
	Error() string
	HadError() bool
}

// New takes a *viper.Viper, and gives you a *Client. If there's an error, it is returned.
// When using this package, this should be the first function you call. Below is an example
// that demonstrates creating the *viper.Viper and passing it to New.
//
// Example:
//	config := viper.New()
//	config.SetConfigName("lwApi")
//	config.AddConfigPath(".")
//	// Match environment variables as well
//	config.AutomaticEnv()
//	if viperErr := config.ReadInConfig(); viperErr != nil {
//		panic(viperErr)
//	}
//	apiClient, newErr := lwApi.New(config)
//	if newErr != nil {
//		panic(newErr)
//	}
func New(config *viper.Viper) (*Client, error) {
	if err := santizeConfig(config); err != nil {
		return nil, err
	}
	timeout := config.GetInt("lwApi.timeout")
	if timeout == 0 {
		timeout = 20
	}

	httpClient := &http.Client{Timeout: time.Duration(time.Duration(timeout) * time.Second)}

	if config.GetBool("lwApi.secure") != true {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient.Transport = tr
	}
	client := Client{
		config:     config,
		httpClient: httpClient,
	}
	return &client, nil
}

// Call takes a path, such as "network/zone/details" and a params structure.
// It is recommended that the params be a map[string]interface{}, but you can use
// anything that serializes to the right json structure.
// A `interface{}` and an error are returned, in typical go fasion.
//
// Example:
//	args := map[string]interface{}{
//		"uniq_id": "ABC123",
//	}
//	got, gotErr := apiClient.Call("bleed/asset/details", args)
//	if gotErr != nil {
//		panic(gotErr)
//	}
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
	mapDecodedResp, ok := decodedResp.(map[string]interface{})
	if !ok {
		return nil, errors.New("endpoint did not return the expected JSON structure")
	}
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
//		lwApi.LWAPIError
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

// CallRaw is just like Call, except it returns the raw json as a byte slice. However, in contrast to
// Call, CallRaw does *not* check the API response for LiquidWeb specific exceptions as defined in
// the type LWAPIError. As such, if calling this function directly, you must check for LiquidWeb specific
// exceptions yourself.
//
// Example:
//	args := map[string]interface{}{
//		"uniq_id": "ABC123",
//	}
//	got, gotErr := apiClient.CallRaw("bleed/asset/details", args)
//	if gotErr != nil {
//		panic(gotErr)
//	}
//	// Check got now for LiquidWeb specific exceptions, as described above.
func (client *Client) CallRaw(method string, params interface{}) ([]byte, error) {
	thisViper := client.config
	//  api wants the "params" prefix key. Do it here so consumers dont have
	// to do this everytime.
	args := map[string]interface{}{
		"params": params,
	}
	encodedArgs, encodeErr := json.Marshal(args)
	if encodeErr != nil {
		return nil, encodeErr
	}
	// formulate the HTTP POST request
	url := fmt.Sprintf("%s/%s", thisViper.GetString("lwApi.url"), method)
	req, reqErr := http.NewRequest("POST", url, bytes.NewReader(encodedArgs))
	if reqErr != nil {
		return nil, reqErr
	}
	// HTTP basic auth
	req.SetBasicAuth(thisViper.GetString("lwApi.username"), thisViper.GetString("lwApi.password"))

	if client.BeforeRequest != nil {
		client.BeforeRequest(req)
	}

	// make the POST request
	start := time.Now()
	resp, doErr := client.httpClient.Do(req)
	elapsed := time.Since(start)

	if doErr != nil {
		return nil, doErr
	}
	defer resp.Body.Close()

	if client.AfterRequest != nil {
		client.AfterRequest(req, resp, elapsed)
	}

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
	if config.GetString("lwApi.username") == "" {
		return fmt.Errorf("lwApi.username is missing from config file: [%s]", config.ConfigFileUsed())
	}
	if config.GetString("lwApi.password") == "" {
		return fmt.Errorf("lwApi.password is missing from config file: [%s]", config.ConfigFileUsed())
	}
	if config.GetString("lwApi.url") == "" {
		return fmt.Errorf("lwApi.url is missing from config file: [%s]", config.ConfigFileUsed())
	}
	return nil
}

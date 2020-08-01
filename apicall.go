package apicall

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// ApiCall hold configuration
// to make a request
type ApiCall struct {
	Url string
	Method string
	Headers http.Header
}

func New(method, url string) *ApiCall {
	return &ApiCall{
		url,
		method,
		nil,
	}
}

// Send it will send a request and parse response in order to
// be compatible with BaseStandard
func (a *ApiCall) Send() (*BaseStandard, error) {
	response, err := makeRequest(a.Method, a.Url)
	if err != nil {
		return nil, err
	}

	baseResponse, err := formatResponse(response)
	if err != nil {
		return nil, err
	}

	return baseResponse, nil
}

// makeRequest is a function used internally only to make request
func makeRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// formatResponse it will pack raw response into our structure
func formatResponse(response *http.Response) (*BaseStandard, error){
	var baseResponse BaseStandard
	binary, _ := ioutil.ReadAll(response.Body)
	err := json.Unmarshal(binary, &baseResponse)
	if err != nil {
		return &baseResponse, err
	}

	baseResponse.Ok = true
	if response.StatusCode < http.StatusOK && response.StatusCode >= http.StatusMultipleChoices {
		baseResponse.Ok = false
	}

	if baseResponse.Ok {
		baseResponse.Ok = baseResponse.HasItems()
	}

	if baseResponse.Ok && baseResponse.AuditInfo.Errors.Items != nil {
		baseResponse.Ok = len(baseResponse.AuditInfo.Errors.Items) <= 0
	}

	return &baseResponse, nil
}
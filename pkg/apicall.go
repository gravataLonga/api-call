package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

// ApiCall hold configuration
// to make a request
type ApiCall struct {
	Url     string
	Method  string
	Headers http.Header
	BaseUrl string
	Timeout time.Duration
	ctx     *context.Context
	cancel  *context.CancelFunc
}

type Option func(ApiCall) *ApiCall

func NewApiCall(options ...Option) *ApiCall {
	a := &ApiCall{}
	for _, option := range options {
		a = option(*a)
	}
	return a
}

func WithUrl(url string) Option {
	return func(a ApiCall) *ApiCall {
		a.Url = url
		return &a
	}
}

func WithMethod(method string) Option {
	return func(a ApiCall) *ApiCall {
		a.Method = method
		return &a
	}
}

func WithBaseUrl(base string) Option {
	return func(a ApiCall) *ApiCall {
		a.BaseUrl = base
		return &a
	}
}

func WithTimeout(duration time.Duration) Option {
	return func(a ApiCall) *ApiCall {
		a.Timeout = duration
		return &a
	}
}

// Send it will send a request and parse response in order to
// be compatible with BaseStandard
func (a *ApiCall) Send() (*BaseStandard, error) {
	var baseResponse = newBaseStandard(a)
	response, err := makeRequest(*a.ctx, a.Method, a.Url)

	if err != nil {
		return formatExceptionResponse(baseResponse, response, err), nil
	}

	err = formatResponse(baseResponse, response)
	if err != nil {
		return nil, err
	}

	return baseResponse, nil
}

// makeRequest is a function used internally only to make request
func makeRequest(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func newBaseStandard(a *ApiCall) *BaseStandard {
	ctx := context.Background()
	// Base Settings for MakingRequest
	var baseResponse = new(BaseStandard)

	baseResponse.AuditInfo.Host, _ = os.Hostname()
	baseResponse.AuditInfo.Timestamp = time.Now()
	baseResponse.AuditInfo.ClientIP, _ = externalIP()
	var cancel context.CancelFunc
	if a.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, a.Timeout)
	}

	a.ctx = &ctx
	a.cancel = &cancel
	operationId, _ := baseResponse.NewOperationId()
	baseResponse.AuditInfo.OperationId = operationId

	return baseResponse
}

// formatResponse it will pack raw response into our structure
func formatResponse(baseResponse *BaseStandard, response *http.Response) error {
	binary, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(binary, baseResponse)
	if err != nil {
		return err
	}

	baseResponse.AuditInfo.StatusCode = response.StatusCode
	baseResponse.AuditInfo.Duration = time.Since(baseResponse.Timestamp)

	(*baseResponse).AuditInfo.Ok = true
	if baseResponse.AuditInfo.StatusCode < http.StatusOK && baseResponse.AuditInfo.StatusCode >= http.StatusMultipleChoices {
		(*baseResponse).AuditInfo.Ok = false
	}

	if (*baseResponse).AuditInfo.Ok {
		(*baseResponse).AuditInfo.Ok = baseResponse.HasItems()
	}

	if (*baseResponse).AuditInfo.Ok && (*baseResponse).AuditInfo.Errors.Items != nil {
		(*baseResponse).AuditInfo.Ok = len((*baseResponse).AuditInfo.Errors.Items) <= 0
	}

	return nil
}

func formatExceptionResponse(baseResponse *BaseStandard, response *http.Response, err error) *BaseStandard {
	var meta Meta
	if err == context.DeadlineExceeded {
		meta.Code = "1"
		meta.Description = "Timeout"
	}

	if err == context.Canceled {
		meta.Code = "2"
		meta.Description = "Canceled"
	}

	baseResponse.Errors.Items = append(baseResponse.Errors.Items, meta)
	return baseResponse
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {

		// interface down
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// loopback interface
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
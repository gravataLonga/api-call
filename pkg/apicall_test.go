package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPassingOptions(t *testing.T) {
	tables := []struct {
		options []Option
		a       ApiCall
	}{
		{
			[]Option{
				WithBaseUrl("https://google.pt"),
			},
			ApiCall{
				http.Header{},
				"https://google.pt",
				0,
				nil,
				nil,
			},
		},
	}

	for index, table := range tables {
		testName := fmt.Sprintf("Test %v", index)
		t.Run(testName, func(t *testing.T) {

			a := NewApiCall(table.options...)

			if a.BaseUrl != table.a.BaseUrl {
				t.Errorf("Base url not match, expect: %v and got %v", a.BaseUrl, table.a.BaseUrl)
			}

			if a.Timeout != table.a.Timeout {
				t.Errorf("Timeout not match, expect: %v and got %v", a.Timeout, table.a.Timeout)
			}

		})
	}
}

func TestBuildUrl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		json := `{"auditInfo":{},"items":[{"method":"` + request.Method + `","url":"` + request.URL.Path + `"}],"interfaceSettings":{}}`
		_, _ = writer.Write([]byte(json))
	}))
	defer ts.Close()
	apicall := NewApiCall(
		WithTimeout(7*time.Second),
		WithBaseUrl(ts.URL),
	)
	response, err := apicall.Send("POST", "/testing", strings.NewReader(`{"body":"hello"}`))
	assert.Nil(t, err)
	assert.Empty(t, response.AuditInfo.Errors)
	assert.Zero(t, response.AuditInfo.Errors)
	assert.Empty(t, response.AuditInfo.Warning)
	assert.Zero(t, response.AuditInfo.Warning)
	assert.Empty(t, response.AuditInfo.Info)
	assert.Zero(t, response.AuditInfo.Info)
}

func TestCanAddHeaders(t *testing.T) {
	apicall := NewApiCall()
	apicall.Headers.Set("hello", "world")

	assert.NotNil(t, apicall.Headers)
	assert.Equal(t, "world", apicall.Headers.Get("Hello"))
}

func TestCanAddHeaderWithOption(t *testing.T) {
	apicall := NewApiCall(func(a ApiCall) *ApiCall {
		a.Headers.Set("hello", "world")
		return &a
	})

	assert.NotNil(t, apicall.Headers)
	assert.Equal(t, "world", apicall.Headers.Get("Hello"))
}

func TestOptionBearerToken(t *testing.T) {
	apicall := NewApiCall(
		WithAuthentication("Username", "Password"),
	)

	assert.NotNil(t, apicall.Headers)
	assert.Equal(t, "Basic VXNlcm5hbWU6UGFzc3dvcmQ=", apicall.Headers.Get("Authorization"))
}

func TestBasicSend(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		json := `{"auditInfo":{},"items":[{"echo":"Hello World"}],"interfaceSettings":{}}`
		_, _ = writer.Write([]byte(json))
	}))
	defer ts.Close()

	apicall := NewApiCall(
		WithTimeout(7 * time.Second),
	)

	response, err := apicall.Send("GET", ts.URL, nil)
	host, _ := os.Hostname()
	ip, _ := externalIP()

	assert.Nil(t, err, "Don't expect error here")
	assert.NotNil(t, apicall)
	assert.Equal(t, host, response.Host)
	assert.Equal(t, ip, response.ClientIP)
	assert.Equal(t, 200, response.StatusCode)
	assert.NotEmpty(t, response.OperationId)
	assert.Empty(t, response.AuditInfo.Errors)
	assert.Zero(t, response.AuditInfo.Errors)
	assert.Empty(t, response.AuditInfo.Warning)
	assert.Zero(t, response.AuditInfo.Warning)
	assert.Empty(t, response.AuditInfo.Info)
	assert.Zero(t, response.AuditInfo.Info)
	assert.True(t, response.IsOk())
}

func TestCanParseItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		base := new(BaseStandard)
		resp := json.RawMessage(`[{"echo":"Hello World"}]`)
		base.Items = &resp
		response, err := json.Marshal(base)
		assert.Nil(t, err, "Unable to marshal json")
		_, _ = writer.Write(response)
	}))
	defer ts.Close()
	type MyItems struct {
		Echo string `json:"echo"`
	}
	expectedItems := []MyItems{{"Hello World"}}
	myItems := []MyItems{}

	apicall := NewApiCall(
		WithTimeout(7 * time.Second),
	)
	response, err := apicall.Send("GET", ts.URL, nil)

	err = response.GetItems(&myItems)
	assert.NotNil(t, myItems)
	assert.Nil(t, err, "Don't expect error here")
	assert.Len(t, myItems, 1)
	assert.Empty(t, response.AuditInfo.Errors)
	assert.Zero(t, response.AuditInfo.Errors)
	assert.Empty(t, response.AuditInfo.Warning)
	assert.Zero(t, response.AuditInfo.Warning)
	assert.Empty(t, response.AuditInfo.Info)
	assert.Zero(t, response.AuditInfo.Info)
	assert.Exactly(t, myItems, expectedItems, "Unable to get items from response")
	assert.True(t, response.IsOk())
}

func TestCanParseOk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(201)
		writer.Header().Set("Content-Type", "application/json")
		base := new(BaseStandard)
		resp := json.RawMessage(`[{"echo":"Hello World"}]`)
		base.Items = &resp
		response, err := json.Marshal(base)
		assert.Nil(t, err, "Unable to marshal json")
		_, _ = writer.Write(response)
	}))
	defer ts.Close()

	apicall := NewApiCall(
		WithTimeout(7 * time.Second),
	)
	response, err := apicall.Send("GET", ts.URL, nil)
	assert.Nil(t, err)
	assert.Empty(t, response.AuditInfo.Errors)
	assert.Zero(t, response.AuditInfo.Errors)
	assert.Empty(t, response.AuditInfo.Warning)
	assert.Zero(t, response.AuditInfo.Warning)
	assert.Empty(t, response.AuditInfo.Info)
	assert.Zero(t, response.AuditInfo.Info)
	assert.Exactly(t, true, response.IsOk())
}

func TestCanNotParseOk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(299)
		writer.Header().Set("Content-Type", "application/json")
		base := new(BaseStandard)
		response, err := json.Marshal(base)
		assert.Nil(t, err, "Unable to marshal json")
		_, _ = writer.Write(response)
	}))
	defer ts.Close()

	apicall := NewApiCall(
		WithTimeout(7 * time.Second),
	)
	response, err := apicall.Send("GET", ts.URL, nil)
	assert.Nil(t, err)
	assert.Exactly(t, false, response.IsOk())
}

func TestItemsCannotBeEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(299)
		writer.Header().Set("Content-Type", "application/json")
		base := new(BaseStandard)
		resp := json.RawMessage(`{}`)
		base.Items = &resp
		response, err := json.Marshal(base)
		assert.Nil(t, err, "Unable to marshal json")
		_, _ = writer.Write(response)
	}))
	defer ts.Close()

	apicall := NewApiCall(
		WithTimeout(7 * time.Second),
	)
	response, err := apicall.Send("GET", ts.URL, nil)
	assert.Nil(t, err)
	assert.Exactly(t, false, response.IsOk())
}

func TestIfReturnErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(299)
		writer.Header().Set("Content-Type", "application/json")
		base := new(BaseStandard)
		resp := json.RawMessage(`[{"echo":"Hello World"}]`)
		base.Items = &resp
		base.AuditInfo.Errors.Items = append(base.AuditInfo.Errors.Items, Meta{
			Code:        "200",
			Description: "An error happen!",
		})
		response, err := json.Marshal(base)
		assert.Nil(t, err, "Unable to marshal json")
		_, _ = writer.Write(response)
	}))
	defer ts.Close()

	apicall := NewApiCall(
		WithTimeout(7 * time.Second),
	)
	response, err := apicall.Send("GET", ts.URL, nil)
	assert.Nil(t, err)
	assert.Exactly(t, false, response.IsOk())
	assert.NotEmpty(t, response.AuditInfo.Errors.Items)
	assert.Equal(t, "An error happen!", response.AuditInfo.Errors.Items[0].Description)
	assert.Equal(t, "200", response.AuditInfo.Errors.Items[0].Code)
}

func TestTimeoutReach(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(200 * time.Millisecond)
		writer.Header().Set("Content-Type", "application/json")
		j := `{"auditInfo":{},"items":[{"echo":"Hello World"}],"interfaceSettings":{}}`
		_, _ = writer.Write([]byte(j))
	}))
	defer ts.Close()

	apicall := NewApiCall(
		WithTimeout(100 * time.Millisecond),
	)
	response, err := apicall.Send("GET", ts.URL, nil)
	assert.Nil(t, err)
	assert.Exactly(t, false, response.IsOk())
	assert.Equal(t, "Timeout", response.AuditInfo.Errors.Items[0].Description)
	assert.Equal(t, "1", response.AuditInfo.Errors.Items[0].Code)
}

func TestAddHeader(t *testing.T) {
	apicall := NewApiCall(
		WithTimeout(100 * time.Millisecond),
	)
	apicall.Headers.Set("token", "abcdefghijk")
	assert.Equal(t, "abcdefghijk", apicall.Headers.Get("Token"))
}

func TestGetTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		writer.Header().Set("Content-Type", "application/json")
		j := `{"auditInfo":{},"items":[],"interfaceSettings":{}}`
		_, _ = writer.Write([]byte(j))
	}))
	defer ts.Close()
	apicall := NewApiCall(
		WithTimeout(100 * time.Millisecond),
	)
	resp, _ := apicall.Send("GET", ts.URL, strings.NewReader(`{"hello":"world"}`))
	assert.NotEmpty(t, resp.AuditInfo.Errors.Items)
	assert.Equal(t, "Timeout", resp.AuditInfo.Errors.Items[0].Description)
	assert.Equal(t, "1", resp.AuditInfo.Errors.Items[0].Code)
}

func TestIfUnableDecodeGetFullResponseAsJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write([]byte("Hello World"))
	}))
	defer ts.Close()
	apicall := NewApiCall(
		WithTimeout(100*time.Millisecond),
		WithBaseUrl(ts.URL),
	)
	resp, err := apicall.Send("GET", "/hello-world", nil)

	assert.Nil(t, err)
	assert.NotEmpty(t, resp.AuditInfo.Errors.String())
	assert.Equal(t, "[syntaxerror]: invalid character 'H' looking for beginning of value[1] - Hello World", resp.AuditInfo.Errors.String())
}

func TestSetProperlyHeader(t *testing.T) {
	var header string
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		header = r.Header.Get("Authorization")
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write([]byte("Hello World"))
	}))
	defer ts.Close()
	apicall := NewApiCall(
		WithTimeout(100*time.Millisecond),
		WithBaseUrl(ts.URL),
		WithAuthentication("jonathan", "12345678"),
	)
	apicall.Send("GET", "/hello-world", nil)

	assert.NotEmpty(t, header)
}

func TestCanParseAuditInfo(t *testing.T) {
	type ItemToken struct {
		Token string `json:"token"`
	}
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		json := `{"items":{"token":"token"},"auditInfo":{"duration":0.027,"host":"test.com","dbHost":"dbHost","clientIP":"127.0.0.1","ok":true,"errors":{"items":[]},"info":{"items":[]},"warning":{"items":[]},"total":1,"env":"local","version":"1.0.0 - 2020-01-01 00:00:00","tokenAudience":""},"interfaceSettings":[]}`
		_, _ = writer.Write([]byte(json))
	}))
	defer ts.Close()
	apicall := NewApiCall(
		WithBaseUrl(ts.URL),
	)
	response, err := apicall.Send("POST", "/", nil)
	items := &ItemToken{}
	errGetItems := response.GetItems(items)

	assert.Nil(t, errGetItems)
	assert.Nil(t, err)
	assert.Equal(t, "token", items.Token)
	assert.Equal(t, 0.027, response.AuditInfo.Duration)
	assert.Equal(t, "test.com", response.AuditInfo.Host)
	assert.Equal(t, "127.0.0.1", response.AuditInfo.ClientIP)
	assert.Equal(t, true, response.AuditInfo.Ok)

	assert.Empty(t, response.AuditInfo.Errors.Items)
	assert.Empty(t, response.AuditInfo.Warning.Items)
	assert.Empty(t, response.AuditInfo.Info.Items)
}

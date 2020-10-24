package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
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
				WithMethod(http.MethodGet),
				WithUrl("hello"),
				WithBaseUrl("https://google.pt"),
			},
			ApiCall{
				"hello",
				"GET",
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

			if a.Method != table.a.Method {
				t.Errorf("Method not match, expect: %v and got %v", a.Method, table.a.Method)
			}

			if a.Url != table.a.Url {
				t.Errorf("Url not match, expect: %v and got %v", a.Url, table.a.Url)
			}

			if a.Timeout != table.a.Timeout {
				t.Errorf("Timeout not match, expect: %v and got %v", a.Timeout, table.a.Timeout)
			}

		})
	}
}

func TestBasicSend(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		json := `{"auditInfo":{},"items":[{"echo":"Hello World"}],"interfaceSettings":{}}`
		_, _ = writer.Write([]byte(json))
	}))
	defer ts.Close()

	apicall := NewApiCall(
		WithMethod("GET"),
		WithUrl(ts.URL),
		WithTimeout(7*time.Second),
	)

	response, err := apicall.Send()
	host, _ := os.Hostname()
	ip, _ := externalIP()

	assert.Nil(t, err, "Don't expect error here")
	assert.NotNil(t, apicall)
	assert.Equal(t, host, response.Host)
	assert.Equal(t, ip, response.ClientIP)
	assert.Equal(t, 200, response.StatusCode)
	assert.NotEmpty(t, response.OperationId)
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
		WithMethod("GET"),
		WithUrl(ts.URL),
		WithTimeout(7*time.Second),
	)
	response, err := apicall.Send()

	err = response.GetItems(&myItems)
	assert.NotNil(t, myItems)
	assert.Nil(t, err, "Don't expect error here")
	assert.Len(t, myItems, 1)
	assert.Exactly(t, myItems, expectedItems, "Unable to get items from response")
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
		WithMethod("GET"),
		WithUrl(ts.URL),
		WithTimeout(7*time.Second),
	)
	response, err := apicall.Send()
	assert.Nil(t, err)
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
		WithMethod("GET"),
		WithUrl(ts.URL),
		WithTimeout(7*time.Second),
	)
	response, err := apicall.Send()
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
		WithMethod("GET"),
		WithUrl(ts.URL),
		WithTimeout(7*time.Second),
	)
	response, err := apicall.Send()
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
		WithMethod("GET"),
		WithUrl(ts.URL),
		WithTimeout(7*time.Second),
	)
	response, err := apicall.Send()
	assert.Nil(t, err)
	assert.Exactly(t, false, response.IsOk())
}

func TestCanCancelRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(200 * time.Millisecond)
		writer.Header().Set("Content-Type", "application/json")
		j := `{"auditInfo":{},"items":[{"echo":"Hello World"}],"interfaceSettings":{}}`
		_, _ = writer.Write([]byte(j))
	}))
	defer ts.Close()

	apicall := NewApiCall(
		WithMethod("GET"),
		WithUrl(ts.URL),
		WithTimeout(100*time.Millisecond),
	)
	response, err := apicall.Send()
	assert.Nil(t, err)
	assert.Exactly(t, false, response.IsOk())
}

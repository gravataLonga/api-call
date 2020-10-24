package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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

func Test(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		writer.Header().Set("Content-Type", "application/json")
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		j := `{"auditInfo":{},"items":[` + string(b) + `}],"interfaceSettings":{}}`
		_, _ = writer.Write([]byte(j))
	}))
	defer ts.Close()
	apicall := NewApiCall(
		WithTimeout(100 * time.Millisecond),
	)
	resp, _ := apicall.Send("GET", ts.URL, strings.NewReader(`{"hello":"world"}`))
	fmt.Println(resp)
}

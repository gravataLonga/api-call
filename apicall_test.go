package apicall

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)


func TestBasicSend(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		json := `{"auditInfo":{},"items":[{"echo":"Hello World"}],"interfaceSettings":{}}`
		_, _ = writer.Write([]byte(json))
	}))
	defer ts.Close()

	apicall := New("GET", ts.URL, "", 7 * time.Second)
	response, err := apicall.Send()
	host, _ := os.Hostname()
	ip, _ := externalIP()

	assert.Nil(t, err, "Don't expect error here")
	assert.NotNil(t, apicall)
	assert.NotNil(t, response.cancel)
	assert.NotNil(t, response.ctx)
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


	apicall := New("GET", ts.URL, "", 7 * time.Second)
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

	apicall := New("GET", ts.URL, "", 7 * time.Second)
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

	apicall := New("GET", ts.URL, "", 7 * time.Second)
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

	apicall := New("GET", ts.URL, "", 7 * time.Second)
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

	apicall := New("GET", ts.URL, "", 7 * time.Second)
	response, err := apicall.Send()
	assert.Nil(t, err)
	assert.Exactly(t, false, response.IsOk())
}

func TestCanCancelRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(5 * time.Second)
		writer.Header().Set("Content-Type", "application/json")
		j := `{"auditInfo":{},"items":[{"echo":"Hello World"}],"interfaceSettings":{}}`
		_, _ = writer.Write([]byte(j))
	}))
	defer ts.Close()

	apicall := New("GET", ts.URL, "", 1 * time.Second)
	response, err := apicall.Send()
	assert.Nil(t, err)
	assert.Exactly(t, false, response.IsOk())

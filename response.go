package apicall

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"
)

// Meta hold information of AuditInfo
// for keys Errors, Info, Warning
type Meta struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// AuditInfo holds information about
// request/response from server side
type AuditInfo struct {
	// Duration of request
	Duration 	time.Duration 	    `json:"duration"`
	// Timestamp when the request started
	Timestamp 	time.Time 		    `json:"timestamp"`
	// Host is hostname of made request
	Host     	string  			`json:"host"`
	// ClientIP who made request
	ClientIP 	string  			`json:"clientIP"`
	// Ok if we got success request
	Ok       	bool    			`json:"ok"`
	// StatusCode of result
	StatusCode  int 				`json:"statusCode"`
	// OperationId is a random string for logging purpose
	OperationId string 				`json:"operationId"`
	Errors   	struct {
		Items []Meta `json:"items"`
	} `json:"errors"`
	Info struct {
		Items []Meta `json:"items"`
	} `json:"info"`
	Warning struct {
		Items []Meta `json:"items"`
	} `json:"warning"`
	Total int64 `json:"total"`
}

// BaseStandard it's ao final response
// which be return when called from apicall.Send()
type BaseStandard struct {
	Items 			  *json.RawMessage `json:"items"`
	AuditInfo         `json:"auditInfo"`
	InterfaceSettings interface{} `json:"interfaceSettings"`
	ctx 			  *context.Context
	cancel 			  *context.CancelFunc
}

// GetItems it transform delayed parsed json into structure provider
// e.g. response.GetItems(&MyStruct{})
// MyStruct[0].Foo, etc.
func (r *BaseStandard) GetItems(structType interface{}) (error) {
	items := r.Items
	return json.Unmarshal(*items, structType)
}

// IsOk return boolean, check if response was ok,
// which implies that response must be of range of 200 status code
// and items length > 0
func (r *BaseStandard) IsOk() bool {
	return r.Ok
}

// HasItems it will return true if
// BaseStandard.Items is not nil and len > 0
func (r *BaseStandard) HasItems() bool  {
	if r.Items == nil {
		return false
	}
	var genericItems interface{}
	err := json.Unmarshal(*r.Items, &genericItems)
	if err != nil {
		return false
	}

	return genericItems != nil && reflect.ValueOf(genericItems).Len() >0
}

func (r *BaseStandard) NewOperationId() (string, error) {
	h := md5.New()
	_, err := io.WriteString(h, r.AuditInfo.ClientIP)
	if err != nil {
		return "", err
	}
	_, err = io.WriteString(h, r.AuditInfo.Host)
	if err != nil {
		return "", err
	}
	_, err = io.WriteString(h, r.AuditInfo.Timestamp.String())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
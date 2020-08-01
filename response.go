package apicall

import (
	"encoding/json"
	"reflect"
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
	Duration float64 `json:"duration"`
	Host     string  `json:"host"`
	ClientIP string  `json:"clientIP"`
	Ok       bool    `json:"ok"`
	Errors   struct {
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
	InterfaceSettings []interface{} `json:"interfaceSettings"`
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
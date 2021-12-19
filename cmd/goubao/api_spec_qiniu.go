// SPDX-License-Identifier: MIT

package main

type goIdent struct {
	Pkg  string `json:"pkg"`
	Name string `json:"name"`
}

type goType struct {
	Ident   goIdent `json:"ident"`
	IsPtr   bool    `json:"is_ptr"`
	IsArray bool    `json:"is_array"`
}

type goMethod struct {
	Receiver *goType `json:"receiver"`
	Ident    goIdent `json:"ident"`
}

type routeDescOutput struct {
	Path       string   `json:"path"`
	Func       goMethod `json:"func"`
	HTTPMethod string   `json:"http_method"`
	ReqFormat  string   `json:"req_format"`
	ReqType    goType   `json:"req_type"`
	RespFormat string   `json:"resp_format"`
	RespType   goType   `json:"resp_type"`
}

package main

import (
	"log"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-resty/resty/v2"
)

// initJSVM - creates a new javascript virtual machine
func initJSVM(ctx map[string]interface{}) *goja.Runtime {
	vm := goja.New()
	for k, v := range ctx {
		vm.Set(k, v)
	}
	vm.Set("fetch", jsFetchfunc)
	vm.Set("log", log.Println)
	return vm
}

// jsFetchfunc - the fetch function used inside the js vm
func jsFetchfunc(url string, options ...map[string]interface{}) (map[string]interface{}, error) {
	var option map[string]interface{}
	var method string
	var headers map[string]string
	var body interface{}

	if len(options) > 0 {
		option = options[0]
	}

	if nil != option["method"] {
		method, _ = option["method"].(string)
	}

	if nil != option["headers"] {
		hdrs, _ := option["headers"].(map[string]interface{})
		headers = make(map[string]string)
		for k, v := range hdrs {
			headers[k], _ = v.(string)
		}
	}

	if nil != option["body"] {
		body, _ = option["body"]
	}
	client := resty.New()
	resp, err := client.R().SetHeaders(headers).SetBody(body).Execute(method, url)
	if err != nil {
		return nil, err
	}

	rspHdrs := resp.Header()
	rspHdrsNormalized := map[string]string{}
	for k, v := range rspHdrs {
		rspHdrsNormalized[strings.ToLower(k)] = v[0]
	}

	return map[string]interface{}{
		"status":     resp.Status(),
		"statusCode": resp.StatusCode(),
		"headers":    rspHdrsNormalized,
		"body":       string(resp.Body()),
	}, nil
}

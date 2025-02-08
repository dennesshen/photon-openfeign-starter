package openfeign

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	
	"github.com/dennesshen/photon-openfeign-starter/restTemplate"
)

func StartOpenFeign(ctx context.Context) error {
	for _, feignClient := range GetFeignClients() {
		fcType := reflect.TypeOf(feignClient).Elem()
		fcValue := reflect.ValueOf(feignClient).Elem()
		
		for i := range fcType.NumField() {
			field := fcType.Field(i)
			valueField := fcValue.Field(i)
			if field.Type.Kind() != reflect.Func {
				continue
			}
			
			apiPath := field.Tag.Get("path")
			if apiPath == "" {
				continue
			}
			method := field.Tag.Get("method")
			if method == "" {
				continue
			}
			
			newFunc := reflect.MakeFunc(field.Type, func(args []reflect.Value) []reflect.Value {
				var err error
				executor := restTemplate.NewExecutor()
				
				fullUrl, err := assembleUrl(feignClient.Domain(), apiPath)
				if err != nil {
					return setErrorValue(&field, err)
				}
				
				err = dealInputArgs(args, fullUrl, executor)
				if err != nil {
					return setErrorValue(&field, err)
				}
				executor.SetMethod(method)
				response, err := executor.Execute()
				if err != nil {
					return setErrorValue(&field, err)
				}
				
				return setOutValue(&field, &response, err)
			})
			valueField.Set(newFunc)
		}
		
	}
	
	return nil
}

func dealInputArgs(args []reflect.Value, fullUrl *url.URL, executor *restTemplate.Executor) (err error) {
	for _, value := range args {
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		switch value.Interface().(type) {
		case context.Context:
			executor.SetContext(value.Interface().(context.Context))
		case RequestBody:
			resBody := value.Interface().(RequestBody) //nolint:errcheck
			executor.AddHeader("Content-Type", resBody.ContentType())
			var encode []byte
			encode, err = resBody.Encode()
			executor.SetBody(bytes.NewBuffer(encode))
		case RequestHeaders:
			rh := value.Interface().(RequestHeaders) //nolint:errcheck
			executor.SetHeaders(rh)
		case RequestParam:
			rp := value.Interface().(RequestParam) //nolint:errcheck
			fullUrl.RawQuery = (url.Values)(rp).Encode()
		case PathVariable:
			pv := value.Interface().(PathVariable) //nolint:errcheck
			fullUrl, err = replacePathVariables(fullUrl, pv)
		}
		
		if err != nil {
			return
		}
	}
	
	// 檢查 URL 是否合法
	if fullUrl.Scheme != "http" && fullUrl.Scheme != "https" {
		return fmt.Errorf("invalid url: %s", fullUrl.String())
	}
	executor.SetUrl(fullUrl.String())
	
	return
}

func replacePathVariables(fullUrl *url.URL, variables map[string]string) (*url.URL, error) {
	// 用 map 中的值替換 path 變量
	path := fullUrl.Path
	for key, value := range variables {
		// 替換形如 {key} 的占位符
		placeholder := "{" + key + "}"
		path = strings.Replace(path, placeholder, value, -1)
	}
	
	// 返回替換後的完整路徑
	parseUrl, err := fullUrl.Parse(path)
	if err != nil {
		return nil, err
	}
	return parseUrl, nil
}

func assembleUrl(domain string, path string) (*url.URL, error) {
	fullUrl := domain + path
	parseUrl, err := url.Parse(fullUrl)
	if err != nil {
		return nil, err
	}
	return parseUrl, nil
}

func setErrorValue(field *reflect.StructField, err error) []reflect.Value {
	returnValues := make([]reflect.Value, field.Type.NumOut())
	for i := 0; i < field.Type.NumOut(); i++ {
		outType := field.Type.Out(i)
		if outType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			returnValues[i] = reflect.ValueOf(err)
		} else {
			returnValues[i] = reflect.Zero(outType)
		}
	}
	return returnValues
}

func setOutValue(field *reflect.StructField, response *restTemplate.RawResponse, err error) []reflect.Value {
	returnValues := make([]reflect.Value, field.Type.NumOut())
	
	outErrIndex := 0
	for i := 0; i < field.Type.NumOut(); i++ {
		outType := field.Type.Out(i)
		if outType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			outErrIndex = i
		} else if outType.Kind() == reflect.Struct {
			if outType == reflect.TypeOf((*restTemplate.RawResponse)(nil)).Elem() {
				returnValues[i] = reflect.ValueOf(*response)
			} else {
				valuePtr := reflect.New(outType)
				err = json.Unmarshal(response.RawResBody, valuePtr.Interface())
				returnValues[i] = valuePtr.Elem()
			}
		} else {
			if outType == reflect.TypeOf((*restTemplate.Status)(nil)).Elem() {
				returnValues[i] = reflect.ValueOf(restTemplate.Status(response.Status))
			} else if outType == reflect.TypeOf((*restTemplate.Header)(nil)).Elem() {
				returnValues[i] = reflect.ValueOf(restTemplate.Header(response.Headers))
			} else {
				returnValues[i] = reflect.Zero(outType)
			}
		}
	}
	if err != nil {
		returnValues[outErrIndex] = reflect.ValueOf(err)
	} else {
		returnValues[outErrIndex] = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
	}
	
	return returnValues
}

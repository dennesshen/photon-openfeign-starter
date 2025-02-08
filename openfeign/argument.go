package openfeign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	
	"github.com/dennesshen/photon-core-starter/utils/structure"
)

type RequestBody interface {
	Encode() ([]byte, error)
	ContentType() string
}

type JsonBody struct {
	data interface{}
}

func (jb *JsonBody) AddData(data interface{}) *JsonBody {
	jb.data = data
	return jb
}

func (jb *JsonBody) ContentType() string {
	return "application/json"
}

func (jb *JsonBody) Encode() ([]byte, error) {
	return json.Marshal(jb.data)
}

type MultiPartBody struct {
	data     map[string]interface{}
	filePart struct {
		fieldName string
		file      *os.File
	}
}

func (mpb *MultiPartBody) ContentType() string {
	return "multipart/form-data"
}

func (mpb *MultiPartBody) AddField(key string, value interface{}) *MultiPartBody {
	if mpb.data == nil {
		mpb.data = make(map[string]interface{})
	}
	mpb.data[key] = value
	return mpb
}

func (mpb *MultiPartBody) AddFile(fieldName string, file *os.File) *MultiPartBody {
	mpb.filePart.fieldName = fieldName
	mpb.filePart.file = file
	return mpb
}

func (mpb *MultiPartBody) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	for key, value := range mpb.data {
		err := writer.WriteField(key, fmt.Sprintf("%v", value))
		if err != nil {
			return nil, err
		}
	}
	
	if mpb.filePart.file != nil {
		// 創建表單文件字段
		part, err := writer.CreateFormFile(mpb.filePart.fieldName, mpb.filePart.file.Name())
		if err != nil {
			return nil, err
		}
		
		// 將文件內容寫入到 multipart 表單中
		_, err = io.Copy(part, mpb.filePart.file)
		if err != nil {
			return nil, err
		}
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type FormDataBody struct {
	keyValues structure.KVList
}

func (fbd *FormDataBody) ContentType() string {
	return "application/x-www-form-urlencoded"
}

func (fbd *FormDataBody) AddField(key, value string) *FormDataBody {
	if fbd.keyValues == nil {
		fbd.keyValues = make(structure.KVList, 0)
	}
	fbd.keyValues = append(fbd.keyValues, structure.KV{Key: key, Value: value})
	return fbd
}

func (fbd *FormDataBody) AddFieldByKV(kvList ...structure.KV) *FormDataBody {
	for _, kv := range kvList {
		fbd.AddField(kv.Key, kv.Value)
	}
	return fbd
}

func (fbd *FormDataBody) Encode() ([]byte, error) {
	return []byte(fbd.keyValues.Encode()), nil
}

type RequestHeaders map[string][]string

func (rh RequestHeaders) AddHeader(key string, values ...string) RequestHeaders {
	rh[key] = values
	return rh
}

type RequestParam url.Values

func (rp RequestParam) AddParam(key, value string) RequestParam {
	(url.Values)(rp).Add(key, value)
	return rp
}

func (rp RequestParam) AddParamByKv(kvList ...structure.KV) RequestParam {
	for _, kv := range kvList {
		rp.AddParam(kv.Key, kv.Value)
	}
	return rp
}

type PathVariable map[string]string

func (pv PathVariable) AddVariable(key, value string) PathVariable {
	pv[key] = value
	return pv
}

func (pv PathVariable) AddVarByKV(kvList ...structure.KV) PathVariable {
	for _, kv := range kvList {
		pv.AddVariable(kv.Key, kv.Value)
	}
	return pv
}

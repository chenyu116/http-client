/*
   Copyright Roger

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package http_client

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"

	goquery "github.com/google/go-querystring/query"
)

type Body interface {
	Encode(receiver io.ReadWriter) (string, error)
}

// bodyOriginal provides the wrapped body value as a Body for requests.
func NewOriginalBody(body io.Reader) Body {
	return bodyOriginal{body: body}
}

type bodyOriginal struct {
	body io.Reader
}

func (p bodyOriginal) Encode(receiver io.ReadWriter) (string, error) {
	_, err := io.Copy(receiver, p.body)
	if err != nil {
		return "", err
	}
	return "", nil
}

// jsonBodyProvider encodes a JSON tagged struct value as a Body for requests.
// See https://golang.org/pkg/encoding/json/#MarshalIndent for details.
func NewJsonBody(body interface{}, escapeHTML ...bool) Body {
	esH := false
	if len(escapeHTML) > 0 && escapeHTML[0] {
		esH = true
	}

	return bodyProviderJson{body: body, escapeHTML: esH}
}

type bodyProviderJson struct {
	body       interface{}
	escapeHTML bool
}

func (p bodyProviderJson) Encode(receiver io.ReadWriter) (string, error) {
	encoder := json.NewEncoder(receiver)
	encoder.SetEscapeHTML(p.escapeHTML)
	err := encoder.Encode(p.body)
	if err != nil {
		return "", err
	}
	return ContentTypeJson, nil
}

// formBodyProvider encodes a url tagged struct value as Body for requests.
// See https://godoc.org/github.com/google/go-querystring/query for details.
func NewFormBody(body interface{}) Body {
	return bodyProviderForm{body: body}
}

type bodyProviderForm struct {
	body interface{}
}

func (p bodyProviderForm) Encode(receiver io.ReadWriter) (string, error) {
	values, err := goquery.Values(p.body)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(receiver, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	return ContentTypeForm, nil
}

func NewFileBody(fileName, fieldName string, body ...interface{}) Body {
	bpf := bodyProviderFile{fileName: fileName, fieldName: fieldName}
	if len(body) > 0 {
		bpf.body = body[0]
	}
	return bpf
}

type bodyProviderFile struct {
	body      interface{}
	fileName  string
	fieldName string
}

func (p bodyProviderFile) Encode(receiver io.ReadWriter) (string, error) {
	if p.fileName == "" {
		return "", fmt.Errorf("%s not defined", "fileName")
	}
	if p.fieldName == "" {
		return "", fmt.Errorf("%s not defined", "fieldName")
	}
	file, err := os.OpenFile(p.fieldName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer file.Close()
	writer := multipart.NewWriter(receiver)
	defer func() func() {
		return func() {
			_ = writer.Close()
		}
	}()()

	fw, err := writer.CreateFormFile(p.fieldName, p.fileName)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(fw, file)
	if err != nil {
		return "", err
	}

	if p.body != nil {
		values, err := goquery.Values(p.body)
		if err != nil {
			return "", err
		}
		for k, _ := range values {
			err = writer.WriteField(k, values.Get(k))
			if err != nil {
				return "", err
			}
		}
	}

	return writer.FormDataContentType(), nil
}

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
	"bytes"
	"testing"
)

type TestParams struct {
	Name  string `url:"name"`
	Count int    `url:"count"`
}

type TestBody struct {
	Code      string `json:"code"`
	Token     string `json:"token"`
}

var params = TestParams{Name: "recent", Count: 25}

func TestNew(t *testing.T) {
	h := New()
	if h.header == nil {
		t.Errorf("Header map not initialized with make")
	}
}

func TestProxy(t *testing.T) {
	config := NewConfig()
	config.UseProxy = true
	config.ProxyHost = "http://127.0.0.1:12333"
	code, err := New(config).BaseURL("https://golang.org").Send()
	if err != nil {
		t.Errorf("expected %v", err)
	}
	if code >= 400 {
		t.Errorf("response code: %d", code)
	}
}

func TestRequest_query(t *testing.T) {
	res := new(bytes.Buffer)
	code, err := New().Get("http://example.com").SetBasicAuth("test", "test").Send(res)
	if err != nil {
		t.Error(err)
	}

	if code >= 400 {
		t.Errorf("response code: %d", code)
	}
}

func TestRequest_json(t *testing.T) {
	j := TestBody{
		Code:      "code",
		Token:     "token",
	}
	config := NewConfig()
	config.Retry.CheckStatusCode = false
	res := new(bytes.Buffer)
	code, err := New(config).Post("http://example.com").SetBody(NewJsonBody(j)).Send(res)
	if err != nil {
		t.Error(err)
	}

	if code >= 400 {
		t.Errorf("response code: %d", code)
	}

	t.Fatal(res.String())
}

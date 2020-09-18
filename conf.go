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
	"crypto/tls"
	"time"
)

// HTTPTimeout http timeout
type HTTPTimeout struct {
	ConnectTimeout        time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	ResponseHeaderTimeout time.Duration
	MaxTimeout            time.Duration
	TLSHandshakeTimeout   time.Duration
}

type Retry struct {
	Times int
	// if set to true , the status code greater than 400 will resend request
	// if set to false, the client only check the request error
	CheckStatusCode bool
}

// Config configure
type Config struct {
	HTTPTimeout     HTTPTimeout // HTTP的超时时间设置
	UseProxy        bool        // 是否使用代理
	ProxyHost       string      // 代理服务器地址
	IsAuthProxy     bool        // 代理服务器是否使用用户认证
	ProxyUser       string      // 代理服务器认证用户名
	ProxyPassword   string      // 代理服务器认证密码
	ReUseTCP        bool        // 为同一地址多次请求复用TCP连接
	Retry           Retry       // 重试设置
	TLSClientConfig *tls.Config // tls config
}

// 获取默认配置
func NewConfig() Config {
	return Config{
		HTTPTimeout: HTTPTimeout{
			ConnectTimeout:        time.Second * 3,
			ReadTimeout:           time.Second * 3,
			WriteTimeout:          time.Second * 3,
			ResponseHeaderTimeout: time.Second * 5,
			MaxTimeout:            time.Second * 300,
			TLSHandshakeTimeout:   time.Second * 5,
		},
		UseProxy:      false,
		ProxyHost:     "",
		IsAuthProxy:   false,
		ProxyUser:     "",
		ProxyPassword: "",
		Retry: Retry{
			Times:           3,
			CheckStatusCode: true,
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

package http_client

import (
	"bytes"
	"context"
	"encoding/base64"
	goquery "github.com/google/go-querystring/query"
	"io"
	"net"
	"net/http"
	"net/url"
)

func New(cfg ...Config) *HttpClient {
	config := NewConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(network, addr, config.HTTPTimeout.ConnectTimeout)
			if err != nil {
				return nil, err
			}
			return newTimeoutConn(conn, config.HTTPTimeout), nil
		},
		TLSHandshakeTimeout:   config.HTTPTimeout.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.HTTPTimeout.ResponseHeaderTimeout,
		TLSClientConfig:       config.TLSClientConfig,
	}
	// Proxy
	if config.UseProxy {
		proxyURL, err := url.Parse(config.ProxyHost)
		if err == nil && proxyURL != nil {
			if config.IsAuthProxy {
				proxyURL.User = url.UserPassword(config.ProxyUser, config.ProxyPassword)
			}
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return &HttpClient{
		client: &http.Client{Transport: transport},
		config: config,
		header: make(http.Header),
	}
}

func (h *HttpClient) New() *HttpClient {
	// copy Headers pairs into new Header map
	headerCopy := make(http.Header)
	for k, v := range h.header {
		headerCopy[k] = v
	}
	return &HttpClient{
		client: h.client,
		method: h.method,
		rawURL: h.rawURL,
		header: headerCopy,
		query:  append([]interface{}{}, h.query...),
		body:   h.body,
	}
}

type HttpClient struct {
	client *http.Client

	config Config

	method string

	rawURL string
	// stores key-values pairs to add to Rattle's Headers
	header http.Header
	// url tagged query structs
	query []interface{}
	// body provider
	body Body
	// http.Response
	resp *http.Response
}

func (h *HttpClient) setMethod(method string) {
	h.method = method
}

func (h *HttpClient) setPath(path string) *HttpClient {
	hostURL, hostErr := url.Parse(h.rawURL)
	pathURL, pathErr := url.Parse(path)
	if hostErr == nil && pathErr == nil {
		h.rawURL = hostURL.ResolveReference(pathURL).String()
	}
	return h
}

func (h *HttpClient) BaseURL(rawUrl string) *HttpClient {
	h.rawURL = rawUrl
	return h
}

// Head sets the Request method to HEAD and sets the given pathURL.
func (h *HttpClient) Head(pathURL string) *HttpClient {
	h.setMethod(HEAD)
	return h.setPath(pathURL)
}

// Get sets the Request method to GET and sets the given pathURL.
func (h *HttpClient) Get(pathURL string) *HttpClient {
	h.setMethod(GET)
	return h.setPath(pathURL)
}

// Post sets the Request method to POST and sets the given pathURL.
func (h *HttpClient) Post(pathURL string) *HttpClient {
	h.setMethod(POST)
	return h.setPath(pathURL)
}

// Put sets the Request method to PUT and sets the given pathURL.
func (h *HttpClient) Put(pathURL string) *HttpClient {
	h.setMethod(PUT)
	return h.setPath(pathURL)
}

// Patch sets the Request method to PATCH and sets the given pathURL.
func (h *HttpClient) Patch(pathURL string) *HttpClient {
	h.setMethod(PATCH)
	return h.setPath(pathURL)
}

// Delete sets the Sling method to DELETE and sets the given pathURL.
func (h *HttpClient) Delete(pathURL string) *HttpClient {
	h.setMethod(DELETE)
	return h.setPath(pathURL)
}

// Options sets the Sling method to OPTIONS and sets the given pathURL.
func (h *HttpClient) Options(pathURL string) *HttpClient {
	h.setMethod(OPTIONS)
	return h.setPath(pathURL)
}

// SetHeader sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalized.
func (h *HttpClient) SetHeader(key, value string) *HttpClient {
	h.header.Set(key, value)
	return h
}

// SetBasicAuth sets the Authorization header to use HTTP Basic Authentication
// with the provided username and password. With HTTP Basic Authentication
// the provided username and password are not encrypted.
func (h *HttpClient) SetBasicAuth(username, password string) *HttpClient {
	return h.SetHeader("Authorization", "Basic "+genBasicAuth(username, password))
}

// genBasicAuth returns the Host64 encoded username:password for basic auth copied
// from net/http.
func genBasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (h *HttpClient) SetBody(body Body) *HttpClient {
	h.body = body
	return h
}

// GetRequest returns a new http.Request created with the request properties.
// Returns any errors parsing the rawURL, encoding query structs, encoding
// the body, or creating the http.Request.
func (h *HttpClient) genRequest() (*http.Request, error) {
	reqURL, err := url.Parse(h.rawURL)
	if err != nil {
		return nil, err
	}

	err = genQuery(reqURL, h.query)
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	var reqContentType string
	if h.body != nil {
		reqContentType, err = h.body.Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(h.method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}

	setHeaders(req, h.header)
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.119 Safari/537.36")
	}

	if reqContentType != "" {
		req.Header.Set(ContentType, reqContentType)
	} else {
		req.Header.Del(ContentType)
	}

	return req, err
}

// AddQuery add queries for GET request
func (h *HttpClient) AddQueries(params ...interface{}) *HttpClient {
	if len(params) > 0 {
		h.query = append(h.query, params...)
	}
	return h
}

// genQuery parses url tagged query structs using go-querystring to
// encode them to url.Values and format them onto the url.RawQuery. Any
// query parsing or encoding errors are returned.
func genQuery(reqURL *url.URL, params []interface{}) error {
	urlValues, err := url.ParseQuery(reqURL.RawQuery)
	if err != nil {
		return err
	}
	// encodes query structs into a url.Values map and merges maps
	for _, param := range params {
		queryValues, err := goquery.Values(param)
		if err != nil {
			return err
		}
		for key, values := range queryValues {
			for _, value := range values {
				urlValues.Add(key, value)
			}
		}
	}
	// url.Values format to a sorted "url encoded" string, e.g. "key=val&foo=bar"
	reqURL.RawQuery = urlValues.Encode()
	return nil
}

// setHeaders adds the key, value pairs from the given http.Header to the
// Rattle. Values for existing keys are appended to the keys values.
func setHeaders(req *http.Request, headers http.Header) {
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

func (h *HttpClient) Send(receiver ...io.ReadWriter) (statusCode int, err error) {
	req, err := h.genRequest()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer func() func() {
		return func() {
			req.Close = true
		}
	}()
	h.resp, err = h.client.Do(req)
	if err != nil {
		if h.resp != nil {
			return h.resp.StatusCode, err
		}
		return http.StatusBadRequest, err
	}
	h.resp.Close = true
	defer func() func() {
		return func() {
			_ = h.resp.Body.Close()
		}
	}()()

	if len(receiver) > 0 {
		_, err = io.Copy(receiver[0], h.resp.Body)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return h.resp.StatusCode, nil
}

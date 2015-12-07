package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	libraryVersion   = "0.1"
	userAgent        = "go-confluence/" + libraryVersion
	defaultMediaType = "application/json"
)

// A Client manages communication with the Confluence API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// The BasicAuth username
	username string

	// The BasicAuth password
	password string

	// Base URL for API requests.
	RestURL *url.URL

	// User agent used when communicating with the Confluence API.
	UserAgent string

	// Services used for managing Confluence Content
	// https://docs.atlassian.com/confluence/REST/latest/#d3e1011
	Content *ContentService
}

// NewClient returns a new Confluence API client. It uses basic auth so needs a
// username and password.
func NewClient(username string, password string, serverUrl string) *Client {
	baseURL, _ := url.Parse(serverUrl)

	c := &Client{
		client:    http.DefaultClient,
		username:  username,
		password:  password,
		RestURL:   baseURL,
		UserAgent: userAgent,
	}

	c.Content = &ContentService{client: c}
	return c
}

func (c *Client) baseRequest(method, urlStr string, baseURL url.URL, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := baseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", defaultMediaType)
	req.Header.Add("User-Agent", c.UserAgent)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)
	return req, nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the RestURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	return c.baseRequest(method, urlStr, *c.RestURL, body)
}

// Do sends an API request and returns the API response. The API response is
// decoded and stored in the value pointed to by v, or returned as an error if
// an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}

// An ErrorResponse reports the errors caused by an API request.
//
type ErrorResponse struct {
	Response *http.Response // HTTP response
	Data     []byte         // the error details
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %s",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Data)
}

// CheckResponse checks the API response for errors, and returns them if
// present.  A response is considered an error if it has a status code outside
// the 200 range.  API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse.  Any other
// response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorResponse.Data = data
	}
	return errorResponse
}

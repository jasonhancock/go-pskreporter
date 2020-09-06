package pskreporter

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const queryURL = "https://retrieve.pskreporter.info/query"

// Query executes a search query against the PSK Reporter API.
func Query(opts ...QueryOption) (*Response, error) {
	u, err := url.Parse(queryURL)
	if err != nil {
		return nil, err
	}

	o := options{
		vals: u.Query(),
	}

	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	u.RawQuery = o.vals.Encode()
	log.Println(u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Println(resp.StatusCode)

	var r Response
	if err := xml.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

type options struct {
	vals url.Values
}

// QueryOption is used to customize the query.
type QueryOption func(*options) error

// WithSenderCallsign set the sender callsign in the query.
func WithSenderCallsign(s string) QueryOption {
	return func(o *options) error {
		o.vals.Set("senderCallsign", s)
		return nil
	}
}

// WithFlowStartSeconds is the negative number of seconds to indicate how much
// data to retreive. This cannot be more than 24 hours.
func WithFlowStartSeconds(s int) QueryOption {
	return func(o *options) error {
		if s > 0 {
			return errors.New("WithFlowStartSeconds must be negative")
		}

		if s < -86400 {
			return errors.New("WithFlowStartSeconds cannot be greater than 24 hours")
		}

		o.vals.Set("flowStartSeconds", fmt.Sprintf("%d", s))
		return nil
	}
}

// WithNoActive will not return the active monitors if non zero.
func WithNoActive(i int) QueryOption {
	return func(o *options) error {
		o.vals.Set("noactive", fmt.Sprintf("%d", i))
		return nil
	}
}

package pskreporter

import (
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const queryURL = "https://retrieve.pskreporter.info/query"

// Doer is an interface the http.Client conforms to.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a client that will communicate with the PSKReporter service.
type Client struct {
	doer          Doer
	baseURL       string
	cacheDir      string
	cacheDuration time.Duration
}

// WithHTTPClient set the http client to use.
func WithHTTPClient(c Doer) ClientOption {
	return func(o *clientOptions) error {
		o.doer = c
		return nil
	}
}

// WithBaseURL set the API base url to use.
func WithBaseURL(s string) ClientOption {
	return func(o *clientOptions) error {
		if _, err := url.Parse(s); err != nil {
			return err
		}
		o.baseURL = s
		return nil
	}
}

// WithCacheDir will turn on caching. Directory must already exist.
func WithCacheDir(dir string) ClientOption {
	return func(o *clientOptions) error {
		o.cacheDir = dir
		return nil
	}
}

// WithCacheDuration determines how long a result will be served out of the cache
// before fetching a new one.
func WithCacheDuration(dur time.Duration) ClientOption {
	return func(o *clientOptions) error {
		if dur < 0 {
			return errors.New("cache duration must be positive")
		}
		o.cacheDuration = dur
		return nil
	}
}

// New instantiates a new Client.
func New(opts ...ClientOption) (*Client, error) {
	o := &clientOptions{
		doer:          http.DefaultClient,
		baseURL:       queryURL,
		cacheDuration: 280 * time.Second,
	}

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	return &Client{
		doer:          o.doer,
		baseURL:       o.baseURL,
		cacheDir:      o.cacheDir,
		cacheDuration: o.cacheDuration,
	}, nil
}

type clientOptions struct {
	doer          Doer
	baseURL       string
	cacheDir      string
	cacheDuration time.Duration
}

// ClientOption is used to customize the client.
type ClientOption func(*clientOptions) error

// Query executes a search query against the PSK Reporter API.
func (c *Client) Query(opts ...QueryOption) (*Response, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	o := queryOptions{
		vals: u.Query(),
	}

	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	u.RawQuery = o.vals.Encode()

	if c.cacheDir != "" {
		file := filepath.Join(c.cacheDir, hash(u.RawQuery))
		if fi, err := os.Stat(file); err == nil {
			if fi.ModTime().After(time.Now().Add(-1 * c.cacheDuration)) {
				fh, err := os.Open(file)
				if err != nil {
					return nil, fmt.Errorf("opening cached file: %w", err)
				}
				defer fh.Close()
				var r Response
				if err := xml.NewDecoder(fh).Decode(&r); err == nil {
					return &r, nil
				}
				// If we're here, there was an error, with the cached result, so go ahead and
				// make the request.
			}
		}
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http response %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var r Response
	if err := xml.Unmarshal(b, &r); err != nil {
		return nil, err
	}

	if c.cacheDir != "" {
		file := filepath.Join(c.cacheDir, hash(u.RawQuery))
		fh, err := os.Create(file)
		if err == nil {
			defer fh.Close()
			fh.Write(b)
		}
	}

	return &r, nil
}

type queryOptions struct {
	vals url.Values
}

// QueryOption is used to customize the query.
type QueryOption func(*queryOptions) error

// WithSenderCallsign set the sender callsign in the query.
func WithSenderCallsign(s string) QueryOption {
	return func(o *queryOptions) error {
		if _, ok := o.vals["receiverCallsign"]; ok {
			return errCallsignExclusive
		}
		if _, ok := o.vals["callsign"]; ok {
			return errCallsignExclusive
		}
		o.vals.Set("senderCallsign", s)
		return nil
	}
}

// WithReceiverCallsign set the receiver callsign in the query.
func WithReceiverCallsign(s string) QueryOption {
	return func(o *queryOptions) error {
		if _, ok := o.vals["senderCallsign"]; ok {
			return errCallsignExclusive
		}
		if _, ok := o.vals["callsign"]; ok {
			return errCallsignExclusive
		}
		o.vals.Set("receiverCallsign", s)
		return nil
	}
}

var errCallsignExclusive = errors.New("only one of callsign, senderCallsign, or receiverCallsign can be specified at a time")

// WithCallsign sets the Callsign of interest.
func WithCallsign(s string) QueryOption {
	return func(o *queryOptions) error {
		if _, ok := o.vals["senderCallsign"]; ok {
			return errCallsignExclusive
		}
		if _, ok := o.vals["receiverCallsign"]; ok {
			return errCallsignExclusive
		}

		o.vals.Set("callsign", s)
		return nil
	}
}

// WithMode sets the mode of operation in the query.
func WithMode(s string) QueryOption {
	return func(o *queryOptions) error {
		o.vals.Set("mode", s)
		return nil
	}
}

// WithReportLimit limits the number of records returned.
func WithReportLimit(s int) QueryOption {
	return func(o *queryOptions) error {
		o.vals.Set("rptlimit", fmt.Sprintf("%d", s))
		return nil
	}
}

var errFlowStartGreaterDay = errors.New("WithFlowStartSeconds cannot be greater than 24 hours")
var errFlowStartNotNegative = errors.New("WithFlowStartSeconds must be negative")

// WithFlowStartSeconds is the negative number of seconds to indicate how much
// data to retrieve. This cannot be more than 24 hours.
func WithFlowStartSeconds(s int) QueryOption {
	return func(o *queryOptions) error {
		if s > 0 {
			return errFlowStartNotNegative
		}

		if s < -86400 {
			return errFlowStartGreaterDay
		}

		o.vals.Set("flowStartSeconds", fmt.Sprintf("%d", s))
		return nil
	}
}

// WithNoActive will not return the active monitors if non zero.
func WithNoActive(i int) QueryOption {
	return func(o *queryOptions) error {
		o.vals.Set("noactive", fmt.Sprintf("%d", i))
		return nil
	}
}

// WithAppContact sets a contact email address in case the psk reporter folks want to get in touch.
func WithAppContact(email string) QueryOption {
	return func(o *queryOptions) error {
		o.vals.Set("appcontact", email)
		return nil
	}
}

// WithRROnly will only return the reception reports if non-zero.
func WithRROnly(i int) QueryOption {
	return func(o *queryOptions) error {
		o.vals.Set("rronly", fmt.Sprintf("%d", i))
		return nil
	}
}

var errLowerFrequencyGreaterThanUpper = errors.New("lower frequency must be less than upper frequency")

// WithFrequencyRange sets a lower and upper bound for frequencies. Example: 14000000-14100000
func WithFrequencyRange(lower, upper int) QueryOption {
	return func(o *queryOptions) error {
		if lower > upper {
			return errLowerFrequencyGreaterThanUpper
		}
		o.vals.Set("frange", fmt.Sprintf("%d-%d", lower, upper))
		return nil
	}
}

// WithNoLocator will return reception reports without a locator if non-zero.
func WithNoLocator(i int) QueryOption {
	return func(o *queryOptions) error {
		o.vals.Set("nolocator", fmt.Sprintf("%d", i))
		return nil
	}
}

// WithStatistics will return some statistics if non-zero.
func WithStatistics(i int) QueryOption {
	return func(o *queryOptions) error {
		o.vals.Set("statistics", fmt.Sprintf("%d", i))
		return nil
	}
}

// WithLastSequenceNumber sets the last sequence number in the query.
func WithLastSequenceNumber(s string) QueryOption {
	return func(o *queryOptions) error {
		o.vals.Set("lastseqno", s)
		return nil
	}
}

// hash computes the md5 checksum of the given strings
func hash(args ...string) string {
	h := md5.New()
	for _, arg := range args {
		io.WriteString(h, arg)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

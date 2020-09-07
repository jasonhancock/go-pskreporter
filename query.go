package pskreporter

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
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

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected http response %d", resp.StatusCode)
	}

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
	return func(o *options) error {
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
	return func(o *options) error {
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
	return func(o *options) error {
		o.vals.Set("mode", s)
		return nil
	}
}

// WithReportLimit limits the number of records returned.
func WithReportLimit(s int) QueryOption {
	return func(o *options) error {
		o.vals.Set("rptlimit", fmt.Sprintf("%d", s))
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

// WithAppContact sets a contact email address in case the psk reporter folks want to get in touch.
func WithAppContact(email string) QueryOption {
	return func(o *options) error {
		o.vals.Set("appcontact", email)
		return nil
	}
}

// WithRROnly will only return the reception reports if non-zero.
func WithRROnly(i int) QueryOption {
	return func(o *options) error {
		o.vals.Set("rronly", fmt.Sprintf("%d", i))
		return nil
	}
}

// WithFrequencyRange sets a lower and upper bound for frequencies. Example: 14000000-14100000
func WithFrequencyRange(lower, upper int) QueryOption {
	return func(o *options) error {
		if lower > upper {
			return errors.New("lower frequency must be less than upper frequency")
		}
		o.vals.Set("frange", fmt.Sprintf("%d-%d", lower, upper))
		return nil
	}
}

// WithNoLocator will return reception reports without a locator if non-zero.
func WithNoLocator(i int) QueryOption {
	return func(o *options) error {
		o.vals.Set("nolocator", fmt.Sprintf("%d", i))
		return nil
	}
}

// WithStatistics will return some statistics if non-zero.
func WithStatistics(i int) QueryOption {
	return func(o *options) error {
		o.vals.Set("statistics", fmt.Sprintf("%d", i))
		return nil
	}
}

// WithLastSequenceNumber sets the last sequence number in the query.
func WithLastSequenceNumber(s string) QueryOption {
	return func(o *options) error {
		o.vals.Set("lastseqno", s)
		return nil
	}
}

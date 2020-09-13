package pskreporter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
			fh, err := os.Open("testdata/output.xml")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer fh.Close()

			io.Copy(w, fh)
		})

		svr := httptest.NewServer(mux)
		defer svr.Close()

		c, err := New(WithBaseURL(svr.URL + "/foo"))
		require.NoError(t, err)

		resp, err := c.Query(WithCallsign("AG6K"))
		require.NoError(t, err)
		checkResponse(t, resp)
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("bad response", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("hello world"))
			})

			svr := httptest.NewServer(mux)
			defer svr.Close()

			c, err := New(WithBaseURL(svr.URL + "/foo"))
			require.NoError(t, err)

			_, err = c.Query(WithCallsign("AG6K"))
			require.Error(t, err)
		})

		t.Run("http 500", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})

			svr := httptest.NewServer(mux)
			defer svr.Close()

			c, err := New(WithBaseURL(svr.URL + "/foo"))
			require.NoError(t, err)

			_, err = c.Query(WithCallsign("AG6K"))
			require.Error(t, err)
			require.Contains(t, err.Error(), "unexpected http response")
		})

		t.Run("bad options", func(t *testing.T) {
			c, err := New()
			require.NoError(t, err)

			_, err = c.Query(WithCallsign("AG6K"), WithReceiverCallsign("AG6K"))
			require.Equal(t, errCallsignExclusive, err)
		})

		t.Run("bad client option", func(t *testing.T) {
			_, err := New(WithBaseURL("http://[::1]:namedport"))
			require.Error(t, err)
		})

		t.Run("doer error", func(t *testing.T) {

			c, err := New(WithHTTPClient(&doerError{}))
			require.NoError(t, err)

			_, err = c.Query()
			require.Error(t, err)
			require.Contains(t, err.Error(), "error")
		})
	})
}

type doerError struct{}

func (d *doerError) Do(*http.Request) (*http.Response, error) {
	return nil, errors.New("error")
}

func TestQueryOptions(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		tests := []struct {
			desc     string
			opts     []QueryOption
			expected map[string]string
		}{
			{
				"WithSenderCallsign",
				[]QueryOption{WithSenderCallsign("ABCD")},
				map[string]string{"senderCallsign": "ABCD"},
			},
			{
				"WithReceiverCallsign",
				[]QueryOption{WithReceiverCallsign("ABCD")},
				map[string]string{"receiverCallsign": "ABCD"},
			},
			{
				"WithCallsign",
				[]QueryOption{WithCallsign("ABCD")},
				map[string]string{"callsign": "ABCD"},
			},
			{
				"WithMode",
				[]QueryOption{WithMode("FT8")},
				map[string]string{"mode": "FT8"},
			},
			{
				"WithReportLimit",
				[]QueryOption{WithReportLimit(10)},
				map[string]string{"rptlimit": "10"},
			},
			{
				"WithFlowStartSeconds",
				[]QueryOption{WithFlowStartSeconds(-10)},
				map[string]string{"flowStartSeconds": "-10"},
			},
			{
				"WithAppContact",
				[]QueryOption{WithAppContact("foo@example.com")},
				map[string]string{"appcontact": "foo@example.com"},
			},
			{
				"WithFrequencyRange",
				[]QueryOption{WithFrequencyRange(123, 456)},
				map[string]string{"frange": "123-456"},
			},
			{
				"WithLastSequenceNumber",
				[]QueryOption{WithLastSequenceNumber("abc123")},
				map[string]string{"lastseqno": "abc123"},
			},
			{
				"WithNoActive",
				[]QueryOption{WithNoActive(2)},
				map[string]string{"noactive": "2"},
			},
			{
				"WithNoLocator",
				[]QueryOption{WithNoLocator(2)},
				map[string]string{"nolocator": "2"},
			},
			{
				"WithRROnly",
				[]QueryOption{WithRROnly(2)},
				map[string]string{"rronly": "2"},
			},
			{
				"WithStatistics",
				[]QueryOption{WithStatistics(2)},
				map[string]string{"statistics": "2"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.desc, func(t *testing.T) {
				o := queryOptions{
					vals: make(url.Values),
				}

				for _, opt := range tt.opts {
					require.NoError(t, opt(&o))
				}

				for k, v := range tt.expected {
					require.Equal(t, v, o.vals.Get(k))
				}
			})
		}
	})

	t.Run("errors", func(t *testing.T) {
		tests := []struct {
			desc string
			opt  QueryOption
			vals url.Values
			err  error
		}{
			{
				"WithFrequencyRange lower > upper",
				WithFrequencyRange(2, 1),
				nil,
				errLowerFrequencyGreaterThanUpper,
			},
			{
				"WithFlowStartSeconds not negative",
				WithFlowStartSeconds(1),
				nil,
				errFlowStartNotNegative,
			},
			{
				"WithFlowStartSeconds greater 1 day",
				WithFlowStartSeconds(-86401),
				nil,
				errFlowStartGreaterDay,
			},
			{
				"WithCallsign - senderCallsign set",
				WithCallsign("a"),
				url.Values{"senderCallsign": []string{"foo"}},
				errCallsignExclusive,
			},
			{
				"WithCallsign - receiverCallsign set",
				WithCallsign("a"),
				url.Values{"receiverCallsign": []string{"foo"}},
				errCallsignExclusive,
			},
			{
				"WithSenderCallsign - callsign set",
				WithSenderCallsign("a"),
				url.Values{"callsign": []string{"foo"}},
				errCallsignExclusive,
			},
			{
				"WithSenderCallsign - receiverCallsign set",
				WithSenderCallsign("a"),
				url.Values{"receiverCallsign": []string{"foo"}},
				errCallsignExclusive,
			},
			{
				"WithReceiverCallsign - callsign set",
				WithReceiverCallsign("a"),
				url.Values{"callsign": []string{"foo"}},
				errCallsignExclusive,
			},
			{
				"WithReceiverCallsign - senderCallsign set",
				WithReceiverCallsign("a"),
				url.Values{"senderCallsign": []string{"foo"}},
				errCallsignExclusive,
			},
		}

		for _, tt := range tests {
			t.Run(tt.desc, func(t *testing.T) {
				o := queryOptions{
					vals: tt.vals,
				}

				err := tt.opt(&o)
				require.Error(t, err)
				require.Equal(t, tt.err, err)
			})
		}
	})
}

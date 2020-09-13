package pskreporter

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsing(t *testing.T) {
	fh, err := os.Open("testdata/output.xml")
	require.NoError(t, err)
	defer fh.Close()

	var resp Response
	require.NoError(t, xml.NewDecoder(fh).Decode(&resp))

	checkResponse(t, &resp)
}

func checkResponse(t *testing.T, resp *Response) {
	require.Equal(t, "ag6k", resp.SenderSearch.Callsign)
	require.Equal(t, "1599163380", resp.SenderSearch.RecentFlowStartSeconds)

	require.Len(t, resp.ActiveCallsigns, 20)
	require.Equal(t, "R2PU", resp.ActiveCallsigns[0].Callsign)

	require.Len(t, resp.ReceptionReports, 340)
	require.Equal(t, "W5CJ", resp.ReceptionReports[0].ReceiverCallsign)

	require.Equal(t, "1599164931", resp.MaxFlowStartSeconds.Value)

	require.Equal(t, "14631964162", resp.LastSequenceNumber.Value)

	require.Equal(t, "1599164934", resp.CurrentSeconds)

	require.Len(t, resp.ActiveReceivers, 4695)
	require.Equal(t, "DL0046SWL", resp.ActiveReceivers[0].Callsign)
}

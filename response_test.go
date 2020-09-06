package pskreporter

import (
	"encoding/json"
	"encoding/xml"
	"log"
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
	PrettyPrint(resp)
}

// PrettyPrint some shit
func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "\t")
	log.Printf("\n%s\n\n", string(b))
}

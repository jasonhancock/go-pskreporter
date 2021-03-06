package pskreporter_test

import (
	"log"

	pskr "github.com/jasonhancock/go-pskreporter"
)

func Example() {
	c, err := pskr.New()
	if err != nil {
		log.Fatal(err)
	}

	// Query for the people that heard callsign AG6K over the last 30 minutes.
	r, err := c.Query(
		pskr.WithSenderCallsign("AG6K"),
		pskr.WithFlowStartSeconds(-1800),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("number of people that heard it: %d", len(r.ReceptionReports))
}

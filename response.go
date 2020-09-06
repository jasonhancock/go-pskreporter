package pskreporter

import (
	"encoding/xml"
)

// query response is defined here: https://pskreporter.info/pskdev.html

// Response is the response format from the API.
type Response struct {
	XMLName             xml.Name            `xml:"receptionReports"`
	Text                string              `xml:",chardata"`
	CurrentSeconds      string              `xml:"currentSeconds,attr"`
	ActiveReceivers     []ActiveReceiver    `xml:"activeReceiver"`
	LastSequenceNumber  LastSequenceNumber  `xml:"lastSequenceNumber"`
	MaxFlowStartSeconds MaxFlowStartSeconds `xml:"maxFlowStartSeconds"`
	ReceptionReports    []ReceptionReport   `xml:"receptionReport"`
	SenderSearch        SenderSearch        `xml:"senderSearch"`
	ActiveCallsigns     []ActiveCallsign    `xml:"activeCallsign"`
}

// ActiveCallsign represents an active call sign in the response.
type ActiveCallsign struct {
	Text      string `xml:",chardata"`
	Callsign  string `xml:"callsign,attr"`
	Reports   string `xml:"reports,attr"`
	DXCC      string `xml:"DXCC,attr"`
	DXCCcode  string `xml:"DXCCcode,attr"`
	Frequency string `xml:"frequency,attr"`
}

// ActiveReceiver represents an active receiver in the response.
type ActiveReceiver struct {
	Text               string `xml:",chardata"`
	Callsign           string `xml:"callsign,attr"`
	Locator            string `xml:"locator,attr"`
	Frequency          string `xml:"frequency,attr"`
	Region             string `xml:"region,attr"`
	DXCC               string `xml:"DXCC,attr"`
	DecoderSoftware    string `xml:"decoderSoftware,attr"`
	AntennaInformation string `xml:"antennaInformation,attr"`
	Mode               string `xml:"mode,attr"`
	Bands              string `xml:"bands,attr"`
}

// ReceptionReport represents a reception report in the response.
type ReceptionReport struct {
	Text             string `xml:",chardata"`
	ReceiverCallsign string `xml:"receiverCallsign,attr"`
	ReceiverLocator  string `xml:"receiverLocator,attr"`
	SenderCallsign   string `xml:"senderCallsign,attr"`
	SenderLocator    string `xml:"senderLocator,attr"`
	Frequency        string `xml:"frequency,attr"`
	FlowStartSeconds string `xml:"flowStartSeconds,attr"`
	Mode             string `xml:"mode,attr"`
	IsSender         string `xml:"isSender,attr"`
	ReceiverDXCC     string `xml:"receiverDXCC,attr"`
	ReceiverDXCCCode string `xml:"receiverDXCCCode,attr"`
	SNR              string `xml:"sNR,attr"`
}

// SenderSearch represents the sender search in the response.
type SenderSearch struct {
	Text                   string `xml:",chardata"`
	Callsign               string `xml:"callsign,attr"`
	RecentFlowStartSeconds string `xml:"recentFlowStartSeconds,attr"`
}

// MaxFlowStartSeconds represents the max flow start seconds in the response.
type MaxFlowStartSeconds struct {
	Text  string `xml:",chardata"`
	Value string `xml:"value,attr"`
}

// LastSequenceNumber is the last sequence number in the response.
type LastSequenceNumber struct {
	Text  string `xml:",chardata"`
	Value string `xml:"value,attr"`
}

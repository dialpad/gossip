package parser

import (
	"strings"
	"testing"
)

// Baseline with proper content header (pass, with our without fix for TEL-14941)
func Test420BadExtension1(t *testing.T) {
	input := []string{"SIP/2.0 420 Bad Extension",
		"Via: SIP/2.0/UDP 170.10.123.456;received=170.10.789.012;rport=5060;branch=z9hG4bK8Q120N45yjgcH",
		"From: \"1120416105 1120416105\" <sip:+1234567890123@10.1.1.1>;tag=HKK7gp16D0m3a",
		"To: <sip:+5422200001@200.3.3.3;user=phone>;tag=349268587-1629127825171-",
		"Call-ID: BW123025171160821-1042155598@200.5.5.5",
		"CSeq: 40010127 INVITE",
		"Content-Length: 0",
		"Unsupported: timer"}
	var err error
	_, err = ParseMessage([]byte(strings.Join(input, "\r\n")), false)

	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
}

// No content header (fail, pass after fix for TEL-14941)
func Test420BadExtension2(t *testing.T) {
	input := []string{"SIP/2.0 420 Bad Extension",
		"Via: SIP/2.0/UDP 170.10.123.456;received=170.10.789.012;rport=5060;branch=z9hG4bK8Q120N45yjgcH",
		"From: \"1120416105 1120416105\" <sip:+1234567890123@10.1.1.1>;tag=HKK7gp16D0m3a",
		"To: <sip:+5422200001@200.3.3.3;user=phone>;tag=349268587-1629127825171-",
		"Call-ID: BW123025171160821-1042155598@200.5.5.5",
		"CSeq: 40010127 INVITE",
		"Unsupported: timer"}
	var err error
	_, err = ParseMessage([]byte(strings.Join(input, "\r\n")), false)

	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
}

// Multiple content headers (fail, with our without fix for TEL-14941)
func Test420BadExtension3(t *testing.T) {
	input := []string{"SIP/2.0 420 Bad Extension",
		"Via: SIP/2.0/UDP 170.10.123.456;received=170.10.789.012;rport=5060;branch=z9hG4bK8Q120N45yjgcH",
		"From: \"1120416105 1120416105\" <sip:+1234567890123@10.1.1.1>;tag=HKK7gp16D0m3a",
		"To: <sip:+5422200001@200.3.3.3;user=phone>;tag=349268587-1629127825171-",
		"Call-ID: BW123025171160821-1042155598@200.5.5.5",
		"CSeq: 40010127 INVITE",
		"Content-Length: 1",
		"Unsupported: timer",
		"Content-Length: 2"}
	var err error
	_, err = ParseMessage([]byte(strings.Join(input, "\r\n")), false)

	if err == nil {
		t.Errorf("Expect error \"Multiple content-length headers on message SIP/2.0 420 BadExtension\"")
	}
}

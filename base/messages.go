package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// A representation of a SIP method.
// This is syntactic sugar around the string type, so make sure to use
// the Equals method rather than built-in equality, or you'll fall foul of case differences.
// If you're defining your own Method, uppercase is preferred but not compulsory.
type Method string

// Determine if the given method equals some other given method.
// This is syntactic sugar for case insensitive equality checking.
func (method *Method) Equals(other *Method) bool {
	if method != nil && other != nil {
		return strings.EqualFold(string(*method), string(*other))
	} else {
		return method == other
	}
}

// It's nicer to avoid using raw strings to represent methods, so the following standard
// method names are defined here as constants for convenience.
const (
	INVITE    Method = "INVITE"
	ACK       Method = "ACK"
	CANCEL    Method = "CANCEL"
	BYE       Method = "BYE"
	REGISTER  Method = "REGISTER"
	OPTIONS   Method = "OPTIONS"
	SUBSCRIBE Method = "SUBSCRIBE"
	NOTIFY    Method = "NOTIFY"
	REFER     Method = "REFER"
	INFO      Method = "INFO"
	UPDATE    Method = "UPDATE"
	PRACK     Method = "PRACK"
	PUBLISH   Method = "PUBLISH"
	MESSAGE   Method = "MESSAGE"
)

var Methods = map[string]Method{
	"INVITE":    INVITE,
	"ACK":       ACK,
	"CANCEL":    CANCEL,
	"BYE":       BYE,
	"REGISTER":  REGISTER,
	"OPTIONS":   OPTIONS,
	"SUBSCRIBE": SUBSCRIBE,
	"NOTIFY":    NOTIFY,
	"REFER":     REFER,
	"INFO":      INFO,
	"UPDATE":    UPDATE,
	"PRACK":     PRACK,
	"PUBLISH":   PUBLISH,
	"MESSAGE":   MESSAGE,
}

// Internal representation of a SIP message - either a Request or a Response.
type SipMessage interface {
	// Yields a flat, string representation of the SIP message suitable for sending out over the wire.
	String() string

	// Adds a header to this message.
	AddHeader(h GenericHeader)

	// Returns a slice of all headers of the given type.
	// If there are no headers of the requested type, returns an empty slice.
	Headers(name string) []GenericHeader

	// Return all headers attached to the message, as a slice.
	AllHeaders() []GenericHeader

	// Yields a short string representation of the message useful for logging.
	Short() string

	// Remove the specified header from the message.
	RemoveHeader(header GenericHeader) error

	// Get the body of the message, as a string.
	GetBody() string

	// Set the body of the message.
	SetBody(body string)

	// Returns the MarshalledJson
	MarshalJSON() ([]byte, error)
}

// A shared type for holding headers and their ordering.
type headers struct {
	// The logical SIP headers attached to this message.
	headers map[string][]GenericHeader

	// The order the headers should be displayed in.
	headerOrder []string
}

func newHeaders() (result headers) {
	result.headers = make(map[string][]GenericHeader)
	return result
}

func (h headers) String() string {
	buffer := bytes.Buffer{}
	// Construct each header in turn and add it to the message.
	for typeIdx, name := range h.headerOrder {
		headers := h.headers[name]
		for idx, header := range headers {
			buffer.WriteString(header.String())
			if typeIdx < len(h.headerOrder) || idx < len(headers) {
				buffer.WriteString("\r\n")
			}
		}
	}
	return buffer.String()
}

// Add the given header.
func (hs *headers) AddHeader(h GenericHeader) {
	if hs.headers == nil {
		hs.headers = map[string][]GenericHeader{}
		hs.headerOrder = []string{}
	}
	name := h.Name()
	if _, ok := hs.headers[name]; ok {
		hs.headers[name] = append(hs.headers[name], h)
	} else {
		hs.headers[name] = []GenericHeader{h}
		hs.headerOrder = append(hs.headerOrder, name)
	}
}

// Gets some headers.
func (hs *headers) Headers(name string) []GenericHeader {
	if hs.headers == nil {
		hs.headers = map[string][]GenericHeader{}
		hs.headerOrder = []string{}
	}
	if headers, ok := hs.headers[name]; ok {
		return headers
	} else {
		return []GenericHeader{}
	}
}

// A SIP request (c.f. RFC 3261 section 7.1).
type Request struct {
	// Which method this request is, e.g. an INVITE or a REGISTER.
	Method Method

	// The Request URI. This indicates the user to whom this request is being addressed.
	Recipient Uri

	// The version of SIP used in this message, e.g. "SIP/2.0".
	SipVersion string

	// A Request has headers.
	headers

	// The order the headers should be displayed in.
	headerOrder []string

	// The application data of the message.
	Body string
}

func NewRequest(method Method, recipient Uri, sipVersion string, headers []GenericHeader, body string) (request *Request) {
	request = new(Request)
	request.Method = method
	request.SipVersion = sipVersion
	request.Recipient = recipient
	request.headers = newHeaders()
	request.Body = body

	for _, header := range headers {
		request.AddHeader(header)
	}

	return
}

func (request *Request) String() string {
	var buffer bytes.Buffer

	// Every SIP request starts with a Request Line - RFC 2361 7.1.
	buffer.WriteString(fmt.Sprintf("%s %s %s\r\n",
		(string)(request.Method),
		request.Recipient.String(),
		request.SipVersion))

	buffer.WriteString(request.headers.String())

	// If the request has a message body, add it.
	buffer.WriteString("\r\n" + request.Body)

	return buffer.String()
}

func (request *Request) Short() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%s %s %s",
		(string)(request.Method),
		request.Recipient.String(),
		request.SipVersion))

	return buffer.String()
}

func (request *Request) AllHeaders() []GenericHeader {
	allHeaders := make([]GenericHeader, 0)
	for _, key := range request.headers.headerOrder {
		allHeaders = append(allHeaders, request.headers.headers[key]...)
	}

	return allHeaders
}

func (request *Request) RemoveHeader(header GenericHeader) error {
	errNoMatch := fmt.Errorf("cannot remove header '%s' from request '%s' as it is not present",
		header.String(), request.Short())
	name := header.Name()

	headersOfSameType, isMatch := request.headers.headers[name]

	if !isMatch || len(headersOfSameType) == 0 {
		return errNoMatch
	}

	found := false
	for idx, hdr := range headersOfSameType {
		if hdr == header {
			request.headers.headers[name] = append(headersOfSameType[:idx], headersOfSameType[idx+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errNoMatch
	}

	if len(request.headers.headers[name]) == 0 {
		// The header we removed was the only one of its type.
		// Tidy up the header structure by removing the empty list value from the header map,
		// and removing the entry from the headerOrder list.
		delete(request.headers.headers, name)

		for idx, entry := range request.headerOrder {
			if entry == name {
				request.headers.headerOrder = append(request.headerOrder[:idx], request.headerOrder[idx+1:]...)
			}
		}
	}

	return nil
}

func (request *Request) GetBody() string {
	return request.Body
}

func (request *Request) SetBody(body string) {
	request.Body = body
}

// A SIP response object  (c.f. RFC 3261 section 7.2).
type Response struct {
	// The version of SIP used in this message, e.g. "SIP/2.0".
	SipVersion string

	// The response code, e.g. 200, 401 or 500.
	// This indicates the outcome of the originating request.
	StatusCode uint16

	// The reason string provides additional, human-readable information used to provide
	// clarification or explanation of the status code.
	// This will vary between different SIP UAs, and should not be interpreted by the receiving UA.
	Reason string

	// A response has headers.
	headers

	// The application data of the message.
	Body string
}

func NewResponse(sipVersion string, statusCode uint16, reason string, headers []GenericHeader, body string) (response *Response) {
	response = new(Response)
	response.SipVersion = sipVersion
	response.StatusCode = statusCode
	response.Reason = reason
	response.Body = body
	response.headers = newHeaders()
	response.headerOrder = make([]string, 0)

	for _, header := range headers {
		response.AddHeader(header)
	}

	return
}

func (response *Response) String() string {
	var buffer bytes.Buffer

	// Every SIP response starts with a Status Line - RFC 2361 7.2.
	buffer.WriteString(fmt.Sprintf("%s %d %s\r\n",
		response.SipVersion,
		response.StatusCode,
		response.Reason))

	// Write the headers.
	buffer.WriteString(response.headers.String())

	// If the request has a message body, add it.
	buffer.WriteString("\r\n" + response.Body)

	return buffer.String()
}

func (response *Response) Short() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%s %d %s\r\n",
		response.SipVersion,
		response.StatusCode,
		response.Reason))

	return buffer.String()
}

func (response *Response) AllHeaders() []GenericHeader {
	allHeaders := make([]GenericHeader, 0)
	for _, key := range response.headers.headerOrder {
		allHeaders = append(allHeaders, response.headers.headers[key]...)
	}

	return allHeaders
}

func (response *Response) RemoveHeader(header GenericHeader) error {
	errNoMatch := fmt.Errorf("cannot remove header '%s' from response '%s' as it is not present",
		header.String(), response.Short())
	name := header.Name()

	headersOfSameType, isMatch := response.headers.headers[name]

	if !isMatch || len(headersOfSameType) == 0 {
		return errNoMatch
	}

	found := false
	for idx, hdr := range headersOfSameType {
		if hdr == header {
			response.headers.headers[name] = append(headersOfSameType[:idx], headersOfSameType[idx+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errNoMatch
	}

	if len(response.headers.headers[name]) == 0 {
		// The header we removed was the only one of its type.
		// Tidy up the header structure by removing the empty list value from the header map,
		// and removing the entry from the headerOrder list.
		delete(response.headers.headers, name)

		for idx, entry := range response.headers.headerOrder {
			if entry == name {
				response.headers.headerOrder = append(response.headers.headerOrder[:idx], response.headers.headerOrder[idx+1:]...)
			}
		}
	}

	return nil
}

func (response *Response) GetBody() string {
	return response.Body
}

func (response *Response) SetBody(body string) {
	response.Body = body
}

func (request *Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Method     Method
		Recipient  Uri
		SipVersion string
		Headers    []GenericHeader
		Body       string
		Type       string
	}{
		Method:     request.Method,
		Recipient:  request.Recipient,
		SipVersion: request.SipVersion,
		Headers:    request.AllHeaders(),
		Body:       request.Body,
		Type:       "request",
	})
}

func (response *Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		SipVersion string
		StatusCode uint16
		Reason     string
		Headers    []GenericHeader
		Body       string
		Type       string
	}{
		SipVersion: response.SipVersion,
		StatusCode: response.StatusCode,
		Reason:     response.Reason,
		Headers:    response.AllHeaders(),
		Body:       response.Body,
		Type:       "response",
	})
}

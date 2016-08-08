package parser

import "github.com/FireSpotter/gossip/log"

// parserBuffer is a specialized buffer for use in the parser package.
// It is written to via the non-blocking Write.
// It exposes various blocking read methods, which wait until the requested
// data is avaiable, and then return it.
type parserBuffer struct {
	Msg    []byte
	index  int
	Length int
}

// Create a new parserBuffer object (see struct comment for object details).
// Note that resources owned by the parserBuffer may not be able to be GCed
// until the Dispose() method is called.
func newParserBuffer() *parserBuffer {
	var pb parserBuffer
	pb.index = 0
	return &pb
}

// Block until the buffer contains at least one CRLF-terminated line.
// Return the line, excluding the terminal CRLF, and delete it from the buffer.
// Returns an error if the parserbuffer has been stopped.
func (pb *parserBuffer) NextLine() (response string, err error) {
	var b byte
	var byteLine []byte
	for b != '\r' && b != '\n' && pb.index < pb.Length {
		b = pb.Msg[pb.index]

		byteLine = append(byteLine, b)
		pb.index += 1
	}
	if b == '\r' {
		pb.index += 1
	}
	response = string(byteLine)
	return
}

// Block until the buffer contains at least n characters.
// Return precisely those n characters, then delete them from the buffer.
func (pb *parserBuffer) NextChunk(n int) (response string, err error) {
	var data []byte
	var b byte

	for total := 0; total < n && pb.index < pb.Length; {
		b = pb.Msg[pb.index]
		data = append(data, b)
		pb.index += 1
		total += 1
	}

	response = string(data)
	log.Debug("Parser buffer returns chunk '%s'", response)
	return
}

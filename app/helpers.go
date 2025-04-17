package main

import (
	"net"
	"strings"
)

func (app *application) writeResponse(conn net.Conn, line string, headers map[string]string, body string, closed bool) {
	headerStr := ""
	line = line + "\r\n"

	for k, v := range headers {
		headerStr = headerStr + k + ": " + v + "\r\n"
	}

	if closed {
		headerStr = headerStr + "Connection: Close\r\n"
	}

	// End of headers
	headerStr = headerStr + "\r\n"

	conn.Write([]byte(line + headerStr + body))
}

func (app *application) extractRequest(request []byte) (string, map[string]string, string) {
	headers := make(map[string]string)
	reqStr := string(request)
	parts := strings.Split(reqStr, "\r\n")
	plen := len(parts)

	// NOTE: Format of the request:
	// requestLine\r\nheader[0]\r\nheader[1]\r\n...header[n]\r\n\r\nbody
	// part[0] is request line
	// part[1:n-2] is header, part[n-2] is just space due to \r\n\r\n
	// part[n-1] is body
	//
	line := parts[0]
	hpart := parts[1 : plen-2]
	for _, header := range hpart {
		headerParts := strings.Split(header, ": ")
		if len(headerParts) == 2 {
			headers[headerParts[0]] = headerParts[1]
		}
	}
	body := parts[plen-1]
	return line, headers, body
}

package main

import "net"

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

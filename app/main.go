package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
)

/* Request body sample
// Request line
GET                          // HTTP method
/index.html                  // Request target
HTTP/1.1                     // HTTP version
\r\n                         // CRLF that marks the end of the request line

// Headers
Host: localhost:4221\r\n     // Header that specifies the server's host and port
User-Agent: curl/7.64.1\r\n  // Header that describes the client's user agent
Accept: _something_\r\n              // Header that specifies which media types the client can accept
\r\n                         // CRLF that marks the end of the headers

// Request body (empty)
*/

/* Response body sample
// Status line
HTTP/1.1 200 OK
\r\n                          // CRLF that marks the end of the status line

// Headers
Content-Type: text/plain\r\n  // Header that specifies the format of the response body
Content-Length: 3\r\n         // Header that specifies the size of the response body, in bytes
\r\n                          // CRLF that marks the end of the headers

// Response body
abc                           // The string from the request
*/

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

const supportedCompression = "gzip"

// NOTE: Please don't judge me
// I'll refactor my code later :D
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {
		os.Exit(1)
	}

	for {

		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}

}

func extractRequest(req []byte) (string, map[string]string, string) {
	headers := make(map[string]string)
	reqString := string(req)
	reqParts := strings.Split(reqString, "\r\n")
	partLen := len(reqParts)

	// NOTE: Format of the request:
	// requestLine\r\nheader[0]\r\nheader[1]\r\n...header[n]\r\n\r\nbody
	// part[0] is request line
	// part[1:n-2] is header, part[n-2] is just space due to \r\n\r\n
	// part[n-1] is body
	//
	reqLine := reqParts[0]
	headerParts := reqParts[1 : partLen-2]
	for _, header := range headerParts {
		headerParts := strings.Split(header, ": ")
		if len(headerParts) == 2 {
			headers[headerParts[0]] = headerParts[1]
		}
	}
	body := reqParts[partLen-1]
	return reqLine, headers, body
}

func handleConnection(conn net.Conn) {

	for {
		req := make([]byte, 1024)
		resHeader := make(map[string]string)
		// NOTE: This slice will pad '\x00' aka 'null zero' for unused space

		n, err := conn.Read(req)
		if err != nil {
			if err.Error() == "EOF" {
				conn.Close()
				return
			}
			fmt.Println("Error reading request: ", err.Error())
			conn.Close()
			return
		}
		req = req[:n]
		reqLine, headers, body := extractRequest(req)

		reqLineParts := strings.Split(reqLine, " ")
		path := reqLineParts[1]

		closeCon := false

		if value, ok := headers["Connection"]; ok && value == "close" {
			closeCon = true
		}

		switch {
		case path == "/":
			writeResponse(conn, "HTTP/1.1 200 OK", resHeader, "", closeCon)

		case strings.HasPrefix(path, "/echo"):
			str := strings.TrimPrefix(path, "/echo/")
			if val, ok := headers["Accept-Encoding"]; ok {
				encodings := strings.Split(val, ", ")
				if slices.Contains(encodings, supportedCompression) {
					var b bytes.Buffer
					gw := gzip.NewWriter(&b)
					_, err := gw.Write([]byte(str))
					gw.Close()
					if err != nil {
						writeResponse(conn, "HTTP/1.1 500 Internal Server Error", resHeader, "", closeCon)
						break
					}

					// Construct new resHeader
					resHeader["Content-Type"] = "text/plain"
					resHeader["Content-Encoding"] = "gzip"
					resHeader["Content-Length"] = strconv.Itoa(len(b.String()))
					writeResponse(conn, "HTTP/1.1 200 OK", resHeader, b.String(), closeCon)
					break
				}
			}

			resHeader["Content-Type"] = "text/plain"
			resHeader["Content-Encoding"] = "gzip"
			resHeader["Content-Length"] = strconv.Itoa(len(str))
			delete(resHeader, "Content-Encoding")
			writeResponse(conn, "HTTP/1.1 200 OK", resHeader, str, closeCon)

		case strings.HasPrefix(path, "/user-agent"):
			agent := ""

			if val, ok := headers["User-Agent"]; ok {
				agent = val
			}

			resHeader["Content-Type"] = "text/plain"
			resHeader["Content-Length"] = strconv.Itoa(len(agent))
			writeResponse(conn, "HTTP/1.1 200 OK", resHeader, agent, closeCon)

		case strings.HasPrefix(path, "/files/"):
			dir := os.Args[2]
			fileName := dir + strings.TrimPrefix(path, "/files/")

			method := strings.Split(reqLine, " ")[0]
			if method == "GET" {
				_, err := os.Stat(fileName)
				if err != nil {
					writeResponse(conn, "HTTP/1.1 404 Not Found", resHeader, "", closeCon)
					break
				}
				content, err := os.ReadFile(fileName)
				if err != nil {
					writeResponse(conn, "HTTP/1.1 500 Internal Server Error", resHeader, "", closeCon)
					break
				}

				// Construct new resHeader
				resHeader["Content-Type"] = "application/octet-stream"
				resHeader["Content-Length"] = strconv.Itoa(len(content))
				writeResponse(conn, "HTTP/1.1 200 OK", resHeader, string(content), closeCon)
			} else if method == "POST" {
				writeBytes := make([]byte, len(strings.TrimRight(body, "\x00")))
				copy(writeBytes, body)
				err := os.WriteFile(fileName, writeBytes, 0644)
				if err != nil {
					writeResponse(conn, "HTTP/1.1 500 Internal Server Error", resHeader, "", closeCon)
					break
				}

				writeResponse(conn, "HTTP/1.1 201 Created", map[string]string{}, "", closeCon)
				conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
			}
		default:
			writeResponse(conn, "HTTP/1.1 404 Not Found", resHeader, "", closeCon)
		}

		if closeCon {
			conn.Close()
			return
		}

		clear(resHeader)
	}
}

func writeResponse(conn net.Conn, reqLine string, headers map[string]string, body string, isClose bool) {
	headerStr := ""
	reqLine = reqLine + "\r\n"

	for k, v := range headers {
		headerStr = headerStr + k + ": " + v + "\r\n"
	}

	if isClose {
		headerStr = headerStr + "Connection: Close\r\n"
	}

	// End of headers
	headerStr = headerStr + "\r\n"

	conn.Write([]byte(reqLine + headerStr + body))
}

package main

import (
	"fmt"
	"net"
	"os"
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

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		go handleConnection(l)
	}

}

func extractRequest(req []byte) (string, []string, string) {
	reqString := string(req)
	reqParts := strings.Split(reqString, "\r\n")
	partLen := len(reqParts)
	reqLine := reqParts[0]
	headers := reqParts[1 : partLen-1]
	body := reqParts[partLen-1]
	return reqLine, headers, body
}

func handleConnection(l net.Listener) {
	conn, err := l.Accept()

	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	defer conn.Close()

	// NOTE: This slice will pad '\x00' aka 'null zero' for unused space
	req := make([]byte, 1024)

	conn.Read(req)
	reqLine, headers, body := extractRequest(req)

	reqLineParts := strings.Split(reqLine, " ")
	path := reqLineParts[1]

	switch {
	case path == "/":
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	case strings.HasPrefix(path, "/echo"):
		str := strings.TrimPrefix(path, "/echo/")
		conn.Write([]byte("HTTP/1.1 200 OK\r\n" + "Content-Type: text/plain\r\nContent-Length: " + strconv.Itoa(len(str)) + "\r\n\r\n" + str))
	case strings.HasPrefix(path, "/user-agent"):
		agent := ""
		for _, header := range headers {
			if strings.HasPrefix(header, "User-Agent:") {
				agent = strings.TrimPrefix(header, "User-Agent: ")
			}
		}
		conn.Write([]byte("HTTP/1.1 200 OK\r\n" + "Content-Type: text/plain\r\nContent-Length: " + strconv.Itoa(len(agent)) + "\r\n\r\n" + agent))
	case strings.HasPrefix(path, "/files/"):
		dir := os.Args[2]
		fileName := dir + strings.TrimPrefix(path, "/files/")

		method := strings.Split(reqLine, " ")[0]
		if method == "GET" {
			_, err := os.Stat(fileName)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				break
			}
			content, err := os.ReadFile(fileName)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				break
			}
			conn.Write([]byte("HTTP/1.1 200 OK\r\n" + "Content-Type: application/octet-stream\r\nContent-Length: " + strconv.Itoa(len(content)) + "\r\n\r\n" + string(content)))
		} else if method == "POST" {
			writeBytes := make([]byte, len(strings.TrimRight(body, "\x00")))
			copy(writeBytes, body)
			err := os.WriteFile(fileName, writeBytes, 0644)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				break
			}
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}

	default:
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

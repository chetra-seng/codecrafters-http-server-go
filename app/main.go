package main

import (
	"fmt"
	"net"
	"os"
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

type application struct{}

var app *application

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {
		os.Exit(1)
	}

	for {
		con, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go app.handleRequest(con)
	}

}



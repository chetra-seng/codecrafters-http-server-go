package main

import (
	"fmt"
	"net"
	"strings"
)

func (app *application) handleRequest(con net.Conn) {

	for {
		req := make([]byte, 1024)
		resHeader := make(map[string]string)
		// NOTE: This slice will pad '\x00' aka 'null zero' for unused space

		n, err := con.Read(req)
		if err != nil {
			if err.Error() == "EOF" {
				con.Close()
				return
			}
			fmt.Println("Error reading request: ", err.Error())
			con.Close()
			return
		}
		req = req[:n]
		line, headers, body := app.extractRequest(req)

		lparts := strings.Split(line, " ")
		path := lparts[1]

		closed := false

		if value, ok := headers["Connection"]; ok && value == "close" {
			closed = true
		}

		switch {
		case path == "/":
			app.handleHome(con, headers, closed)

		case strings.HasPrefix(path, "/echo"):
			app.handleEcho(con, headers, closed, path)

		case strings.HasPrefix(path, "/user-agent"):
			app.handleUserAgent(con, headers, closed)

		case strings.HasPrefix(path, "/files/"):
			app.handleFiles(con, headers, body, closed, path, line)
		default:
			app.handleNotFound(con, headers, closed)
		}

		if closed {
			con.Close()
			return
		}

		clear(resHeader)
	}
}

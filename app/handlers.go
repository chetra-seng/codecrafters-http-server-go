package main

import (
	"bytes"
	"compress/gzip"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
)

func (app *application) handleHome(con net.Conn, headers map[string]string, closed bool) {
	app.writeResponse(con, "HTTP/1.1 200 OK", headers, "", closed)
}

func (app *application) handleEcho(con net.Conn, headers map[string]string, closed bool, path string) {
	str := strings.TrimPrefix(path, "/echo/")
	if val, ok := headers["Accept-Encoding"]; ok {
		encodings := strings.Split(val, ", ")
		if slices.Contains(encodings, supportedCompression) {
			var b bytes.Buffer
			gw := gzip.NewWriter(&b)
			_, err := gw.Write([]byte(str))
			gw.Close()
			if err != nil {
				app.writeResponse(con, "HTTP/1.1 500 Internal Server Error", headers, "", closed)
				return
			}

			// Construct new headers
			headers["Content-Type"] = "text/plain"
			headers["Content-Encoding"] = "gzip"
			headers["Content-Length"] = strconv.Itoa(len(b.String()))
			app.writeResponse(con, "HTTP/1.1 200 OK", headers, b.String(), closed)
			return
		}
	}

	headers["Content-Type"] = "text/plain"
	headers["Content-Encoding"] = "gzip"
	headers["Content-Length"] = strconv.Itoa(len(str))
	app.writeResponse(con, "HTTP/1.1 200 OK", headers, str, closed)
}

func (app *application) handleUserAgent(con net.Conn, headers map[string]string, closed bool) {
	agent := ""

	if val, ok := headers["User-Agent"]; ok {
		agent = val
	}

	headers["Content-Type"] = "text/plain"
	headers["Content-Length"] = strconv.Itoa(len(agent))
	app.writeResponse(con, "HTTP/1.1 200 OK", headers, agent, closed)
}

func (app *application) handleFiles(con net.Conn, headers map[string]string, body string, closed bool, path string, line string) {
	dir := os.Args[2]
	fileName := dir + strings.TrimPrefix(path, "/files/")

	method := strings.Split(line, " ")[0]
	switch method {
	case "GET":
		_, err := os.Stat(fileName)
		if err != nil {
			app.writeResponse(con, "HTTP/1.1 404 Not Found", headers, "", closed)
			return
		}
		content, err := os.ReadFile(fileName)
		if err != nil {
			app.writeResponse(con, "HTTP/1.1 500 Internal Server Error", headers, "", closed)
			return
		}

		// Construct new headers
		headers["Content-Type"] = "application/octet-stream"
		headers["Content-Length"] = strconv.Itoa(len(content))
		app.writeResponse(con, "HTTP/1.1 200 OK", headers, string(content), closed)
	case "POST":

		writeBytes := make([]byte, len(strings.TrimRight(body, "\x00")))
		copy(writeBytes, body)
		err := os.WriteFile(fileName, writeBytes, 0644)
		if err != nil {
			app.writeResponse(con, "HTTP/1.1 500 Internal Server Error", headers, "", closed)
			return
		}

		app.writeResponse(con, "HTTP/1.1 201 Created", map[string]string{}, "", closed)
		con.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
	}
}

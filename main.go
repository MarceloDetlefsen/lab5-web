package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Leer primera línea del request HTTP
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading request: %v", err)
		sendError(conn, "500 Internal Server Error", "Error reading request")
		return
	}

	log.Printf("Request line: %s", requestLine)

	requestLine = strings.TrimSpace(requestLine)
	parts := strings.Split(requestLine, " ")

	// Validar formato básico del request
	if len(parts) < 3 {
		sendError(conn, "400 Bad Request", "Invalid HTTP request format")
		return
	}

	method := parts[0]
	path := parts[1]

	// Solo aceptar método GET
	if method != "GET" {
		sendError(conn, "400 Bad Request", "Only GET method is supported")
		return
	}

	// Construir cuerpo HTML dinámico
	body := fmt.Sprintf(`
	<html>
		<body>
			<h1>Hello! You requested: %s</h1>
		</body>
	</html>
	`, path)

	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n"+
			"%s",
		len(body),
		body,
	)

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

func sendError(conn net.Conn, status string, message string) {
	body := fmt.Sprintf(`
	<html>
		<body>
			<h1>%s</h1>
			<p>%s</p>
		</body>
	</html>
	`, status, message)

	response := fmt.Sprintf(
		"HTTP/1.1 %s\r\n"+
			"Content-Type: text/html\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n"+
			"%s",
		status,
		len(body),
		body,
	)

	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending error response: %v", err)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	log.Print("Listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net"
	"strings"

	_ "modernc.org/sqlite"
)

func handleConnection(conn net.Conn, db *sql.DB) {
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
	_ = path //En este momento no se usa el path

	// Solo aceptar método GET
	if method != "GET" {
		sendError(conn, "400 Bad Request", "Only GET method is supported")
		return
	}

	rows, err := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
	if err != nil {
		log.Printf("Database query error: %v", err)
		sendError(conn, "500 Internal Server Error", "Database error")
		return
	}
	defer rows.Close()

	body := `
	<html>
	<head>
		<title>Control de Series</title>
		<style>
			body { font-family: Arial; background: #f4f4f4; padding: 40px; }
			h1 { text-align: center; }
			table {
				margin: auto;
				border-collapse: collapse;
				width: 70%;
				background: white;
			}
			p {
				text-align: center;
				font-style: italic;
				color: #555;
			}
			th, td {
				border: 1px solid #000000;
				padding: 10px;
				text-align: center;
			}
			th {
				background: #ffb545;
				color: white;
			}
			tr:nth-child(even) { background: #6ec8ff; }
			tr:nth-child(odd) { background: #cae6ff; }
		</style>
	</head>
	<body>
	<h1>Control de Series</h1>
	<p> ( No miro series :/ ) <br> Solo puse datos de series que conozco, pero no son mis estadisticas.</p>
	<table>
	<tr>
		<th>ID</th>
		<th>Serie</th>
		<th>Episodio Actual</th>
		<th>Total de Episodios</th>
	</tr>
	`

	for rows.Next() {
		var id int
		var name string
		var current int
		var total int

		err := rows.Scan(&id, &name, &current, &total)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		body += fmt.Sprintf(`
		<tr>
			<td>%d</td>
			<td>%s</td>
			<td>%d</td>
			<td>%d</td>
		</tr>
		`, id, name, current, total)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
	}

	body += `
	</table>
	</body>
	</html>
	`

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
	db, err := sql.Open("sqlite", "file:series.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

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

		go handleConnection(conn, db)
	}
}

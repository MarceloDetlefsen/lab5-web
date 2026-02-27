package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
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

	if len(parts) < 3 {
		sendError(conn, "400 Bad Request", "Invalid HTTP request format")
		return
	}

	method := parts[0]
	path := parts[1]

	// Leer headers
	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // fin de headers
		}
		if strings.HasPrefix(line, "Content-Length:") {
			lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, _ = strconv.Atoi(lengthStr)
		}
	}

	// Routing
	switch {
	case method == "GET" && path == "/":
		handleIndex(conn, db)

	case method == "GET" && path == "/create":
		handleCreateForm(conn)

	case method == "POST" && path == "/create":
		// Leer el body
		body := make([]byte, contentLength)
		totalRead := 0
		for totalRead < contentLength {
			n, err := reader.Read(body[totalRead:])
			totalRead += n
			if err != nil {
				break
			}
		}
		handleCreatePost(conn, db, string(body[:totalRead]))

	default:
		sendError(conn, "404 Not Found", "Page not found")
	}
}

func handleIndex(conn net.Conn, db *sql.DB) {
	rows, err := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
	if err != nil {
		log.Printf("Database query error: %v", err)
		sendError(conn, "500 Internal Server Error", "Database error")
		return
	}
	defer rows.Close()

	// Construir las filas de la tabla
	tableRows := ""
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

		percent := 0
		if total > 0 {
			percent = (current * 100) / total
		}

		tableRows += fmt.Sprintf(`
		<tr>
			<td>%d</td>
			<td>%s</td>
			<td>%d</td>
			<td>%d</td>
			<td>
				<div class="progress-container">
					<div class="progress-bar" style="width:%d%%;">
						%d%%
					</div>
				</div>
			</td>
		</tr>
		`, id, name, current, total, percent, percent)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
	}

	sendHTML(conn, "200 OK", indexTemplate(tableRows))
}

func handleCreateForm(conn net.Conn) {
	sendHTML(conn, "200 OK", createFormTemplate())
}

func handleCreatePost(conn net.Conn, db *sql.DB, rawBody string) {
	log.Printf("POST body recibido: %s", rawBody)

	values, err := url.ParseQuery(rawBody)
	if err != nil {
		log.Printf("Error parsing form body: %v", err)
		sendError(conn, "400 Bad Request", "Invalid form data")
		return
	}

	name := values.Get("series_name")
	currentEp := values.Get("current_episode")
	totalEps := values.Get("total_episodes")

	log.Printf("Parsed -> name=%s, current=%s, total=%s", name, currentEp, totalEps)

	if name == "" || currentEp == "" || totalEps == "" {
		sendError(conn, "400 Bad Request", "Todos los campos son requeridos")
		return
	}

	_, err = db.Exec(
		"INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
		name, currentEp, totalEps,
	)
	if err != nil {
		log.Printf("Error inserting into DB: %v", err)
		sendError(conn, "500 Internal Server Error", "Error al guardar en la base de datos")
		return
	}

	// POST/Redirect/GET
	response := "HTTP/1.1 303 See Other\r\nLocation: /\r\n\r\n"
	conn.Write([]byte(response))
}

func sendHTML(conn net.Conn, status string, body string) {
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
		log.Printf("Error writing response: %v", err)
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

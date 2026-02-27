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
				width: 80%;
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
			.progress-container {
				width: 100%;
				background-color: #ddd;
				border-radius: 10px;
				overflow: hidden;
			}
			.progress-bar {
				height: 20px;
				background-color: #4CAF50;
				text-align: center;
				color: white;
				line-height: 20px;
				font-size: 12px;
			}
			.add-link {
				display: block;
				text-align: center;
				margin-bottom: 20px;
				font-size: 16px;
			}
			a { color: #ffb545; }
		</style>
	</head>
	<body>
	<h1>Control de Series</h1>
	<p> ( No miro series :/ ) <br> Solo puse datos de series que conozco, pero no son mis estadisticas.</p>
	<table>
	<a class="add-link" href="/create">Agregar Serie</a>
	<tr>
		<th>ID</th>
		<th>Serie</th>
		<th>Episodio Actual</th>
		<th>Total de Episodios</th>
		<th>Progreso</th>
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

		percent := 0
		if total > 0 {
			percent = (current * 100) / total
		}

		body += fmt.Sprintf(`
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

	body += `
	</table>
	</body>
	</html>
	`

	sendHTML(conn, "200 OK", body)
}

func handleCreateForm(conn net.Conn) {
	body := `
	<html>
	<head>
		<title>Agregar Serie</title>
		<style>
			body { font-family: Arial; background: #f4f4f4; padding: 40px; }
			h1 { text-align: center; }
			form {
				max-width: 400px;
				margin: auto;
				background: white;
				padding: 30px;
				border-radius: 8px;
				box-shadow: 0 2px 8px rgba(0,0,0,0.1);
			}
			label { display: block; margin-top: 15px; font-weight: bold; }
			input {
				width: 100%;
				padding: 8px;
				margin-top: 5px;
				box-sizing: border-box;
				border: 1px solid #ccc;
				border-radius: 4px;
			}
			button {
				margin-top: 20px;
				width: 100%;
				padding: 10px;
				background: #ffb545;
				color: white;
				border: none;
				border-radius: 4px;
				font-size: 16px;
				cursor: pointer;
			}
			button:hover { background: #e0a030; }
			.back-link { display: block; text-align: center; margin-top: 15px; }
			a { color: #ffb545; }
		</style>
	</head>
	<body>
	<h1>Agregar Nueva Serie</h1>
	<form method="POST" action="/create">
		<label>Nombre de la serie:</label>
		<input type="text" name="series_name" required>

		<label>Episodio actual:</label>
		<input type="number" name="current_episode" min="0" value="1" required>

		<label>Total de episodios:</label>
		<input type="number" name="total_episodes" min="1" required>

		<button type="submit">Agregar Serie</button>
	</form>
	<a class="back-link" href="/">Volver a la lista</a>
	</body>
	</html>
	`

	sendHTML(conn, "200 OK", body)
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

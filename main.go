package main

import (
	"database/sql"
	"log"
	"net"

	_ "modernc.org/sqlite"
)

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

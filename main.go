package main

import (
	"fmt"
	"log"
	"net/http"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func form_results(w http.ResponseWriter, r *http.Request) []string {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		lastname := r.FormValue("lastname")
		stNumber := r.FormValue("stNumber")
		email := r.FormValue("email")
		message := r.FormValue("message")

		list := []string{name, lastname, stNumber, email, message}

		fmt.Fprintf(w, "Form submitted successfully!")

		return list
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	return nil
}

func dbConnection() (*sql.DB, error) {
	db, err := sql.Open("mysql", "bilol:1234@tcp(127.0.0.1:3306)/feedback")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MySQL database!")

	return db, nil
}

func showData(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := dbConnection()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	// Query the database for all records in the user_messages table
	rows, err := db.Query("SELECT * FROM user_messages")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error querying the database:", err)
		return
	}
	defer rows.Close()

	// Create a slice to hold the retrieved data
	var messages []string

	// Iterate over the result rows
	for rows.Next() {
		var (
			id       int
			name     string
			lastname string
			stNumber string
			email    string
			message  string
		)

		// Scan the row values into variables
		err := rows.Scan(&id, &name, &lastname, &stNumber, &email, &message)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("Error scanning row:", err)
			return
		}

		// Append the data to the messages slice
		messages = append(messages, fmt.Sprintf("ID: %d, Name: %s, Lastname: %s, Student Number: %s, Email: %s, Message: %s",
			id, name, lastname, stNumber, email, message))
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error iterating over rows:", err)
		return
	}

	// Render the retrieved data in the response
	w.Header().Set("Content-Type", "text/plain")
	for _, msg := range messages {
		fmt.Fprintln(w, msg)
	}
}

func main() {
	db, err := dbConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			list := form_results(w, r)
			if list != nil {
				// Insert the list values into the database
				db, err := dbConnection()
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					log.Println("Error connecting to the database:", err)
					return
				}
				defer db.Close()

				insertQuery := "INSERT INTO user_messages (name, lastname, stNumber, email, message) VALUES (?, ?, ?, ?, ?)"
				_, err = db.Exec(insertQuery, list[0], list[1], list[2], list[3], list[4])
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					log.Println("Error inserting data into the database:", err)
					return
				}

				fmt.Fprint(w, "Form submitted successfully!")
				return
			}
		}

		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	http.Handle("/", http.FileServer(http.Dir(".")))

	http.HandleFunc("/show", showData)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

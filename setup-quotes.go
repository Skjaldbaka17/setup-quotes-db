package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/jackc/pgx/v4"
)

func getJSON(path string) (map[string][]string, error) {
	// Open JSON
	jsonFile, err := os.Open(path)
	// if os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}
	fmt.Println(path, "has been opened!")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	//Read the opened file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var authors map[string][]string
	//Convert the read value to json and put into the authors-var
	json.Unmarshal(byteValue, &authors)

	return authors, nil
}

func addAuthor(conn *pgx.Conn, name string) (int, error) {
	var id int
	err := conn.QueryRow(context.Background(), "insert into authors (name) values($1) returning id", name).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return -1, err
	}
	return id, nil
}

func addQuote(conn *pgx.Conn, quote string, author_id int) (int, error) {
	var id int
	err := conn.QueryRow(context.Background(), "insert into quotes (quote, author_id) values($1,$2) returning id", quote, author_id).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return -1, err
	}
	return id, nil
}

func fillTableWithData(conn *pgx.Conn, authors map[string][]string) {
	for author, quotes := range authors {
		author_id, err := addAuthor(conn, author)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n Could not add author! %s \n", err, author)
			continue
		}

		for _, quote := range quotes {
			query := fmt.Sprintf("insert into quotes (author_id, quote) values(%d, '%s') returning id", author_id, string(quote))
			_, err := addQuote(conn, quote, author_id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n Query failed: %s", err, query)
				continue
			}
		}
	}
}

func main() {
	//Connect to DB
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		authors, err := getJSON("../Database-650000-Quotes/English/A.json")

		if err != nil {
			fmt.Fprintf(os.Stderr, "Reading JSON file Failed", err)
			return
		}

		fillTableWithData(conn, authors)
	}()

	wg.Wait()
}

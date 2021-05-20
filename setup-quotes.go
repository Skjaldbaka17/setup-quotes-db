package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
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
		fmt.Printf("Author %s added!", author)
	}
}

// ReadDir reads the directory named by dirname and returns
// a list of directory entries sorted by filename.
func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	return list, nil
}

func main() {

	var wg sync.WaitGroup
	basePath := "../Database-650000-Quotes/English"

	re1, _ := regexp.Compile(`.json`)
	info, _ := ReadDir(basePath)

	for idx, name := range info {
		if idx > 2 {
			break
		}
		if re1.MatchString(name.Name()) {
			wg.Add(1)
			go func(name string) {
				conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
					os.Exit(1)
				}
				defer conn.Close(context.Background())
				defer wg.Done()
				fmt.Println(name)
				authors, err := getJSON("../Database-650000-Quotes/English/" + name)

				if err != nil {
					fmt.Fprintf(os.Stderr, "Reading JSON file Failed", err)
					return
				}

				fillTableWithData(conn, authors)
			}(name.Name())
		}
	}

	wg.Wait()
}

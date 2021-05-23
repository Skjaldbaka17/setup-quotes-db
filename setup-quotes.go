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

	"github.com/jackc/pgx/v4/pgxpool"
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

func addAuthor(conn *pgxpool.Pool, name string) (int, error) {
	var id int
	err := conn.QueryRow(context.Background(), "insert into authors (name) values($1) returning id", name).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return -1, err
	}
	return id, nil
}

func addQuote(conn *pgxpool.Pool, quote string, authorid int, isIcelandic bool) (int, error) {
	var id int
	err := conn.QueryRow(context.Background(), "insert into quotes (quote, authorid) values($1,$2) returning id", quote, authorid).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return -1, err
	}
	return id, nil
}

func fillTableWithData(conn *pgxpool.Pool, authors map[string][]string, isIcelandic bool) {
	for author, quotes := range authors {
		authorid, err := addAuthor(conn, author)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n Could not add author! %s \n", err, author)
			continue
		}

		for _, quote := range quotes {
			query := fmt.Sprintf("insert into quotes (authorid, quote) values(%d, '%s') returning id", authorid, string(quote))
			_, err := addQuote(conn, quote, authorid, isIcelandic)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n Query failed: %s", err, query)
				continue
			}
		}
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

func readTextFile(path string) string {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(dat)
}

func setupDBEnv(conn *pgxpool.Pool) error {
	var err error
	if err != nil {
		return err
	}

	_, err = conn.Query(context.Background(), "drop table if exists authors,quotes;")
	if err != nil {
		return err
	}

	file := readTextFile("./sql/authors.sql")
	_, err = conn.Query(context.Background(), file)
	if err != nil {
		return err
	}
	file = readTextFile("./sql/quotes.sql")
	_, err = conn.Query(context.Background(), file)
	if err != nil {
		return err
	}

	file = readTextFile("./sql/searchView.sql")
	_, err = conn.Query(context.Background(), file)
	if err != nil {
		return err
	}

	file = readTextFile("./sql/initQueries.sql")
	_, err = conn.Query(context.Background(), file)
	if err != nil {
		return err
	}

	return err
}

func insertIntoDB(pool *pgxpool.Pool) {
	var wg sync.WaitGroup
	basePath := "../Database-650000-Quotes/"

	re1, _ := regexp.Compile(`.json`)

	for i := 0; i < 2; i++ {
		var path string
		isIcelandic := i == 1
		if isIcelandic {
			path = basePath + "Icelandic"
		} else {
			path = basePath + "English"
		}
		info, _ := ReadDir(path)

		for _, name := range info {
			if re1.MatchString(name.Name()) {

				wg.Add(1)
				go func(name string, isIcelandic bool) {
					defer wg.Done()
					fmt.Println(name)
					authors, err := getJSON(fmt.Sprintf("%s/%s", path, name))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Reading JSON file Failed, %s", err)
						return
					}

					fillTableWithData(pool, authors, isIcelandic)
					fmt.Println(fmt.Sprintf("%s/%s", path, name), "is Done!")
				}(name.Name(), isIcelandic)
			}
		}
	}

	wg.Wait()
}

func main() {
	poolConn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}
	// defer poolConn.Close() //does not work!? Make program run forever, as if waiting for some connection?
	err = setupDBEnv(poolConn)

	fmt.Print(err)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	insertIntoDB(poolConn)
}

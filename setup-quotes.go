package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/Skjaldbaka17/setup-quotes-db/handlers"
	"github.com/jackc/pgx/v4/pgxpool"
)

func fillTableWithData(conn *pgxpool.Pool, authors map[string][]string, isIcelandic bool) {
	for author, quotes := range authors {
		authorid, err := handlers.AddAuthor(conn, author)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n Could not add author! %s \n", err, author)
			continue
		}

		for _, quote := range quotes {
			query := fmt.Sprintf("insert into quotes (authorid, quote, isicelandic) values(%d, '%s', %t) returning id", authorid, string(quote), isIcelandic)
			_, err := handlers.AddQuote(conn, quote, authorid, isIcelandic)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n Query failed: %s", err, query)
				continue
			}
		}
	}
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
		info, _ := handlers.ReadDir(path)

		for _, name := range info {
			if re1.MatchString(name.Name()) {
				wg.Add(1)
				go func(name string, isIcelandic bool) {
					defer wg.Done()
					authors, err := handlers.GetJSON(fmt.Sprintf("%s/%s", path, name))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Reading JSON file Failed, %s", err)
						return
					}

					fillTableWithData(pool, authors, isIcelandic)
					fmt.Println(fmt.Sprintf("%s/%s", path, name), "is Done!")
				}(name.Name(), isIcelandic)
			}
		}
		wg.Wait() //Because we want the english version to finish first because some icelandic quotes are in english and if
		//those quotes are also in the english DB then we want to "combine" them and mark as isIcelandic=false
	}
}

func insertIntoTopicsDB(pool *pgxpool.Pool) {
	var wg sync.WaitGroup
	basePath := "../Database-650000-Quotes/Topics/"

	re1, _ := regexp.Compile(`.json`)
	for i := 0; i < 2; i++ {
		var path string
		isIcelandic := i == 1
		if isIcelandic {
			path = basePath + "IcelandicTopicsJsons"
		} else {
			path = basePath + "EnglishTopicsJsons"
		}
		info, _ := handlers.ReadDir(path)
		for _, name := range info {
			if re1.MatchString(name.Name()) {
				wg.Add(1)
				go func(name string) {
					defer wg.Done()

					if isIcelandic {
						authors, err := handlers.GetIcelandicTopicJSON(fmt.Sprintf("%s/%s", path, name))

						if err != nil {
							fmt.Fprintf(os.Stderr, "Reading JSON file Failed, %s", err)
							return
						}
						nameWithoutType := re1.ReplaceAllString(name, "")
						handlers.AddIcelandicTopic(pool, nameWithoutType, authors, isIcelandic)
					} else {
						authors, err := handlers.GetTopicJSON(fmt.Sprintf("%s/%s", path, name))

						if err != nil {
							fmt.Fprintf(os.Stderr, "Reading JSON file Failed, %s", err)
							return
						}
						nameWithoutType := re1.ReplaceAllString(name, "")
						handlers.AddTopic(pool, nameWithoutType, authors, isIcelandic)
					}

					fmt.Println(fmt.Sprintf("%s/%s", path, name), "is Done!")
				}(name.Name())
			}
		}
	}

	wg.Wait() //Because we want the english version to finish first because some icelandic quotes are in english and if
	//those quotes are also in the english DB then we want to "combine" them and mark as isIcelandic=false

}

func finalDBQueries(pool *pgxpool.Pool) error {
	var wg sync.WaitGroup

	var err error
	fmt.Println("Running final queries...")

	wrapUpFile := handlers.ReadTextFile("./sql/wrapUpQueries.sql")
	scanner := bufio.NewScanner(strings.NewReader(wrapUpFile))

	for scanner.Scan() {
		query := scanner.Text()
		if query == "" {
			continue
		}
		wg.Add(1)
		go func(query string) {
			log.Println("Running: ", query)
			defer wg.Done()
			_, err := pool.Exec(context.Background(), query)
			if err != nil {
				log.Printf("Reading JSON file Failed, %s", err)
			}

		}(query)
	}

	log.Println("Creating searchView...")
	file := handlers.ReadTextFile("./sql/searchView.sql")
	_, err = pool.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}

func main() {
	poolConn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	// defer poolConn.Close() //does not work!? Make program run forever, as if waiting for some connection?
	err = handlers.SetupDBEnv(poolConn)

	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	insertIntoDB(poolConn)

	insertIntoTopicsDB(poolConn)

	defer finalDBQueries(poolConn)
}

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
	"time"

	"github.com/Skjaldbaka17/setup-quotes-db/handlers"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func fillTableWithData(conn *pgxpool.Pool, authors map[string][]string, isIcelandic bool) {
	for author, quotes := range authors {
		authorid, err := handlers.AddAuthor(conn, author, isIcelandic, len(quotes))
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

//TODO count icelandic quotes and english separately, or not you decide
//Will update the nr of quotes both icelandic and english
func updateNrOfQuotesPerAuthor(poolConn *pgxpool.Pool) {
	start := time.Now()
	query := "select id from authors;"

	rows, _ := poolConn.Query(context.Background(), query)

	i := 0
	for rows.Next() {
		if i%100 == 0 {
			log.Println("Nr Rows updated:", i)
		}
		i++
		if err := rows.Err(); err != nil {
			log.Println("next row", err)
			continue
		}
		vals, _ := rows.Values()
		authId := vals[0]

		var count int
		rows1 := poolConn.QueryRow(context.Background(), "select count(*) from quotes where authorid = $1", authId)
		_ = rows1.Scan(&count)
		nrOfQuotes := count
		_, _ = poolConn.Exec(context.Background(), "update authors set nrofquotes = $1 where id = $2 returning *", nrOfQuotes, authId)
	}

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Time: %f", elapsed.Seconds())
}

func readResponse(command string) string {
	fmt.Println("Please", command)
	var response string
	reader := bufio.NewReader(os.Stdin)
	response, _ = reader.ReadString('\n')
	response = strings.Trim(response, "\n")
	if response != "" {
		return response
	} else {

		return readResponse(command)
	}
}

func createAdminUser(poolConn *pgxpool.Pool) error {
	userName := readResponse("choose a username for the GOD user:")
	passWord := readResponse("choose a password:")
	hash, _ := bcrypt.GenerateFromPassword([]byte(passWord), bcrypt.DefaultCost)
	return handlers.SaveAdmin(userName, string(hash), poolConn)
}

func main() {
	poolConn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	_, err = poolConn.Exec(context.Background(), "drop view if exists searchview, topicsview, qodview;")
	if err != nil {
		fmt.Println(err)
		// return err
	}

	file := handlers.ReadTextFile("./sql/qodview.sql")
	_, err = poolConn.Exec(context.Background(), file)
	if err != nil {
		fmt.Println(err)
		// return err
	}

	file = handlers.ReadTextFile("./sql/topicsView.sql")
	_, err = poolConn.Exec(context.Background(), file)
	if err != nil {
		fmt.Println(err)
	}

	file = handlers.ReadTextFile("./sql/searchView.sql")
	_, err = poolConn.Exec(context.Background(), file)
	if err != nil {
		fmt.Println(err)
	}

	// err = createAdminUser(poolConn)
	// if err != nil {
	// 	fmt.Printf("error: %s", err)
	// 	return
	// }
	// fmt.Println("GOD user created!")

	// // defer poolConn.Close() //does not work!? Make program run forever, as if waiting for some connection?
	// err = handlers.SetupDBEnv(poolConn)

	// if err != nil {
	// 	fmt.Printf("error: %s", err)
	// 	return
	// }

	// insertIntoDB(poolConn)

	// insertIntoTopicsDB(poolConn)

	// defer finalDBQueries(poolConn)
}

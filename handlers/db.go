package handlers

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func AddAuthor(conn *pgxpool.Pool, name string, isIcelandic bool, nrOfQuotes int) (int, error) {
	var id int
	var err error
	if isIcelandic {
		if nrOfQuotes >= 0 {
			err = conn.QueryRow(context.Background(), "insert into authors (name,has_icelandic_quotes,nr_of_icelandic_quotes) values($1,$2,$3) on conflict (name) do update set has_icelandic_quotes = $2, nr_of_icelandic_quotes=$3 returning id", name, isIcelandic, nrOfQuotes).Scan(&id)
		} else {
			err = conn.QueryRow(context.Background(), "insert into authors (name,has_icelandic_quotes) values($1,$2) on conflict (name) do update set has_icelandic_quotes = $2 returning id", name, isIcelandic).Scan(&id)
		}
	} else {
		if nrOfQuotes >= 0 {
			err = conn.QueryRow(context.Background(), "insert into authors (name, nr_of_english_quotes) values($1,$2) on conflict (name) do update set name = $1,nr_of_english_quotes=$2 returning id", name, nrOfQuotes).Scan(&id)
		} else {
			err = conn.QueryRow(context.Background(), "insert into authors (name) values($1) on conflict (name) do update set name = $1 returning id", name).Scan(&id)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return -1, err
	}
	return id, nil
}

func AddQuote(conn *pgxpool.Pool, quote string, authorid int, isIcelandic bool) (int, error) {
	var id int
	err := conn.QueryRow(context.Background(), "insert into quotes (quote, author_id, is_icelandic) values($1,$2, $3) on conflict (quote) do update set author_id = $2 returning id", quote, authorid, isIcelandic).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return -1, err
	}
	return id, nil
}

//Create topic inserts a topic row into topics table and returns its id, or an error if fail
func createTopic(conn *pgxpool.Pool, name string, isIcelandic bool) (int, error) {
	var id int
	err := conn.QueryRow(context.Background(), "insert into topics (name, is_icelandic) values($1,$2) on conflict (name) do update set name = $1 returning id", name, isIcelandic).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return -1, err
	}
	return id, nil
}

func AddQuoteToTopic(conn *pgxpool.Pool, topicId int, quoteId int) error {
	var id int
	err := conn.QueryRow(context.Background(), "insert into topicstoquotes (topic_id, quote_id) values($1,$2) returning id", topicId, quoteId).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return err
	}
	return nil
}

func AddIcelandicTopic(conn *pgxpool.Pool, topicName string, quotes map[string][]string, isIcelandic bool) error {
	topicId, err := createTopic(conn, topicName, isIcelandic)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Creating Topic Failed: %v\n", err)
		return err
	}

	for author, quoteArray := range quotes {
		var authorId, quoteId int
		authorId, err = AddAuthor(conn, author, isIcelandic, -1)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Inserting author for Topic failed: %v\n", err)
			return err
		}

		for _, quote := range quoteArray {
			quoteId, err = AddQuote(conn, quote, authorId, isIcelandic)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Inserting quote for topic failed: %v\n", err)
				return err
			}
			err = AddQuoteToTopic(conn, topicId, quoteId)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Connecting topic to quote failed: %v\n", err)
				return err
			}
		}

	}

	return nil
}

func AddTopic(conn *pgxpool.Pool, topicName string, quotes map[string]string, isIcelandic bool) error {
	topicId, err := createTopic(conn, topicName, isIcelandic)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Creating Topic Failed: %v\n", err)
		return err
	}

	for author, quote := range quotes {
		var authorId, quoteId int
		authorId, err = AddAuthor(conn, author, isIcelandic, -1)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Inserting author for Topic failed: %v\n", err)
			return err
		}
		quoteId, err = AddQuote(conn, quote, authorId, isIcelandic)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Inserting quote for topic failed: %v\n", err)
			return err
		}
		err = AddQuoteToTopic(conn, topicId, quoteId)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Connecting topic to quote failed: %v\n", err)
			return err
		}
	}

	return nil
}

func SaveAdmin(userName string, passWordHash string, email string, apiKey string, conn *pgxpool.Pool) error {
	_, err := conn.Exec(context.Background(), "insert into users (name,password_hash,tier,email,api_key) values($1,$2,'GOD',$3,$4)", userName, passWordHash, email, apiKey)
	return err
}

func dropStuff(conn *pgxpool.Pool) error {
	log.Println("Running: drop view if exists aodview, aodiceview ,searchviews, topicsview, qodview, qodiceview;")
	_, err := conn.Exec(context.Background(), "drop view if exists aodview, aodiceview, qodview, qodiceview;")
	if err != nil {
		return err
	}

	log.Println("Running: drop materialized view if exists searchviews, topicsview, unique_lexeme_authors, unique_lexeme_quotes, unique_lexeme, popularityView;")
	_, err = conn.Exec(context.Background(), "drop materialized view if exists searchviews, topicsview, unique_lexeme_authors, unique_lexeme_quotes, unique_lexeme, popularityView;")
	if err != nil {
		return err
	}

	log.Println("Running: drop table if exists errorhistory, requesthistory, users, qod, aod, aodice, qodice, topicstoquotes, topics, quotes, authors cascade;")
	_, err = conn.Exec(context.Background(), "drop table if exists errorhistory, requesthistory,aod, aodice, qod, qodice, users, topicstoquotes, topics, quotes, authors cascade;")
	if err != nil {
		return err
	}

	return nil
}

func SetupDBEnv(conn *pgxpool.Pool) error {
	var err error
	if err != nil {
		return err
	}

	err = dropStuff(conn)
	if err != nil {
		return err
	}
	file := ReadTextFile("./sql/authors.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}
	file = ReadTextFile("./sql/quotes.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}
	file = ReadTextFile("./sql/topics.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/topicsToQuotes.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/qod.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/qodice.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/aod.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/aodice.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/users.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/requesthistory.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/errorhistory.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/qodview.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}
	file = ReadTextFile("./sql/qodiceview.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/aodview.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}
	file = ReadTextFile("./sql/aodiceview.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	file = ReadTextFile("./sql/popularityView.sql")
	_, err = conn.Exec(context.Background(), file)
	if err != nil {
		return err
	}

	return err
}

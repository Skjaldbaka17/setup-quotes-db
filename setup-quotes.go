package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type Author struct {
}

func getJSON(path string) map[string][]string {
	// Open JSON
	jsonFile, err := os.Open(path)
	// if os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(path, "has been opened!")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	//Read the opened file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var authors map[string][]string
	//Convert the read value to json and put into the authors-var
	json.Unmarshal(byteValue, &authors)

	return authors
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		authors := getJSON("../Database-650000-Quotes/English/A.json")
		for author, quotes := range authors {
			fmt.Println(author)
			for _, quote := range quotes {
				fmt.Println(quote)
			}
		}

	}()

	wg.Wait()
}

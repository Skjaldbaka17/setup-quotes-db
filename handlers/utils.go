package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

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

func ReadTextFile(path string) string {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(dat)
}

func GetIcelandicTopicJSON(path string) (map[string][]string, error) {
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

	var quotes []string

	//Convert the read value to json and put into the authors-var
	json.Unmarshal(byteValue, &quotes)
	authors := make(map[string][]string)
	authors["unknown"] = quotes
	return authors, nil
}

func GetTopicJSON(path string) (map[string]string, error) {
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

	var authors map[string]string
	//Convert the read value to json and put into the authors-var
	json.Unmarshal(byteValue, &authors)

	return authors, nil
}

func GetJSON(path string) (map[string][]string, error) {
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

# setup-quotes

This is a package for creating the Postgres database for the Quotes/Authors, found in the public repo https://github.com/Skjaldbaka17/Database-650000-Quotes, that the Quotes-api uses https://github.com/Skjaldbaka17/quotes-api.

## Requirements

* [Golang](https://golang.org)
* The Quotes-data in the directory above this one: https://github.com/Skjaldbaka17/Database-650000-Quotes

## Setup

Clone the https://github.com/Skjaldbaka17/Database-650000-Quotes repo into the parent directory of this project. 

Run the following to get the quotel data

```bash
    mkdir ../Database-650000-Quotes
    git clone https://github.com/Skjaldbaka17/Database-650000-Quotes ../Database-650000-Quotes
```

Then to create the Postgres DB create a `.env` file with `DATABASE_URL=YOUR_DB_URL` and then run:

```bash
    go mod tidy
    make setup
```

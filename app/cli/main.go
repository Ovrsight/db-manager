package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nizigama/ovrsight/business/databases"
	"log"
)

const envFile = ".env"

func main() {

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalln(err)
	}

	if err := databases.Ping(); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Backing up is possible")
}

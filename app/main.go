package main

import (
	"github.com/joho/godotenv"
	"github.com/nizigama/ovrsight/app/cmd"
	"log"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	cmd.Execute()
}

package main

import (
	"github.com/joho/godotenv"
	"github.com/nizigama/ovrsight/app/cli/commands"
	"log"
)

const envFile = ".env"

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	commands.Execute()
}

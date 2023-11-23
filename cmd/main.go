package main

import (
	"log"

	"github.com/Sinbad-HQ/kyc/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalln(err)
	}
}

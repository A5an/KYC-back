package main

import (
	"log"

	"github.com/Sinbad-HQ/kyc/cmd/server"
)

func main() {
	//fmt.Println(providers.FetchPublicAccessToken("https://sandbox.onebrick.io/v1", "5500802e-196c-4f32-92a0-d28ceaf99a19", "BUChl8HXmNoqKq5B8ukvv0zu23ulvy"))
	if err := server.Run(); err != nil {
		log.Fatalln(err)
	}
}

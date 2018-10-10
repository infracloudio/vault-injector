package main

import (
	"log"

	"github.com/infracloudio/vault-injector/controller/pkg/serve"
)

func main() {
	log.Println("Starting vault controller")
	serve.Serve()
}

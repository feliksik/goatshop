package main

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/feliksik/goatshop/api"
	"net/http"
)

func main() {
	app := cli.NewApp()
	app.Name = "Goat webshop"
	app.Usage = "Sell and administer goats"
	app.Version = "0.0.1"

	shop := api.NewShop()

	n := shop.GetHandler()

	port := 8080
	fmt.Printf("Serving on localhost:%d...\n", port)

	a := http.ListenAndServe(fmt.Sprintf(":%d", port), n)

	if a != nil {
		fmt.Println("Error: "+ a.Error())
	}
}

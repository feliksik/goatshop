package main



import (
	"github.com/urfave/cli"
	"gitlab.com/nerdalize/yak/api"
	"fmt"
	"net/http"
)

func main() {
	app := cli.NewApp()
	app.Name = "Yak webshop"
	app.Usage = "Sell and administer yaks"
	app.Version = "0.0.1"


	shop := api.NewShop()

	n := shop.GetHandler()

	fmt.Println("Serving...")
	http.ListenAndServe(":8080", n)
}

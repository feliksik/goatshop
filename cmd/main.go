package main

//import "encoding/xml"

import (
	"github.com/feliksik/goatshop/goat"
        "fmt"
        "os"
        "gopkg.in/urfave/cli.v1"

        "strconv"
        "encoding/xml"
)

func printStock(s *goat.Stock) {
	fmt.Println("In Stock:")
	fmt.Println(fmt.Sprintf("\t%.3f liters of milk", s.Milk))
	fmt.Println(fmt.Sprintf("\t%d skins of wool", s.Skins))
}

func printHerd(h *goat.Herd) {
	fmt.Println("Herd:")
	for _, y := range h.Goats {
		fmt.Println(fmt.Sprintf("\t%s %.2f years old", y.Name, float64(y.Age)/100))
	}
}

func createCommandlineApp() *cli.App {
	app := cli.NewApp()
	app.Name = "Goat command line util "
	app.Usage = "Administer goats or my webshop"
	app.Version = "0.0.1"
	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		fmt.Fprintf(c.App.Writer, "Usage error: %#v\n", err)
		return nil
	}

	app.Action = func(c *cli.Context) error {
		if len(c.Args()) < 2 {
			return cli.NewExitError(fmt.Sprintf("not enough arguments (try `%s help')", c.App.HelpName), 1)
		}

		fname := c.Args().Get(0)

		xmlFile, err := os.Open(fname)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Error opening file:", err), 1)
		}
		defer xmlFile.Close()


		rundays, err := strconv.Atoi(c.Args().Get(1))
		if err != nil {
			return cli.NewExitError("2nd argument must be int.", 1)
		}

		var herd *goat.Herd
		var stock *goat.Stock = &goat.Stock{}

		err = xml.NewDecoder(xmlFile).Decode(&herd)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Error decoding file:", fname), 1)

		}

		date := 0 // we start at day 0
		for ; date < rundays; date++ {
			herd.Attend(stock)
			herd.DayPasses()
		}

		printStock(stock)
		printHerd(herd)

		return nil
	}
	return app
}

func main() {
	app := createCommandlineApp()
	app.Run(os.Args)
}

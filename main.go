package main

//import "encoding/xml"
import "fmt"
import "os"
import "github.com/urfave/cli"

import "strconv"
import "encoding/xml"

import (
	"gitlab.com/nerdalize/yak/yak"
)

func printStock(s *yak.Stock) {
	fmt.Println("In Stock:")
	fmt.Println(fmt.Sprintf("\t%.3f liters of milk", s.Milk))
	fmt.Println(fmt.Sprintf("\t%d skins of wool", s.Skins))
}

func printHerd(h *yak.Herd) {
	fmt.Println("Herd:")
	for _, y := range h.Yaks {
		fmt.Println(fmt.Sprintf("\t%s %.2f years old", y.Name, float64(y.Age)/100 ) )
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Yak util"
	app.Usage = "Administer yaks"
	app.Version = "0.0.1"

	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		fmt.Fprintf(c.App.Writer, "WRONG: %#v\n", err)
		return nil
	}


	app.Action = func(c *cli.Context) error {
		if len(c.Args()) < 2 {
			return cli.NewExitError("not enough arguments", 1)
		}

		fname := c.Args().Get(0)
		rundays, err := strconv.Atoi(c.Args().Get(1))
		if err != nil {
			return cli.NewExitError("2nd argument must be int.", 1)
		}

		xmlFile, err := os.Open(fname)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Error opening file:", err), 1)
		}
		defer xmlFile.Close()

		var herd *yak.Herd
		var stock *yak.Stock = &yak.Stock{}

		err = xml.NewDecoder(xmlFile).Decode(&herd)

		if err!=nil {
			return cli.NewExitError(fmt.Sprintf("Error decoding file:", fname), 1)

		}

		date := 0; // we start at day 0
		for ; date < rundays ; date ++ {
			herd.Attend(stock)
			herd.DayPasses()
		}

		printStock(stock)
		printHerd(herd)

		return nil
	}

	app.Run(os.Args)

}

package main

import (
	"fmt"
	"os"
)

func parseArgs(args []string) {

}

func main() {
	parseArgs(os.Args)
	os.Args = []string{"q23"}
	for _, arg := range os.Args {
		fmt.Print(arg)
	}
}

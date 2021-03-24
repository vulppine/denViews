package main

import (
	"flag"
	"log"
	"os"
	"strconv"
)

var host string
var Verbose bool

type Record struct {
	rawPath      string
	Views        int `json:"views"`
	Hits         int `json:"hits"`
	validQueries []string
	// ipaddrs      *netaddr.IPSet
}

func v(i interface{}) {
	if Verbose {
		log.Println(i)
	}
}

func init() {
	flag.StringVar(&host, "host", "", "Pass the host URL to this flag. Required if not using the HOSTURL environment variable.")
	flag.BoolVar(&Verbose, "v", false, "Verbose mode.")
}

func main() {
	flag.Parse()

	if host == "" {
		if os.Getenv("HOSTURL") == "" {
			panic("no host provided, aborting")
		}

		host = os.Getenv("HOSTURL")
	}

	switch flag.Arg(0) {
	case "add":
		switch flag.Arg(1) {
		case "qurange":
			m, err := strconv.Atoi(flag.Arg(4))
			if err != nil {
				panic(err)
			}

			AddPageQueryRange(flag.Arg(2), flag.Arg(3), m)
		case "qustring":
			AddPageQueryStrings(flag.Arg(2), flag.Arg(3), flag.Args()[4:len(flag.Args())]...)
		default:
			err := AddPage(flag.Arg(1))
			if err != nil {
				panic(err)
			}
		}
	case "get":
		log.Println(GetPageStats(flag.Arg(1)))
	default:
		panic("no verb provided, aborting")
	}
}

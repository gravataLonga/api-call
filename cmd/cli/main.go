package main

import (
	"flag"
	"fmt"
	"github.com/gravataLonga/api-call/pkg"
)

var (
	baseurl = flag.String("baseurl", "", "Define base url where to make a request")
	url     = flag.String("url", "", "Define url where to make a request")
	method  = flag.String("method", "GET", "Which method")
)

func run() error {
	flag.Parse()

	options := []pkg.Option{}
	if baseurl != nil {
		options = append(options, pkg.WithBaseUrl(*baseurl))
	}

	apiCall := pkg.NewApiCall(options...)
	base, err := apiCall.Send(*method, *url, nil)

	if err != nil {
		return err
	}

	if !base.Ok {
		return fmt.Errorf("Got errors %v", base.Errors.String())
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

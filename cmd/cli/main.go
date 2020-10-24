package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gravataLonga/api-call/pkg"
)

var (
	url    = flag.String("url", "", "Define url where to make a request")
	method = flag.String("method", "GET", "Which method")
)

func run() error {
	flag.Parse()

	options := []pkg.Option{}
	if url != nil {
		options = append(options, pkg.WithUrl(*url))
	}

	if method != nil {
		options = append(options, pkg.WithMethod(*method))
	}

	apiCall := pkg.NewApiCall(options...)
	base, err := apiCall.Send()

	if err != nil {
		return err
	}

	if !base.Ok {
		return errors.New(fmt.Sprintf("Got errors %v", base.Errors.String()))
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

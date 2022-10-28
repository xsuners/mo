package main

import (
	"fmt"
	"os"

	"github.com/skyasker/go-flags"
	"github.com/xsuners/mo/internal/generator"
)

type Args struct {
	generator.Command `command:"gen"`
}

func parse(args *Args) error {
	parser := flags.NewParser(args, flags.Default)
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			fmt.Println(err)
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
	return nil
}

func main() {
	var args Args
	if err := parse(&args); err != nil {
		panic(err)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/funnydog/website/backend"
	"github.com/funnydog/website/config"
	"github.com/funnydog/website/sync"
	"github.com/funnydog/website/template"
	"github.com/pborman/getopt/v2"
)

var (
	helpFlag   bool
	renderFlag bool
	uploadFlag bool
)

func init() {
	getopt.Flag(&helpFlag, 'h')
	getopt.Flag(&renderFlag, 'r')
	getopt.Flag(&uploadFlag, 'u')
}

func main() {
	getopt.Parse()
	if helpFlag {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Print("Options:\n")
		fmt.Println("\t-h\tdisplay the help text")
		fmt.Println("\t-r\trender the templates on filesystem")
		fmt.Println("\t-u\tupload the files to a remote host")
		return
	}
	conf, err := config.Read("config.json")
	if err != nil {
		panic(err)
	}

	if renderFlag {
		err = template.Render(&conf)
		if err != nil {
			panic(err)
		}

		err = template.CopyStatic(&conf)
		if err != nil {
			panic(err)
		}
	}

	if uploadFlag {
		backend, err := backend.Create(&conf)
		if err != nil {
			panic(err)
		}
		defer backend.Close()

		b, err := sync.NewSync(
			backend,
			conf.ChecksumName,
			conf.RenderDir,
		)
		if err != nil {
			panic(err)
		}

		if err = b.Synchronize(); err != nil {
			panic(err)
		}
	}
}

package main

import (
	"io"
	"log"

	"github.com/salberternst/workspace/pkg/cmd"
)

func main() {
	log.SetOutput(io.Discard)

	cmd.Execute()
}

package main

import (
	"github.com/allentom/haruka/cmd/generator"
	"log"

	"github.com/spf13/cobra"
)

func main() {
	log.SetFlags(0)
	cmd := &cobra.Command{
		Use: "haruka",
	}
	cmd.AddCommand(generator.GeneratorCmd)
	_ = cmd.Execute()
}

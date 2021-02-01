package main

import (
	"github.com/allentom/haruka/haruka/generator"
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

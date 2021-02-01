package generator

import (
	"github.com/spf13/cobra"
)

var ModelName string
var ModelType string
var GeneratorCmd = &cobra.Command{
	Use:              "generator",
	Short:            "haruka code generator",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(ModelName) > 0 && ModelType == "rest" {
			err := GenerateRestModel(ModelName)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	GeneratorCmd.Flags().StringVarP(&ModelName, "name", "n", "", "model name")
	GeneratorCmd.Flags().StringVarP(&ModelType, "type", "t", "", "model type")
}

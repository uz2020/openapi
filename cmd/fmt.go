/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

// fmtCmd represents the fmt command
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: format,
}

func init() {
	rootCmd.AddCommand(fmtCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fmtCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fmtCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func format(cmd *cobra.Command, args []string) {
	for _, arg := range args {
		from, err := openapi3.NewLoader().LoadFromFile(arg)
		check(err)
		j, err := from.MarshalJSON()
		check(err)
		y, err := yaml.JSONToYAML(j)
		check(err)

		err = os.WriteFile(arg, y, 0644)
		check(err)
	}
}

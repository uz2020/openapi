/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

// yankCmd represents the yank command
var yankCmd = &cobra.Command{
	Use:   "yank",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: yank,
}

var From string
var To string

func init() {
	rootCmd.AddCommand(yankCmd)
	yankCmd.Flags().StringVarP(&From, "from", "f", "", "From openapi.yaml (required)")
	yankCmd.MarkFlagRequired("from")
	yankCmd.Flags().StringVarP(&To, "to", "t", "", "To openapi.yaml (required)")
	yankCmd.MarkFlagRequired("to")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// yankCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// yankCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type T struct {
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parseMap(aMap map[string]interface{}, from, to *openapi3.T) {
	for key, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			parseMap(val.(map[string]interface{}), from, to)
		case []interface{}:
			parseArray(val.([]interface{}), from, to)
		default:
			if key == "$ref" {
				name := getRefName(concreteVal.(string))
				yankSchema(name, from, to)
			}
		}
	}
}

func parseArray(anArray []interface{}, from, to *openapi3.T) {
	for _, val := range anArray {
		switch val.(type) {
		case map[string]interface{}:
			parseMap(val.(map[string]interface{}), from, to)
		case []interface{}:
			parseArray(val.([]interface{}), from, to)
		}
	}
}

func getRefName(ref string) string {
	seq := strings.Split(ref, "/")
	return seq[len(seq)-1]
}

func yankSchema(name string, from, to *openapi3.T) {
	schema := from.Components.Schemas[name]
	if _, ok := to.Components.Schemas[name]; ok {
		return
	}
	to.Components.Schemas[name] = schema

	j, err := schema.MarshalJSON()
	check(err)

	m := map[string]interface{}{}

	// Parsing/Unmarshalling JSON encoding/json
	err = json.Unmarshal(j, &m)
	check(err)

	parseMap(m, from, to)
}

func yankOp(op *openapi3.Operation, from, to *openapi3.T) {
	if op == nil {
		return
	}

	for _, tag := range op.Tags {
		if to.Tags.Get(tag) == nil {
			to.Tags = append(to.Tags, from.Tags.Get(tag))
		}
	}

	if op.Parameters != nil {
		for _, v := range op.Parameters {
			if v.Ref == "" {
				continue
			}

			name := v.Value.Name
			if _, ok := to.Components.Parameters[name]; ok {
				// 已经有定义
				continue
			}
			to.Components.Parameters[name] = from.Components.Parameters[name]
		}
	}

	if op.RequestBody != nil {
		if op.RequestBody.Ref != "" {
			name := getRefName(op.RequestBody.Ref)
			if _, ok := to.Components.RequestBodies[name]; !ok {
				target := from.Components.RequestBodies[name]
				to.Components.RequestBodies[name] = target

				schemaRef := target.Value.Content.Get("application/json").Schema.Ref
				if schemaRef != "" {
					name = getRefName(schemaRef)
					yankSchema(name, from, to)
				}
			}
		}
	}

	for _, response := range op.Responses {
		if response.Ref == "" {
			continue
		}

		name := getRefName(response.Ref)
		if _, ok := to.Components.Responses[name]; ok {
			continue
		}
		target := from.Components.Responses[name]
		to.Components.Responses[name] = target

		schemaRef := target.Value.Content.Get("application/json").Schema.Ref
		if schemaRef == "" {
			continue
		}
		name = getRefName(schemaRef)
		yankSchema(name, from, to)
	}
}

func yank(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("no interface")
		return
	}

	from, err := openapi3.NewLoader().LoadFromFile(From)
	check(err)

	to, err := openapi3.NewLoader().LoadFromFile(To)
	check(err)

	for _, arg := range args {
		path := from.Paths[arg]
		to.Paths[arg] = path

		yankOp(path.Get, from, to)
		yankOp(path.Delete, from, to)
		yankOp(path.Post, from, to)
		yankOp(path.Patch, from, to)
		yankOp(path.Put, from, to)
	}

	j, _ := to.MarshalJSON()
	y, _ := yaml.JSONToYAML(j)
	fmt.Println(string(y))
}

/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// transformCmd represents the transform command
var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Performs Fhirbase transformation on a single FHIR resource loaded from a JSON file",
	Long: `
Transform command applies Fhirbase transformation algorithm to a
single FHIR resource loaded from provided JSON file and outputs result
to the STDOUT. This command exists mostly for demonstration and
debugging of Fhirbase transformation logic.

For detailed explanation of Fhirbase transformation algorithm please
proceed to the Fhirbase documentation. TODO: direct documentation
link.`,
	Example: "fhirbase [--fhir=FHIR version] transform path/to/fhir-resource.json",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			err := TransformCommand(cmd, arg)
			if err != nil {
				log.Fatalf("Error transforming resource: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(transformCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transformCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transformCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
func getByPath(tr map[string]interface{}, path []interface{}) map[string]interface{} {
	res := tr

	for _, k := range path {
		key := k.(string)
		res = res[key].(map[string]interface{})

		if res == nil {

			log.Printf("cannot get trNode by path: %v", path)

			return nil

		}

	}

	return res

}

func transform(node interface{}, trNode map[string]interface{}, tr map[string]interface{}) (interface{}, error) {

	// log.Printf("=> %v %v", node, trNode)

	_, isSlice := node.([]interface{})

	if trNode["tr/act"] != nil && !isSlice {

		var res interface{}
		trAct := trNode["tr/act"].(string)

		if trAct == "union" {
			args := trNode["tr/arg"].(map[string]interface{})
			ttype := args["type"].(string)
			transformed := make(map[string]interface{})

			if tr[ttype] != nil {

				if ttype == "Reference" {
					r, _ := transform(node, map[string]interface{}{"tr/act": "reference"}, tr)
					transformed[ttype] = r
				} else {
					r, _ := transform(node, tr[ttype].(map[string]interface{}), tr)
					transformed[ttype] = r
				}

			} else {
				transformed[ttype] = node
			}
			res = transformed

		} else if trAct == "reference" {
			v := node.(map[string]interface{})
			newref := make(map[string]interface{})
			if v["reference"] != nil {
				refstr, _ := v["reference"].(string)
				refcomps := strings.Split(refstr, "/")

				if len(refcomps) == 2 {
					newref["id"] = refcomps[1]
					newref["resourceType"] = refcomps[0]
				} else {
					newref["id"] = refstr
				}

			}

			if v["display"] != nil {
				newref["display"] = v["display"].(string)
			}

			res = newref

		}

		return res, nil

	}

	switch node.(type) {

	case map[string]interface{}:

		res := make(map[string]interface{})

		for k, v := range node.(map[string]interface{}) {

			if (trNode != nil) && (trNode[k] != nil) {

				nextTrNode := trNode[k].(map[string]interface{})

				args := nextTrNode["tr/arg"]

				key := k

				if args != nil {

					argsMap := args.(map[string]interface{})

					if argsMap != nil {

						key = argsMap["key"].(string)

					}

				}

				if nextTrNode["tr/move"] != nil {

					nextTrNode = getByPath(tr, nextTrNode["tr/move"].([]interface{}))

				}

				r, _ := transform(v, nextTrNode, tr)

				res[key] = r

			} else {

				r, _ := transform(v, nil, tr)

				res[k] = r

			}

		}

		return res, nil

	case []interface{}:

		res := make([]interface{}, 0, 8)

		for _, v := range node.([]interface{}) {

			r, _ := transform(v, trNode, tr)

			res = append(res, r)

		}

		return res, nil

	default:

		return node, nil

	}

}

var transformDatas = make(map[string]interface{})

//go:embed transform/*

var transformFiles embed.FS

func getTransformData(fhirVersion string) (map[string]interface{}, error) {

	if transformDatas[fhirVersion] != nil {

		return transformDatas[fhirVersion].(map[string]interface{}), nil

	}

	filename := fmt.Sprintf("fhirbase-import-%s.json", fhirVersion)

	filepath := path.Join("transform", filename)

	trData, err := transformFiles.ReadFile(filepath)

	if err != nil {

		return nil, fmt.Errorf("cannot find transformations data for FHIR version %s", fhirVersion)

	}

	iter := jsoniter.ConfigFastest.BorrowIterator(trData)

	defer jsoniter.ConfigFastest.ReturnIterator(iter)

	tr := iter.Read().(map[string]interface{})

	if tr == nil {

		return nil, fmt.Errorf("cannot parse transformations data for FHIR version %s", fhirVersion)

	}

	transformDatas[fhirVersion] = tr

	return tr, nil

}

func doTransform(res map[string]interface{}, fhirVersion string) (map[string]interface{}, error) {

	tr, err := getTransformData(fhirVersion)

	if err != nil {

		return nil, fmt.Errorf("cannot get transformations data for FHIR version %s: %v", fhirVersion, err)

	}

	rt, ok := res["resourceType"].(string)

	if !ok {

		return nil, fmt.Errorf("cannot determine resourceType for resource %v", res)

	}

	trNode := tr[rt]

	if trNode == nil {

		// TODO: some warning output here?

		return res, nil

	}

	trNodeMap := trNode.(map[string]interface{})

	out, err := transform(res, trNodeMap, tr)

	if err != nil {

		return nil, fmt.Errorf("error transforming resource: %v", err)

	}

	outMap, ok := out.(map[string]interface{})

	if !ok {

		return nil, fmt.Errorf("incorrect format after transformation: %v", out)

	}

	return outMap, nil

}

// TransformCommand transforms FHIR resource to internal JSON representation

func TransformCommand(c *cobra.Command, arg string) error {

	fhirVersion := viper.GetString("fhirVersion")

	file, err := os.Open(arg)
	if err != nil {
		return fmt.Errorf("Cannot open file %s: %v", arg, err)
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Cannot read file %s: %v", arg, err)
	}

	iter := jsoniter.ConfigFastest.BorrowIterator(fileContent)

	defer jsoniter.ConfigFastest.ReturnIterator(iter)

	res := iter.Read()

	if res == nil {

		return fmt.Errorf("Cannot parse JSON from file %s", arg)

	}

	out, err := doTransform(res.(map[string]interface{}), fhirVersion)

	if err != nil {

		return fmt.Errorf("Error transforming resource: %v", err)

	}

	outJson, err := jsoniter.ConfigFastest.MarshalIndent(out, "", " ")

	os.Stdout.Write(outJson)

	os.Stdout.Write([]byte("\n"))

	return nil

}

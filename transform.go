package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gobuffalo/packr"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

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
	log.Printf("=> %v %v", node, trNode)

	_, isSlice := node.([]interface{})

	if trNode["tr/act"] != nil && !isSlice {
		var res interface{}
		trAct := trNode["tr/act"].(string)

		if trAct == "union" {
			args := trNode["tr/arg"].(map[string]interface{})
			ttype := args["type"].(string)
			transformed := make(map[string]interface{})

			if tr[ttype] != nil {
				r, _ := transform(node, tr[ttype].(map[string]interface{}), tr)
				transformed[ttype] = r
			} else {
				transformed[ttype] = node
			}

			res = transformed
		} else if trAct == "reference" {
			v := node.(map[string]interface{})

			transformed := make(map[string]interface{})
			refcomps := strings.Split(v["reference"].(string), "/")
			newref := make(map[string]interface{})
			newref["id"] = refcomps[len(refcomps)-1]
			newref["type"] = refcomps[len(refcomps)-2]

			if v["display"] != nil {
				newref["display"] = v["display"].(string)
			}

			transformed["reference"] = newref
			res = transformed
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

func getTransformData(fhirVersion string) (map[string]interface{}, error) {
	if transformDatas[fhirVersion] != nil {
		return transformDatas[fhirVersion].(map[string]interface{}), nil
	}

	box := packr.NewBox("./transform")
	trData, err := box.MustBytes(fmt.Sprintf("fhirbase-import-%s.json", fhirVersion))

	if err != nil {
		return nil, errors.Wrapf(err, "failed to find transformations data for FHIR version %s", fhirVersion)
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
		return nil, errors.Wrap(err, "cannot get transform data")
	}

	rt, ok := res["resourceType"].(string)

	if !ok {
		return nil, fmt.Errorf("cannot determine resourceType for resource %v", res)
	}

	trNode := tr[rt].(map[string]interface{})

	out, err := transform(res, trNode, tr)

	if err != nil {
		return nil, errors.Wrapf(err, "cannot perform transformation for resource %v", res)
	}

	outMap, ok := out.(map[string]interface{})

	if !ok {
		return nil, fmt.Errorf("incorrect format after transformation: %v", out)
	}

	return outMap, nil
}

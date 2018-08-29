package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"
)

func parseJson(str string) map[string]interface{} {
	iter := jsoniter.ConfigFastest.BorrowIterator([]byte(str))
	defer jsoniter.ConfigFastest.ReturnIterator(iter)

	return iter.Read().(map[string]interface{})

}

func encodeJson(m map[string]interface{}) string {
	stream := jsoniter.ConfigFastest.BorrowStream(nil)
	defer jsoniter.ConfigFastest.ReturnStream(stream)

	stream.WriteVal(m)
	return string(stream.Buffer())
}

var cases = [][]string{
	[]string{
		`
{
"resourceType":"CarePlan",
"careTeam": [
  {"reference": "Practitioner/1", "display": "John"},
  {"reference": "Practitioner/2", "display": "Ian"}
],
"identifier": [
  {"system": "foo", "value": "bar"},
  {"system": "foo", "value": "baz", "assigner": { "reference": "Practitioner/42", "display": "John Doe"}}
]}`,
		`{
"resourceType":"CarePlan",
"careTeam":[
  {"reference":{"id":"1","type":"Practitioner","display":"John"}},
  {"reference":{"id":"2","type":"Practitioner","display":"Ian"}}
],
"identifier":[
  {"system":"foo","value":"bar"},
  {"system":"foo","value":"baz","assigner":{"reference":{"id":"42","type":"Practitioner","display":"John Doe"}}}
]}`,
	}, []string{
		`{
  "resourceType":"Patient",
  "name": [{"given": ["Mike"], "family": "Lapshin"}],
  "deceasedBoolean": true,
  "managingOrganization": { "reference": "Organization/1", "display": "ACME corp"}
}`, `{
  "managingOrganization":{"reference":{"id":"1","type":"Organization","display":"ACME corp"}},
  "resourceType":"Patient",
  "deceased": { "boolean": true },
  "name":[{"family":"Lapshin","given":["Mike"]}],
}`,
	}, []string{
		`{"resourceType":"Patient", "managingOrganization": { "reference": "Organization/1", "display": "ACME corp"}}`,
		`{"resourceType":"Patient", "managingOrganization":{"reference":{"id":"1","type":"Organization","display":"ACME corp"}}}`,
	},
}

func TestTransform(t *testing.T) {
	for i, c := range cases {
		t.Run(fmt.Sprintf("Transform case #%d", i), func(t *testing.T) {
			in := parseJson(c[0])
			out := parseJson(c[1])

			result, err := doTransform(in, "3.3.0")

			if err != nil {
				t.Error(err)
			}

			if !cmp.Equal(out, result) {
				t.Logf("Expected:\n%v\nGot:\n%v\nDifference:\n%v", out, result, cmp.Diff(result, out))
				t.Fail()
			}
		})
	}
}

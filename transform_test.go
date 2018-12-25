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
  {"id":"1","resourceType":"Practitioner","display":"John"},
  {"id":"2","resourceType":"Practitioner","display":"Ian"}
],
"identifier":[
  {"system":"foo","value":"bar"},
  {"system":"foo","value":"baz","assigner":{"id":"42","resourceType":"Practitioner","display":"John Doe"}}
]}`,
	},
	[]string{
		`
{
"resourceType":"Claim",
"information": [
  {"valueReference": {"reference": "Immunization/123"}}
]}`,
		`{
"resourceType":"Claim",
"information": [
  {"value": {"Reference": { "resourceType": "Immunization", "id": "123" }}}
]}`,
	},
	[]string{
		`{
  "resourceType":"Patient",
  "name": [{"given": ["Mike"], "family": "Lapshin"}],
  "deceasedBoolean": true,
  "multipleBirthInteger": 2,
  "managingOrganization": { "reference": "Organization/1", "display": "ACME corp"}
}`, `{
  "managingOrganization":{"id":"1","resourceType":"Organization","display":"ACME corp"},
  "resourceType":"Patient",
  "deceased": { "boolean": true },
  "multipleBirth": { "integer": 2 },
  "name":[{"family":"Lapshin","given":["Mike"]}],
}`,
	}, []string{
		`{"resourceType":"Patient", "managingOrganization": { "display": "ACME corp"}}`,
		`{"resourceType":"Patient", "managingOrganization":{"display":"ACME corp"}}`,
	},
	[]string{
		`{"resourceType":"FoobarUnknown", "foo": 42}`,
		`{"resourceType":"FoobarUnknown", "foo": 42}`,
	},
}

func TestTransform(t *testing.T) {
	for i, c := range cases {
		t.Run(fmt.Sprintf("Transform case #%d", i), func(t *testing.T) {
			in := parseJson(c[0])
			out := parseJson(c[1])

			result, err := doTransform(in, "3.0.1")

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

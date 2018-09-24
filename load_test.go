package main

import (
	"fmt"
	"strings"
	"testing"
)

var fileTypeCases = map[string]bundleType{
	"{\"foo\": \"bar\"}\n{\"foo\": \"bar\"}":                           ndjsonBundleType,
	"{\"foo\": \"{{\\\"}bar\"}\n{\"foo\": \"bar\"}":                    ndjsonBundleType,
	"{\"foo\": \"{{\\\"}bar\",\n\n\"resourceType\": \"Bundle\"}":       fhirBundleType,
	"{\"foo\": \"bar\", \n\n\n\n\n \"resourceType\": \"Observation\"}": singleResourceBundleType,
	"{\"foo\": \"{{\\\"}bar\", \"resourceType\": \"Patient\"}":         singleResourceBundleType,
}

func TestGuessBundleType(t *testing.T) {
	i := 0
	for str, tpe := range fileTypeCases {
		i++

		t.Run(fmt.Sprintf("File type case #%v", str), func(t *testing.T) {
			bt, err := guessBundleType(strings.NewReader(str))

			if err != nil {
				t.Error(err)
			}

			if bt != tpe {
				t.Logf("bundle type not matched (expected %v, got %v)", tpe, bt)
				t.Fail()
			}

		})
	}

}

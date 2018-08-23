## fhirbase loader

* Bulk API most important
* Conversion
* [Validation]

## Techonology

* python (standard and easy development)
* jvm/clojure (reuse aidbox)
* golang (binary)
* rust (fast)


## What do we want?

* fhirbase as self-contained cli tool (worse is docker container)
* fast - benchmark???
* quick dev

## Process 

* load (rest) - no intermediate file
* convert
* [validate]
* insert (postgres)


## experiment

* nd-json (file system / http)
* stream read
* id, resourceType (convert)
* copy insert

# Fhirbase 3.0

[![Build Status](https://travis-ci.org/fhirbase/fhirbase-core.svg?branch=master)](https://travis-ci.org/fhirbase/fhirbase-core)

* [Documentation](https://fhirbase.gitbook.io/project/)
* [Demo](http://fhirbase.github.io/)
* [Chat](https://chat.fhir.org/#narrow/stream/16-fhirbase)
* [Google Group](https://groups.google.com/forum/#!forum/fhirbase)
* [StackOverflow](???)

## Development


For macos:

```
brew install go
brew install dep

# use your local postgres or
cd dev
source .env
docker-compose up -d
cd ..

make

source dev/.env
bin/fhirbase -d fhirbase init
curl https://storage.googleapis.com/aidbox-public/sample-data.gzip > /tmp/data.gzip 
bin/fhirbase -d fhirbase load /tmp/data.gzip

```

This project is Makefile-based. At first, you need to install Golang
and [Glide](https://github.com/Masterminds/glide). Then building entire project is as simple as:

    $ make

Other possible build targets are:

    $ make fmt     # runs go fmt
    $ make lint    # runs golint
    $ make vendor  # runs Glide to install dependencies
    $ make clean   # cleanups everything

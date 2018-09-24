# Fhirbase

**[Download the Latest Release](https://github.com/fhirbase/fhirbase/releases/tag/nightly-build)**&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;**[Try Online](https://fbdemo.aidbox.app/)**&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Documentation](https://fhirbase.gitbook.io/project/)&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Chat](https://chat.fhir.org/#narrow/stream/16-fhirbase)&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Google Group](https://groups.google.com/forum/#!forum/fhirbase)

[![Build Status](https://travis-ci.org/fhirbase/fhirbase.svg?branch=master)](https://travis-ci.org/fhirbase/fhirbase)

Fhirbase is a command-line utility which enables you to easily import
[FHIR data](https://www.hl7.org/fhir/) into a PostgreSQL database and
work with it in a relational way. Also Fhirbase provides set of stored
procedures to perform [CRUD
operations](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete)
and mantain [Resources
History](https://www.hl7.org/fhir/http.html#history).

<p align="center">
    <img src="https://cdn.rawgit.com/fhirbase/fhirbase/a6aff815/demo/asciicast.svg" />
</p>

## Getting Started

Please proceed to the [Getting
Started](https://fhirbase.gitbook.io/project/getting-started) tutorial
for PostgreSQL and Fhirbase installation instructions.


## Development

To participate in Fhirbase development you'll need to install Golang
and [Dep package
manager](https://golang.github.io/dep/docs/installation.html).

Fhirbase is Makefile-based project, so building it is as simple as
invoking `make` command.

NB you can put Fhirbase source code outside of `GOPATH` env variable
because Makefile sets `GOPATH` value to `fhirbase-root/.gopath`.

To enable hot reload of demo's static assets set `DEV` env variable
like this:

```
DEV=1 fhirbase web
```

## License

Copyright © 2018 [Health Samurai](https://www.health-samurai.io/) team.

Fhirbase is released under the terms of the MIT License.

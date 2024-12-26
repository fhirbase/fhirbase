# STATUS

This is an update to fhirbase to fix some errors that were present in the most
recent version.

## Fixes:

- Segmentation fault error on data load (pointer to db pool could be nil)
- Crash on database initiatation if it was run again. (added check for
  \_resource type prior to creation)

## Enhancements

- Updated to Go 1.23
- Migrated from packr to embed for static files
- Updated all dependencies to current versions
- Replaced `urfave/cli` with `cobra`
- Added `viper` for configuration
- Refactored retry function on Bulk downloader to utilize exponential backoff
  timing based on response headers from FHIR server

# Fhirbase

[Latest Release on labordude's Github](https://github.com/labordude/fhirbase/releases)

**[Download the Latest Release](https://github.com/fhirbase/fhirbase/releases/)**&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;**[Try Online](https://fbdemo.aidbox.app/)**&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Documentation](https://aidbox.gitbook.io/fhirbase/)&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Chat](https://chat.fhir.org/#narrow/stream/16-fhirbase)&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Google Group](https://groups.google.com/forum/#!forum/fhirbase)

[![Build Status](https://travis-ci.org/fhirbase/fhirbase.svg?branch=master)](https://travis-ci.org/fhirbase/fhirbase)

Fhirbase is a command-line utility which enables you to easily import
[FHIR data](https://www.hl7.org/fhir/) into a PostgreSQL database and work with
it in a relational way. Also Fhirbase provides set of stored procedures to
perform
[CRUD operations](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete)
and maintain [Resources History](https://www.hl7.org/fhir/http.html#history).

<p align="center">
    <img src="https://cdn.rawgit.com/fhirbase/fhirbase/a6aff815/demo/asciicast.svg" />
</p>

## Getting Started

Please proceed to the
[Getting Started](https://fhirbase.aidbox.app/getting-started) tutorial for
PostgreSQL and Fhirbase installation instructions.

## Usage Statistics - These have been removed.

## Development

To participate in Fhirbase development you'll need to install Golang and
[Dep package manager](https://golang.github.io/dep/docs/installation.html).

Fhirbase is Makefile-based project, so building it is as simple as invoking
`make` command.

NB you can put Fhirbase source code outside of `GOPATH` env variable because
Makefile sets `GOPATH` value to `fhirbase-root/.gopath`.

To enable hot reload of demo's static assets set `DEV` env variable like this:

```
DEV=1 fhirbase web
```

## License

Copyright © 2018 [Health Samurai](https://www.health-samurai.io/) team.

Fhirbase is released under the terms of the MIT License.

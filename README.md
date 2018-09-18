# Fhirbase 3.0

**[Download the Latest Release](https://github.com/fhirbase/fhirbase/releases/tag/nightly-build)**&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Documentation](https://fhirbase.gitbook.io/project/)&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Demo](http://fhirbase.github.io/)&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Chat](https://chat.fhir.org/#narrow/stream/16-fhirbase)&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Google Group](https://groups.google.com/forum/#!forum/fhirbase)

[![Build Status](https://travis-ci.org/fhirbase/fhirbase.svg?branch=master)](https://travis-ci.org/fhirbase/fhirbase)

Fhirbase is a command-line utility wich enables you to easily import
[FHIR data](https://www.hl7.org/fhir/) into a PostgreSQL database and
work with it in a relational way. Also Fhirbase provides set of stored
procedures to perform [CRUD
operations](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete)
and mantain [Resources
History](https://www.hl7.org/fhir/http.html#history).

# Getting Started

To start using Fhirbase you have to install [PostgreSQL
Database](https://www.postgresql.org/) first. Any version above
**9.6** should be fine.  Please proceed to the section describing
operating system you're running.

## Windows

For Windows the most simple way is to use EnterpriseDB PostgreSQL
Installer. Also there is a [YouTube
Video](https://www.youtube.com/watch?v=e1MwsT5FJRQ) demonstrating the
whole installation process.

## Linux

Most likely that PostgreSQL will be available through your
distribition's package manager. On Ubuntu or Debian, it's as easy as

    $ sudo apt-get install postgresql postgresql-client

For other Linux distributions, please use search functionality of your
package manager or just Google right command.

## MacOS

You can install PostgreSQL with [Homebrew](https://brew.sh/):

    $ brew install postgresql

As an alternative, there is a [Postgres.app](https://postgresapp.com/)
project which provides PostgreSQL as a regular MacOS application with
common drag-and-drop installation.

## Docker (cross-platform)

If you have [Docker](https://www.docker.com/) installed, you might
want to start PostgreSQL as a Docker container:

    $ docker run --name fhirbase-postgres -p=5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres:latest

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

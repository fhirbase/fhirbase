# Fhirbase 3.0

**[Download the Latest Release](https://github.com/fhirbase/fhirbase/releases/tag/nightly-build)**&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;**[Try Online](https://fbdemo.aidbox.app/)**&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Documentation](https://fhirbase.gitbook.io/project/)&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Chat](https://chat.fhir.org/#narrow/stream/16-fhirbase)&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;[Google Group](https://groups.google.com/forum/#!forum/fhirbase)

[![Build Status](https://travis-ci.org/fhirbase/fhirbase.svg?branch=master)](https://travis-ci.org/fhirbase/fhirbase)

Fhirbase is a command-line utility wich enables you to easily import
[FHIR data](https://www.hl7.org/fhir/) into a PostgreSQL database and
work with it in a relational way. Also Fhirbase provides set of stored
procedures to perform [CRUD
operations](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete)
and mantain [Resources
History](https://www.hl7.org/fhir/http.html#history).

## Installing PostgreSQL

To start using Fhirbase you have to install [PostgreSQL
database](https://www.postgresql.org/) first. Any version above
**9.6** should be fine.  Please proceed to the section describing
operating system you're running.

### Windows

For Windows the most simple way is to use [EnterpriseDB PostgreSQL
Installer](https://www.enterprisedb.com/downloads/postgres-postgresql-downloads). Also
there is a [YouTube
video](https://www.youtube.com/watch?v=e1MwsT5FJRQ) demonstrating the
whole installation process.

### Linux

Most likely that PostgreSQL will be available through your
distribition's package manager. On Ubuntu or Debian, it's as easy as

    $ sudo apt-get install postgresql postgresql-client

For other Linux distributions, please use search functionality of your
package manager or just Google the right command.

### MacOS

You can install PostgreSQL with [Homebrew](https://brew.sh/):

    $ brew install postgresql

As an alternative, there is a [Postgres.app](https://postgresapp.com/)
project which provides PostgreSQL as a regular MacOS application with
common drag-and-drop installation.

### Docker (cross-platform)

If you have [Docker](https://www.docker.com/) installed, you might
want to start PostgreSQL as a Docker container:

    $ docker run --name fhirbase-postgres -p=5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres:latest

## Checking Postgres connection

After you finished installing Postgres, please check it's running and
accepting TCP/IP connections.

You can do this with `psql` command-line client or with
[pgAdmin](https://www.pgadmin.org/) GUI application. For `psql`, use
following command:

```
$ psql -h localhost -p 5432 -U postgres -W postgres
Password for user postgres: xxxxxxxxxx
psql (10.4, server 9.6.3)
Type "help" for help.

postgres=# _
```

If you got `postgres=#` prompt waiting for your input, your
installation is fine. You can quit `psql` typing `\q` followed by
Enter.

If you got an error like this instead:

```
psql: could not connect to server: Connection refused
        Is the server running on host "localhost" (127.0.0.1) and accepting
        TCP/IP connections on port 5432?
```

then it looks like PostgreSQL is either not running or not accepting
TCP/IP connections. Please check presence of `postgres` process using
your operating system's Process Viewer (or Task Manager). If `postges`
process is running, please consider updating `pg_hba.conf` file as
described in this [wiki
article](https://wiki.postgresql.org/wiki/Client_Authentication).

TODO: better troubleshooting guide.

## Downloading and installing Fhirbase

Go to [GitHub Releases
page](https://github.com/fhirbase/fhirbase/releases) and download most
recent release of Fhirbase. Make sure you've picked right file, it
should match your operating system and CPU architecture. Nightly-build
release is being updated after every commit to master branch, so it
might be unstable.

When you downloaded Fhirbase binary, rename it from `fhirbase-os-arch`
for to just `fhirbase` and put in some directory listed in you `$PATH`
environment variable.

If you don't know what the `$PATH` variable is, it's ok, you can skip
this step. But you'll have to type full path to the `fhirbase`
executable every time you'll invoke Fhirbase command. So intead of typing

    $ fhirbase help

you'll need to type something like this:

    $ /Users/xxxx/Downloads/fhirbase help

So please remember that if your shell says it cannot find `fhirbase`
command, most likely `fhirbase` binary wasn't correctly placed inside
`$PATH`.

## Creating and initializing database

Next step is to create database where we're going to store FHIR
data. Again, we can use `psql` command-line client or some GUI client
like pgAdmin. Connect to Postgres as you previously did:

```
$ psql -h localhost -p 5432 -U postgres -W postgres
```

To create database there is a `CREATE DATABASE` statement:

```
postgres=# CREATE DATABASE fhirbase;
```

Do not forget to put semicolon at the end of statement. Of course, you
can change "fhirbase" to any database name you want.

If you got `CREATE DATABASE` text as a response to your command, your
database was succesfully created. Type `\q` to quit `psql`.

Now we have database, but before loading any FHIR data into it we need
to create schema - some tables and stored procedures. This operation
is performed with `fhirbase init` command:

```
fhirbase -h localhost -p 5432 -d fhirbase -U postgres -W postgres --fhir=3.3.0 init
```

Let's briefly describe meaning of command-line arguments:

* `-h localhost` specifies host where Postgres is running
* `-p 5432` specifies Postgres port
* `-d fhirbase` specifies name of the database to connect to (don't forget to change it if you named database differently)
* `-U postgres` name of Postgres user, "postgres" is fine for most cases
* `-W postgres` user's password
* `--fhir=3.3.0` FHIR version you're going to use

If you got no error, database was initialized. Most common errors are
connection errors (check your Postgres credentials) and SQL errors
(you're trying to initialize non-empty database).

## Importing FHIR data

We're ready to import data into our newly created database. We can
import data from two different sources:

* local files (ndjson files)
* [Bulk Data API](https://github.com/smart-on-fhir/fhir-bulk-data-docs) endpoints

Let's try to use local file first. Fhirbase GitHub repository contains
sample file, go ahead and [download
it](https://github.com/fhirbase/fhirbase/blob/master/demo/bundle.ndjson.gzip). When
this file is on your local filesystem, invoke `fhirbase load` command:

```
fhirbase -h localhost -p 5432 -d fhirbase -U postgres -W postgres --fhir=3.3.0 load -m insert /path/to/downloaded/file/bundle.ndjson.gzip
```

Usually it takes from 25 to 45 seconds to import this file. At the end of the input Fhirbase will output statistics about inserted data:

```
Done, inserted 127454 resources in 27 seconds:
       ImagingStudy 135
 AllergyIntolerance 272
              Claim 17937
               Goal 1100
          Encounter 15384
          Procedure 14109
       Immunization 6548
           CarePlan 1544
              Basic 600
       Organization 768
   DiagnosticReport 4143
  MedicationRequest 2553
            Patient 600
          Condition 4142
        Observation 57619
```

Also you can note `-m insert` argument in previous command, it forces
Fhirbase to use `INSERT` statements instead of `COPY` statement. For
some bundles you'll want to use `INSERT` and for some `COPY` will be
faster. You can find the difference in the help section of the `load`
command:

```
fhirbase help load
```

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

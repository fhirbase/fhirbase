FROM clkao/postgres-plv8:9.6-2.0

RUN apt-get update -q && apt-get install -y -q wget

RUN wget https://github.com/fhirbase/fhirbase-core/releases/download/nightly-build/fhirbase-linux-amd64 && \
    mv fhirbase-linux-amd64 fhirbase &&  \
    chmod +x fhirbase

COPY bundle.ndjson.gzip .

RUN mkdir /pgdata && chown postgres:postgres /pgdata

USER postgres

RUN PGDATA=/pgdata /docker-entrypoint.sh postgres & until psql -U postgres -c '\q'; do \
        >&2 echo "Postgres is starting up..."; \
        sleep 5; \
    done && \
    psql -U postgres -c 'create database fhirbase;' && \
    ./fhirbase -d fhirbase init && \
    ./fhirbase -d fhirbase load --mode=insert ./bundle.ndjson.gzip

EXPOSE 3000

CMD pg_ctl -D /pgdata start && until psql -U postgres -c '\q'; do \
        >&2 echo "Postgres is starting up..."; \
        sleep 5; \
    done && \
    /fhirbase -d fhirbase web

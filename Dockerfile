FROM postgres:10.5

COPY demo/bundle.ndjson.gzip .
COPY bin/fhirbase-linux-amd64 /usr/bin/fhirbase

RUN chmod +x /usr/bin/fhirbase

RUN mkdir /pgdata && chown postgres:postgres /pgdata

USER postgres

RUN PGDATA=/pgdata /docker-entrypoint.sh postgres & until psql -U postgres -c '\q'; do \
        >&2 echo "Postgres is starting up..."; \
        sleep 5; \
    done && \
    psql -U postgres -c 'create database fhirbase;' && \
    fhirbase -d fhirbase init && \
    fhirbase -d fhirbase load --mode=insert ./bundle.ndjson.gzip

EXPOSE 3000

CMD pg_ctl -D /pgdata start && until psql -U postgres -c '\q'; do \
        >&2 echo "Postgres is starting up..."; \
        sleep 5; \
    done && \
    exec fhirbase -d fhirbase web

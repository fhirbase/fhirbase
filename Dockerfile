# Build with modules
FROM golang:1.12.1-alpine as builder
RUN apk add --no-cache ca-certificates curl git build-base

# Get dependencies first for docker caching
WORKDIR /fhirbase
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy source
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags '-w -extldflags "-static"'

FROM postgres:10.5
WORKDIR /fhirbase

COPY --from=builder /fhirbase/demo/bundle.ndjson.gzip .
COPY --from=builder /fhirbase/fhirbase /usr/bin/fhirbase
COPY --from=builder /fhirbase/schema /fhirbase/schema
COPY --from=builder /fhirbase/transform /fhirbase/transform
COPY --from=builder /fhirbase/web /fhirbase/web

RUN chmod +x /usr/bin/fhirbase

RUN mkdir /pgdata && chown postgres:postgres /pgdata

USER postgres

RUN PGDATA=/pgdata /docker-entrypoint.sh postgres  & \
    until psql -U postgres -c '\q'; do \
        >&2 echo "Postgres is starting up..."; \
        sleep 5; \
    done && \
    psql -U postgres -c 'create database fhirbase;' && \
    fhirbase -d fhirbase init && \
    fhirbase -d fhirbase load --mode=insert ./bundle.ndjson.gzip \
    pg_ctl -D /pgdata stop

EXPOSE 3000

CMD pg_ctl -D /pgdata start && until psql -U postgres -c '\q'; do \
        >&2 echo "Postgres is starting up..."; \
        sleep 5; \
    done && \
    exec fhirbase -d fhirbase web

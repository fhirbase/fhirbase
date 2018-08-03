import psycopg2
from subprocess import Popen, PIPE
from threading import Thread


def db_create_import_table(cursor):
    cursor.execute(
            """
            create temporary table import
            (id varchar(250), txid int, resource_type varchar(250), resource text);

            CREATE OR REPLACE FUNCTION pg_temp.transform_resource()
            RETURNS trigger
            LANGUAGE plpgsql
            AS $BODY$
            BEGIN
                execute format('insert into %s (id, txid, ts, resource_type, status, resource)'
                               ' values($1, $2, now(), $3, cast ($4 as resource_status), cast ($5 as jsonb))'
                               ' on conflict do nothing', lower(NEW.resource_type))
                using NEW.id, NEW.txid, NEW.resource_type, 'created', NEW.resource;

                RETURN NULL;
            END
            $BODY$;

            CREATE TRIGGER import_trigger
            BEFORE INSERT
            ON import
            FOR EACH ROW
            EXECUTE PROCEDURE pg_temp.transform_resource();
            """)


def db_create_transaction_id(cursor):
    cursor.execute("insert into transaction(resource) values('{}') returning id")
    return str(cursor.fetchone()[0])


def db_copy_from(cursor, convout):
    cursor.copy_expert("copy import (id, txid, resource_type, resource) " \
            "from stdin csv quote e\'\\x01\' delimiter e\'\\x02\'", convout)


def import_lines(dbname, fhir_version, iterator):
    dbconn = psycopg2.connect(dbname=dbname, user="fhirbase",
            password="fhirbase", host="localhost", port="2345")

    try:
        with dbconn:
            cursor = dbconn.cursor()
            db_create_import_table(cursor)
            txid = db_create_transaction_id(cursor)
            conv_cmd = ['/usr/bin/env', 'java', '-jar', 'aidboxconv.jar',
                        txid, 'fhir-{0}'.format(fhir_version)]
            total = 0
            with Popen(conv_cmd, stdin=PIPE, stdout=PIPE) as convproc:
                copythread = Thread(target=db_copy_from, args=(cursor, convproc.stdout))
                copythread.start()
                for line in iterator:
                    convproc.stdin.write(line.encode())
                    convproc.stdin.write(b'\n')
                    total += 1
                convproc.stdin.close()
                copythread.join()
    finally:
        dbconn.close()

    print('{0} entries imported'.format(total))

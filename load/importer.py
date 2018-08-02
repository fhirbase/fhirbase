from subprocess import Popen, PIPE


def import_lines(dbname, fhir_version, iterator):
    def gen_psql_cmd(cmd):
        return ['/usr/bin/env', 'psql', '-d', dbname,
                '-c', cmd]

    create_transaction_cmd = gen_psql_cmd(
        'insert into transaction(resource) values(\'{}\') returning id')

    with Popen(create_transaction_cmd, stdout=PIPE) as proc:
        # TODO: exec psql by pg module
        txid = proc.stdout.read().decode().split('\n')[2].strip()

    # TODO: create temp table import
    """
    create table import (id varchar(250), txid int, resource_type varchar(250), resource text);
    """
    # TODO: create trigger
    """
    CREATE OR REPLACE FUNCTION transform_resource()
    RETURNS trigger
    LANGUAGE plpgsql
    AS $BODY$
    BEGIN
        execute format('insert into %s (id, txid, ts, resource_type, status, resource) values($1, $2, now(), $3, cast ($4 as resource_status), cast ($5 as jsonb)) on conflict do nothing', lower(NEW.resource_type))
        using NEW.id, NEW.txid, NEW.resource_type, 'created', NEW.resource;
    
        RETURN NULL;
    END
    $BODY$;
    
    CREATE TRIGGER test_trigger
    BEFORE INSERT
    ON import
    FOR EACH ROW
    EXECUTE PROCEDURE transform_resource();
    """

    conv_cmd = ['/usr/bin/env', 'java', '-jar', 'aidboxconv.jar',
                txid, 'fhir-{0}'.format(fhir_version)]
    copy_cmd = gen_psql_cmd(
        'copy import (id, txid, resource_type, resource) '
        'from stdin csv quote e\'\\x01\' delimiter e\'\\x02\'')
    total = 0
    with Popen(conv_cmd, stdin=PIPE, stdout=PIPE) as convproc, \
            Popen(copy_cmd, stdin=convproc.stdout, stdout=PIPE):
        for line in iterator:
            convproc.stdin.write(line.encode())
            convproc.stdin.write(b'\n')
            total += 1
        convproc.stdin.close()

    print('{0} entries imported'.format(total))

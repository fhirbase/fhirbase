import gzip
import ujson as json
import os

import psycopg2


def db_connect(dbname):
    pguser = os.getenv("PGUSER", "fhirbase")
    pgpassword = os.getenv("PGPASSWORD", "fhirbase")
    pghost = os.getenv("PGHOST", "localhost")
    pgport = os.getenv("PGPORT", "2345")
    return psycopg2.connect(dbname=dbname, user=pguser,
            password=pgpassword, host=pghost, port=pgport)


class CopyStream(object):
    def __init__(self, stream):
        self.stream = stream

    def readline(self, size=None):
        line = self.stream.readline()
        if not line:
            return ''
        parsed_line = json.loads(line)

        prepared_line = chr(2).join([
            parsed_line['id'],
            '0',
            parsed_line['resourceType'],
            'created',
            json.dumps(parsed_line)]) 

        return prepared_line + '\n'

    read = readline


if __name__ == '__main__':
    dbname = 'postgres'
    dbconn = db_connect(dbname)

    try:
        cursor = dbconn.cursor()

        with gzip.open('./sample-data.gzip', 'r') as gzfile:
            copy_stream = CopyStream(gzfile)

            cursor.copy_expert(
                "copy patient (id, txid, resource_type, status, resource) "
                "from stdin csv quote e\'\\x01\' delimiter e\'\\x02\'",
                copy_stream)
        cursor.execute("select count(*) from patient;")    
        print(cursor.fetchall())
        dbconn.commit()
    finally:
        dbconn.close()

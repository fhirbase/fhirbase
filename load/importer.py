import os
from collections import defaultdict

import psycopg2
from subprocess import Popen, PIPE
from threading import Thread

import io

from struct import pack


class CopyBinaryStream(object):
    def __init__(self, objects_stream, col_len):
        # self.st = StringIO()
        self.co = 0
        self.objects_stream = objects_stream
        self.col_len = col_len
        self.state = 0

    def readline(self, size=None):
        cpy = io.BytesIO()

        if self.state == 0: # start signature
            cpy.write(pack('!11sii', b'PGCOPY\n\377\r\n\0', 0, 0))
            cpy.seek(0)
            t = cpy.getvalue()
            self.state = 1
            return t
        elif self.state == 2: 			# File trailer
            cpy.write(pack('!h', -1))
            cpy.seek(0)
            t = cpy.getvalue()
            self.state == 3

            return t
        elif self.state == 3:
            return ''

        try:
            fields_all, hs = self.objects_stream.next()
        except StopIteration:
            self.state = 2
            return ''

        cpy.write(pack('!h', self.col_len))

        for value in fields_all:
            if value is None:
                hh =  pack('!i0s', 0, "")
                cpy.write(hh)
            else:
                size = len(value)
                hh =  pack('!i{}s'.format(size), size, value)
                cpy.write(hh)

        # hstore
        hstore = io.BytesIO()
        hh = pack('!i', len(hs)) # overall
        hstore.write(hh)

        for kk, vv in hs.viewitems():
            k = kk
            hh = pack('!i{}s'.format(len(k)), len(k), k)
            hstore.write(hh)

            v = vv
            hh = pack('!i{}s'.format(len(v)), len(v), v)
            hstore.write(hh)

        hstore.seek(0)
        gg = hstore.getvalue()

        hh = pack('!i', len(gg)) # overall
        cpy.write(hh)
        cpy.write(gg)

        # Copy data to database
        cpy.seek(0)

        t = cpy.getvalue()
        return t

    read = readline


def db_create_transaction_id(cursor):
    cursor.execute("insert into transaction(resource) values('{}') returning id")
    return str(cursor.fetchone()[0])


def db_copy(resource_type, acc, cursor):
    buffer = io.BytesIO(b'\n'.join(acc))
    buffer = CopyBinaryStream(buffer, 6)
    cursor.copy_expert(
        "COPY {0} (id, txid, ts, resource_type, status, resource) FROM STDIN WITH BINARY".format(resource_type.decode().lower()), buffer)


def db_accumulate_and_copy(cursor, convout):
    acc = defaultdict(list)

    while True:
        line = convout.readline()
        if not line:
            break
        id, txid, resource_type, resource = line.rstrip(b'\n').split(b'\x02')
        status = b'created'
        ts = b'2018-08-13 13:13:04.543863+00'

        acc[resource_type].append(b'\x02'.join(
            [id, txid, ts, resource_type, status, resource]))

        if len(acc[resource_type]) == 1000:
            db_copy(resource_type, acc[resource_type], cursor)
            del acc[resource_type]

    for (resource_type, lines) in acc.items():
        db_copy(resource_type, lines, cursor)


def db_connect(dbname):
    pguser = os.getenv("PGUSER", "fhirbase")
    pgpassword = os.getenv("PGPASSWORD", "fhirbase")
    pghost = os.getenv("PGHOST", "localhost")
    pgport = os.getenv("PGPORT", "2345")
    return psycopg2.connect(dbname=dbname, user=pguser,
            password=pgpassword, host=pghost, port=pgport)


def import_lines(dbname, fhir_version, iterator):
    dbconn = db_connect(dbname)

    try:
        with dbconn:
            cursor = dbconn.cursor()
            txid = db_create_transaction_id(cursor)
            conv_cmd = ['/usr/bin/env', 'java', '-jar', 'aidboxconv.jar',
                        txid, 'fhir-{0}'.format(fhir_version)]
            total = 0
            with Popen(conv_cmd, stdin=PIPE, stdout=PIPE) as convproc:
                copythread = Thread(target=db_accumulate_and_copy, args=(cursor, convproc.stdout))
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

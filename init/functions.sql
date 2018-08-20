
-- default id function implementation
-- you can override it by replacing
CREATE OR REPLACE FUNCTION fhirbase_genid()
RETURNS text AS $$
  select gen_random_uuid()::text
$$ LANGUAGE sql;

CREATE TYPE _resource AS (
  id text,
  txid bigint,
  ts timestamptz,
  resource_type text,
  status resource_status,
  resource jsonb
);

CREATE OR REPLACE FUNCTION _fhirbase_to_resource(x _resource)
RETURNS jsonb AS $$
 select x.resource || jsonb_build_object(
  'resourceType', x.resource_type,
  'id', x.id,
  'meta', coalesce(x.resource->'meta', '{}'::jsonb) || jsonb_build_object(
    'lastUpdated', x.ts,
    'versionId', x.txid::text
  )
 );
$$ LANGUAGE sql;

-- raise notice 'ID: %', rid;

CREATE OR REPLACE FUNCTION fhirbase_create(resource jsonb, txid bigint)
RETURNS jsonb AS $FUNCTION$
DECLARE
  _sql text;
  rt text;
  rid text;
  result jsonb;
BEGIN
    rt   := resource->>'resourceType';
    rid  := coalesce(resource->>'id', fhirbase_genid());
    _sql := format($SQL$
      WITH archived AS (
        INSERT INTO %s (id, txid, ts, status, resource)
        SELECT id, txid, ts, status, resource
        FROM %s
        WHERE id = $2
        RETURNING *
      ), inserted AS (
         INSERT INTO %s (id, ts, txid, status, resource)
         VALUES ($2, current_timestamp, $1, 'created', $3)
         ON CONFLICT (id)
         DO UPDATE SET
          txid = $1,
          ts = current_timestamp,
          status = 'recreated',
          resource = $3
         RETURNING *
      )

      select _fhirbase_to_resource(i.*) from inserted i

      $SQL$,
      rt || '_history', rt, rt, rt);

  EXECUTE _sql
  USING txid, rid, (resource - 'id')
  INTO result;

  return result;

END
$FUNCTION$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fhirbase_create(resource jsonb)
RETURNS jsonb AS $FUNCTION$
   SELECT fhirbase_create(resource, nextval('transaction_id_seq'));
$FUNCTION$ LANGUAGE sql;

CREATE OR REPLACE FUNCTION fhirbase_update(resource jsonb, txid bigint)
RETURNS jsonb AS $FUNCTION$
DECLARE
  _sql text ;
  rt text;
  rid text;
  result jsonb;
BEGIN
    rt   := resource->>'resourceType';
    rid  := resource->>'id';

    CASE WHEN (rid IS NULL) THEN
      RAISE EXCEPTION 'Resource does not have and id' USING HINT = 'Resource does not have and id';
    ELSE
    END CASE;

    _sql := format($SQL$
      WITH archived AS (
        INSERT INTO %s (id, txid, ts, status, resource)
        SELECT id, txid, ts, status, resource
        FROM %s
        WHERE id = $2
        RETURNING *
      ), inserted AS (
         INSERT INTO %s (id, ts, txid, status, resource)
         VALUES ($2, current_timestamp, $1, 'created', $3)
         ON CONFLICT (id)
         DO UPDATE SET
          txid = $1,
          ts = current_timestamp,
          status = 'updated',
          resource = $3
         RETURNING *
      )

      select _fhirbase_to_resource(i.*) from inserted i

      $SQL$,
      rt || '_history', rt, rt, rt);

  EXECUTE _sql
  USING txid, rid, (resource - 'id')
  INTO result;

  return result;

END
$FUNCTION$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fhirbase_update(resource jsonb)
RETURNS jsonb AS $FUNCTION$
   SELECT fhirbase_update(resource, nextval('transaction_id_seq'));
$FUNCTION$ LANGUAGE sql;

CREATE OR REPLACE FUNCTION fhirbase_read(resource_type text, id text)
RETURNS jsonb AS $FUNCTION$
DECLARE
  _sql text;
  result jsonb;
BEGIN
  _sql := format($SQL$
    SELECT _fhirbase_to_resource(row(r.*)::_resource) FROM %s r WHERE r.id = $1
  $SQL$,
  resource_type
  );

  EXECUTE _sql USING id INTO result;

  return result;
END
$FUNCTION$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fhirbase_delete(resource_type text, id text, txid bigint)
RETURNS jsonb AS $FUNCTION$
DECLARE
  _sql text;
  rt text;
  rid text;
  result jsonb;
BEGIN
    rt   := resource_type;
    rid  := id;
    _sql := format($SQL$
      WITH archived AS (
        INSERT INTO %s (id, txid, ts, status, resource)
        SELECT id, txid, ts, status, resource
        FROM %s WHERE id = $2
        RETURNING *
      ), deleted AS (
         INSERT INTO %s (id, txid, ts, status, resource)
         SELECT id, $1, current_timestamp, status, resource
         FROM %s WHERE id = $2
         RETURNING *
      ), dropped AS (
         DELETE FROM %s WHERE id = $2 RETURNING *
      )
      select _fhirbase_to_resource(i.*) from archived i

      $SQL$,
      rt || '_history', rt, rt || '_history', rt, rt);

  EXECUTE _sql
  USING txid, rid
  INTO result;

  return result;

END
$FUNCTION$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fhirbase_delete(resource_type text, id text)
RETURNS jsonb AS $FUNCTION$
   SELECT fhirbase_delete(resource_type, id, nextval('transaction_id_seq'));
$FUNCTION$ LANGUAGE sql;



-- SELECT fhirbase_create('{"resourceType": "Patient", "id": "ivan", "name": [{"family": "Ivanov"}]}'::jsonb);
-- SELECT fhirbase_create('{"resourceType": "Patient", "name": [{"family": "Ivanov"}]}'::jsonb);

-- truncate Patient;
-- truncate Patient_history;
-- SELECT fhirbase_create('{"resourceType": "Patient", "id": "nicola"}'::jsonb);
-- SELECT fhirbase_update('{"resourceType": "Patient", "id": "nicola", "name": [{"family": "Ryzhikov"}]}'::jsonb);
-- SELECT fhirbase_read('Patient', 'nicola');
-- SELECT fhirbase_delete('Patient', 'nicola');
-- SELECT fhirbase_read('Patient', 'nicola');
-- SELECT fhirbase_create('{"resourceType": "Patient", "id": "nicola"}'::jsonb);

-- SELECT _fhirbase_to_resource(row(r.*)::_resource) FROM Patient r;

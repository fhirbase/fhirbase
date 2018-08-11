
-- default id function implementation
-- you can override it by replacing
CREATE OR REPLACE FUNCTION fhirbase_genid()
RETURNS text AS $$
  select gen_random_uuid()::text
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
        SELECT id, $1, ts, status, resource
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

      select resource || jsonb_build_object(
        'resourceType', i.resource_type,
        'id', i.id,
        'meta', coalesce(resource->'meta', '{}'::jsonb) || jsonb_build_object(
           'lastUpdated', ts,
           'versionId', txid::text
        )
      )
      from inserted i

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

SELECT fhirbase_create('{"resourceType": "Patient", "id": "nicola"}'::jsonb);
-- SELECT fhirbase_create('{"resourceType": "Patient", "id": "ivan", "name": [{"family": "Ivanov"}]}'::jsonb);
SELECT fhirbase_create('{"resourceType": "Patient", "name": [{"family": "Ivanov"}]}'::jsonb);

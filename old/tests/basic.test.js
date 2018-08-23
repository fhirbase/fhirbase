const pg = require('./connection');


// SELECT fhirbase_update('{"resourceType": "Patient", "id": "nicola", "name": [{"family": "Ryzhikov"}]}'::jsonb);
// SELECT fhirbase_read('Patient', 'nicola');
// SELECT fhirbase_delete('Patient', 'nicola');
// SELECT fhirbase_read('Patient', 'nicola');
// SELECT fhirbase_create('{"resourceType": "Patient", "id": "nicola"}'::jsonb);
// SELECT _fhirbase_to_resource(row(r.*)::_resource) FROM Patient r;

it('basic', async ()=> {
  await pg.conn.query('truncate patient');
  await pg.conn.query('truncate patient_history');

  var res  = await pg.conn.query(
   'SELECT fhirbase_create(\'{"resourceType": "Patient", "id": "nicola"}\'::jsonb);'
  );

  console.log(res);

  expect(res.rows[0].id).toEqual("nicola");

});

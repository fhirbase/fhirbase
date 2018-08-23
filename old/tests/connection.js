const { Client } = require('pg');
const client = new Client();

async function init () {

  await client.connect();

  const res = await client.query('SELECT * from Patient');

  await client.end();

};

async function query(){
  return client.query.apply(client, arguments);
}

init();

module.exports = {
  conn: client,
  q: query
};

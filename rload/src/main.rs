// Updated example from http://rosettacode.org/wiki/Hello_world/Web_server#Rust
// to work with Rust 1.0 beta


extern crate postgres;
extern crate postgres_binary_copy;
extern crate curl;
extern crate flate2;
extern crate json;

use postgres::{Connection, TlsMode};
use std::env;
use curl::http;

use std::fs::File;
use std::io::prelude::*;
use std::io::{self, BufReader};
use std::path::Path;
use flate2::read::GzDecoder;
use flate2::read::MultiGzDecoder;
use postgres::types::{Type, ToSql};
use postgres_binary_copy::BinaryCopyReader;

fn var(key: &str) -> String {
    match env::var(key) {
        Ok(val) => return val,
        Err(_) => return "ups".to_string(),
    }
}

fn conn_str() -> String {
   return format!("postgres://{}:{}@{}:{}/{}", var("PGUSER"), var("PGPASSWORD"), var("PGHOST"), var("PGPORT"), var("PGDATABASE"));
}


fn test() {

    // let url = "https://aidbox.app/User?_format=yaml&__secret=jobanarot";
    // let resp = http::handle()
    //     .get(url)
    //     .exec()
    //     .unwrap_or_else(|e| {
    //         panic!("Failed to get {}; error is {}", url, e);
    //     });

    // if resp.get_code() != 200 {
    //     println!("Unable to handle HTTP response code {}", resp.get_code());
    //     return;
    // }

    // let body = std::str::from_utf8(resp.get_body()).unwrap_or_else(|e| {
    //     panic!("Failed to parse response from {}; error is {}", url, e);
    // });

    // println!("{}",body);

    // let conn = Connection::connect(conn_str(), TlsMode::None).unwrap();
    // let res = &conn.query("select 'hi'", &[]).unwrap();
    // let row = res.get(0);
    // let json: String = row.get(0);

    // println!("RES: {}", json);

    let f = File::open("tmp/out.gzip").unwrap();
    let reader = BufReader::new(f);
    let gzip = GzDecoder::new(reader);
    let greader = BufReader::new(gzip);
    let mut stream = greader.lines();


    let types = &[postgres::types::Type::Varchar];
    // let res:String = stream.next().unwrap().ok().unwrap();
    // for res in stream {
    //     let jsonstr =  res.ok().unwrap();
    //     let res = json::parse(&jsonstr).unwrap();
    //     println!("RES: {} / {}", res["resourceType"], res["id"]);

    // }

}

fn main() {
    test()
}

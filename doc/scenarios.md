Roles
=====

* Data Storage - software or server which has API to read/write FHIR
  resources (Data Storage don't have to conform FHIR REST API);
* Data Importer - software that interacts with Data Storage API to
  perform fast imports of large amounts of FHIR resources;
* User - a person who interacts with Data Importer program.

Scenario 1: Import FHIR resources from local file
=================================================

1. User prepares or somewhere downloads a [NDJSON](http://ndjson.org/)
   file containing FHIR resources. That file can be optionally
   GZIPed. File can contain resources of different kinds (like FHIR bundle).

2. User runs Data Importer to upload FHIR resources from that file
   into Data Storage. Optionally, Data Importer or Data Storage can
   perform validations to check resource content for FHIR conformance.

3. User checks that FHIR data was successfuly imported with Data
   Storage API. For instance, in PostgreSQL one can invoke:

   > SELECT COUNT(*) FROM patient;

   To check how much Patient resources was imported.

Scenario 2: Bulk Data API client
=================================================

1. User runs Data Importer providing [Bulk Data
   API](https://github.com/smart-on-fhir/fhir-bulk-data-docs) endpoint
   as an argument.

2. Data Importer acts as Bulk Data API client and downloads data
   returned by server.

3. Downloaded data is being imported to Data Storage by Data Importer.

4. User checks that FHIR data was successfuly imported. For instance,
   in PostgreSQL user can invoke:

   > SELECT COUNT(*) FROM patient;

   to check how much Patient resources was imported.


CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'resource_status') THEN
       CREATE TYPE resource_status AS ENUM ('created', 'updated', 'deleted', 'recreated');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS transaction (
  id serial primary key,
  ts timestamptz DEFAULT current_timestamp,
  resource jsonb);


CREATE TABLE IF NOT EXISTS "procedurerequest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ProcedureRequest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "procedurerequest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ProcedureRequest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "devicecomponent" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DeviceComponent',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "devicecomponent_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DeviceComponent',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "devicemetric" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DeviceMetric',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "devicemetric_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DeviceMetric',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "careplan" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'CarePlan',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "careplan_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'CarePlan',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "observation" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Observation',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "observation_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Observation',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "enrollmentrequest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EnrollmentRequest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "enrollmentrequest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EnrollmentRequest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "group" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Group',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "group_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Group',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "referralrequest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ReferralRequest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "referralrequest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ReferralRequest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "appointment" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Appointment',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "appointment_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Appointment',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "diagnosticorder" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DiagnosticOrder',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "diagnosticorder_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DiagnosticOrder',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "questionnaireresponse" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'QuestionnaireResponse',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "questionnaireresponse_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'QuestionnaireResponse',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "episodeofcare" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EpisodeOfCare',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "episodeofcare_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EpisodeOfCare',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "processresponse" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ProcessResponse',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "processresponse_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ProcessResponse',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "supplydelivery" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'SupplyDelivery',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "supplydelivery_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'SupplyDelivery',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "guidancerequest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'GuidanceRequest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "guidancerequest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'GuidanceRequest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "orderresponse" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'OrderResponse',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "orderresponse_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'OrderResponse',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "detectedissue" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DetectedIssue',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "detectedissue_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DetectedIssue',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "medicationadministration" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MedicationAdministration',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "medicationadministration_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MedicationAdministration',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "implementationguide" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ImplementationGuide',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "implementationguide_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ImplementationGuide',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "goal" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Goal',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "goal_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Goal',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "communication" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Communication',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "communication_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Communication',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "schedule" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Schedule',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "schedule_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Schedule',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "documentreference" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DocumentReference',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "documentreference_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DocumentReference',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "coverage" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Coverage',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "coverage_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Coverage',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "auditevent" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'AuditEvent',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "auditevent_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'AuditEvent',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "messageheader" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MessageHeader',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "messageheader_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MessageHeader',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "contract" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Contract',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "contract_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Contract',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "dataelement" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DataElement',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "dataelement_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DataElement',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "claimresponse" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ClaimResponse',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "claimresponse_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ClaimResponse',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "parameters" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Parameters',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "parameters_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Parameters',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "medicationorder" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MedicationOrder',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "medicationorder_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MedicationOrder',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "clinicalimpression" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ClinicalImpression',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "clinicalimpression_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ClinicalImpression',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "familymemberhistory" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'FamilyMemberHistory',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "familymemberhistory_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'FamilyMemberHistory',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "conformance" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Conformance',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "conformance_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Conformance',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "binary" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Binary',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "binary_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Binary',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "composition" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Composition',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "composition_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Composition',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "moduledefinition" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ModuleDefinition',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "moduledefinition_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ModuleDefinition',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "healthcareservice" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'HealthcareService',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "healthcareservice_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'HealthcareService',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "patient" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Patient',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "patient_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Patient',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "medicationdispense" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MedicationDispense',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "medicationdispense_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MedicationDispense',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "decisionsupportrule" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DecisionSupportRule',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "decisionsupportrule_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DecisionSupportRule',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "deviceusestatement" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DeviceUseStatement',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "deviceusestatement_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DeviceUseStatement',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "library" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Library',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "library_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Library',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "basic" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Basic',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "basic_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Basic',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "slot" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Slot',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "slot_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Slot',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "specimen" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Specimen',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "specimen_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Specimen',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "diagnosticreport" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DiagnosticReport',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "diagnosticreport_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DiagnosticReport',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "subscription" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Subscription',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "subscription_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Subscription',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "provenance" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Provenance',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "provenance_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Provenance',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "imagingobjectselection" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ImagingObjectSelection',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "imagingobjectselection_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ImagingObjectSelection',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "practitioner" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Practitioner',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "practitioner_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Practitioner',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "flag" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Flag',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "flag_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Flag',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "explanationofbenefit" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ExplanationOfBenefit',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "explanationofbenefit_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ExplanationOfBenefit',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "operationoutcome" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'OperationOutcome',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "operationoutcome_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'OperationOutcome',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "immunization" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Immunization',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "immunization_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Immunization',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "expansionprofile" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ExpansionProfile',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "expansionprofile_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ExpansionProfile',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "eligibilityrequest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EligibilityRequest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "eligibilityrequest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EligibilityRequest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "paymentnotice" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'PaymentNotice',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "paymentnotice_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'PaymentNotice',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "namingsystem" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'NamingSystem',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "namingsystem_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'NamingSystem',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "medicationstatement" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MedicationStatement',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "medicationstatement_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'MedicationStatement',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "enrollmentresponse" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EnrollmentResponse',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "enrollmentresponse_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EnrollmentResponse',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "nutritionorder" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'NutritionOrder',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "nutritionorder_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'NutritionOrder',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "modulemetadata" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ModuleMetadata',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "modulemetadata_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ModuleMetadata',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "questionnaire" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Questionnaire',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "questionnaire_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Questionnaire',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "account" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Account',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "account_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Account',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "communicationrequest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'CommunicationRequest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "communicationrequest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'CommunicationRequest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "documentmanifest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DocumentManifest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "documentmanifest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DocumentManifest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "eligibilityresponse" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EligibilityResponse',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "eligibilityresponse_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'EligibilityResponse',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "valueset" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ValueSet',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "valueset_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ValueSet',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "claim" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Claim',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "claim_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Claim',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "deviceuserequest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DeviceUseRequest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "deviceuserequest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DeviceUseRequest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "measure" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Measure',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "measure_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Measure',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "list" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'List',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "list_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'List',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "encounter" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Encounter',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "encounter_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Encounter',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "visionprescription" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'VisionPrescription',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "visionprescription_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'VisionPrescription',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "riskassessment" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'RiskAssessment',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "riskassessment_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'RiskAssessment',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "bodysite" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'BodySite',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "bodysite_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'BodySite',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "immunizationrecommendation" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ImmunizationRecommendation',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "immunizationrecommendation_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ImmunizationRecommendation',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "processrequest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ProcessRequest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "processrequest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ProcessRequest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "orderset" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'OrderSet',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "orderset_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'OrderSet',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "relatedperson" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'RelatedPerson',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "relatedperson_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'RelatedPerson',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "medication" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Medication',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "medication_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Medication',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "appointmentresponse" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'AppointmentResponse',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "appointmentresponse_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'AppointmentResponse',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "substance" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Substance',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "substance_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Substance',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "paymentreconciliation" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'PaymentReconciliation',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "paymentreconciliation_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'PaymentReconciliation',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "testscript" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'TestScript',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "testscript_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'TestScript',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "conceptmap" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ConceptMap',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "conceptmap_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ConceptMap',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "person" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Person',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "person_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Person',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "condition" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Condition',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "condition_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Condition',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "decisionsupportservicemodule" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DecisionSupportServiceModule',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "decisionsupportservicemodule_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'DecisionSupportServiceModule',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "structuredefinition" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'StructureDefinition',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "structuredefinition_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'StructureDefinition',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "procedure" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Procedure',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "procedure_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Procedure',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "location" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Location',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "location_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Location',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "organization" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Organization',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "organization_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Organization',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "device" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Device',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "device_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Device',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "supplyrequest" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'SupplyRequest',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "supplyrequest_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'SupplyRequest',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "allergyintolerance" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'AllergyIntolerance',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "allergyintolerance_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'AllergyIntolerance',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "operationdefinition" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'OperationDefinition',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "operationdefinition_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'OperationDefinition',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "imagingstudy" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ImagingStudy',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "imagingstudy_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'ImagingStudy',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "guidanceresponse" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'GuidanceResponse',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "guidanceresponse_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'GuidanceResponse',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "media" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Media',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "media_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Media',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);


CREATE TABLE IF NOT EXISTS "order" (
  id text primary key,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Order',
  status resource_status not null,
  resource jsonb not null
);

CREATE TABLE IF NOT EXISTS "order_history" (
  id text,
  txid bigint not null,
  ts timestamptz DEFAULT current_timestamp,
  resource_type text default 'Order',
  status resource_status not null,
  resource jsonb not null,
	PRIMARY KEY (id, txid)
);

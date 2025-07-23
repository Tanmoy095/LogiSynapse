-- the goose command from entrypoint.sh will run these migrations when the shipment service container startts 
-- ensuring the db has required schema
-- migrations applied using goose
--pgcrypto extension enables uuid generations for unique shipment iDS

-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DROP EXTENSION IF EXISTS "pgcrypto";

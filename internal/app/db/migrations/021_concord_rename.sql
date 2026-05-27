-- Migration 021: Rename synastry → concord in bond_json.
-- The Bond.Concord field's JSON tag changed from "synastry" to "concord".
-- Existing bond_json rows have the old key and need to be updated.

UPDATE bond_events SET bond_json = REPLACE(bond_json, '"synastry":', '"concord":');

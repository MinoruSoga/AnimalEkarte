ALTER TABLE pets
    ADD CONSTRAINT uq_pets_clinic_id_id
    UNIQUE (clinic_id, id);

ALTER TABLE owners
    ADD CONSTRAINT uq_owners_clinic_id_id
    UNIQUE (clinic_id, id);

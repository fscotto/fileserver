CREATE TABLE IF NOT EXISTS documents
(
    id          SERIAL PRIMARY KEY,
    name        TEXT                        NOT NULL,
    id_file     UUID UNIQUE                 NOT NULL,
    fingerprint TEXT UNIQUE                 NOT NULL,
    created_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMP WITHOUT TIME ZONE
);

-- Crea un indice su deleted_at per il supporto soft delete
CREATE INDEX IF NOT EXISTS idx_documents_deleted_at ON documents (deleted_at);

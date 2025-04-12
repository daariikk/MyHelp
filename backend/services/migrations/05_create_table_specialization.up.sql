-- Таблица specialization
CREATE TABLE specialization (
    id SERIAL PRIMARY KEY,
    specialization VARCHAR(100),
    specialization_doctor VARCHAR(100),
    description TEXT
);
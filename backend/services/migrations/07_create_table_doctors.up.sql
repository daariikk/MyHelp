-- Таблица doctors (обновленная с каскадным удалением)
CREATE TABLE doctors (
    id SERIAL PRIMARY KEY,
    surname VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    patronymic VARCHAR(100) NOT NULL,
    specialization_id INT NOT NULL,
    education VARCHAR(512),
    progress VARCHAR(1024),
    rating FLOAT default 5.0,
	photo_path VARCHAR(512),
    FOREIGN KEY (specialization_id) REFERENCES specialization(id) ON DELETE CASCADE
);
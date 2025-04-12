-- Таблица appointments (обновленная с каскадным удалением)
CREATE TABLE appointments (
    id SERIAL PRIMARY KEY,
    doctor_id INT,
    patient_id INT,
    date DATE,
    time TIME,
    status_id INT,
    rating FLOAT,
    FOREIGN KEY (doctor_id) REFERENCES doctors(id) ON DELETE CASCADE,
    FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE,
    FOREIGN KEY (status_id) REFERENCES status_appointment(id)
);
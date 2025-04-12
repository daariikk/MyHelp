-- Таблица doctor_schedules (обновленная с каскадным удалением)
CREATE TABLE doctor_schedules (
    id SERIAL PRIMARY KEY,
    doctor_id INT,
    date DATE,
    start_time TIME,
    end_time TIME,
    is_available BOOLEAN,
    FOREIGN KEY (doctor_id) REFERENCES doctors(id) ON DELETE CASCADE
);
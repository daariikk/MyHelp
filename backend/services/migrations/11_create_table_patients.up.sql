CREATE TABLE patients (
	id SERIAL PRIMARY KEY,
	surname VARCHAR(100),
	name VARCHAR(100),
	patronymic  VARCHAR(100),
	email VARCHAR(256) NOT NULL,
	polic VARCHAR(100) NOT NULL,
	password VARCHAR(1024) NOT NULL,
	is_deleted BOOLEAN default false
);
CREATE TABLE IF NOT EXISTS resource (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(30) NOT NULL,
    content TEXT NOT NULL,
    url TEXT,
    images TEXT[],
    last_edition_at TIMESTAMP NOT NULL,
	last_edition_by VARCHAR(50) NOT NULL,
    section_id BIGINT NOT NULL,
    locked_by BIGINT,
    lock_status BOOLEAN DEFAULT FALSE,
    CONSTRAINT section_resource_id_fk FOREIGN KEY (section_id) REFERENCES section(id) ON DELETE CASCADE
);

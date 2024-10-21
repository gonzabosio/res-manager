CREATE TABLE IF NOT EXISTS resource (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(30) NOT NULL,
    content TEXT NOT NULL,
    url TEXT,
    section_id BIGINT NOT NULL,
    CONSTRAINT section_resource_id_fk FOREIGN KEY (section_id) REFERENCES section(id) ON DELETE CASCADE
);

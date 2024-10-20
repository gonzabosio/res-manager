CREATE TABLE IF NOT EXISTS "section" (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(30) NOT NULL,
    project_id BIGINT NOT NULL,
    CONSTRAINT section_project_id_fk FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE
);
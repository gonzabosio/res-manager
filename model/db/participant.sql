CREATE TABLE IF NOT EXISTS participant(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    team_id BIGINT NOT NULL,
    CONSTRAINT participant_user_id_fk FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    CONSTRAINT participant_team_id_fk FOREIGN KEY (team_id) REFERENCES team(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS index_user_id ON participant(user_id);
CREATE INDEX IF NOT EXISTS index_team_id ON participant(team_id);


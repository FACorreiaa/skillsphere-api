CREATE TYPE skill_type AS ENUM ('offered', 'wanted');

CREATE TABLE user_skills (
                           user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                           skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
                           skill_type skill_type NOT NULL,

  -- Proficiency on a scale of 1-10.
                           proficiency SMALLINT CHECK (proficiency >= 1 AND proficiency <= 10),

                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                           PRIMARY KEY (user_id, skill_id, skill_type)
);

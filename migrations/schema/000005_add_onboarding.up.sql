CREATE TABLE onboarding_steps (
  id INT PRIMARY KEY
);

CREATE TABLE onboarding_options (
  id SERIAL PRIMARY KEY,
  step_id INT NOT NULL REFERENCES onboarding_steps(id) ON DELETE CASCADE,
  option_index INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_onboarding_options_step
  ON onboarding_options(step_id);


INSERT INTO onboarding_steps (id) VALUES
  (1),
  (2),
  (3),
  (4),
  (5);

INSERT INTO onboarding_options (step_id, option_index) VALUES
  (2, 1),
  (2, 2),
  (2, 3),
  (2, 4),
  (2, 5),
  (2, 6),
  (2, 7),
  (2, 8),
  (2, 9);

INSERT INTO onboarding_options (step_id, option_index) VALUES
  (3, 1),
  (3, 2),
  (3, 3),
  (3, 4);

INSERT INTO onboarding_options (step_id, option_index) VALUES
  (4, 1),
  (4, 2),
  (4, 3),
  (4, 4),
  (4, 5),
  (4, 6);

CREATE TABLE salon_onboarding_selection (
  salon_id UUID NOT NULL, 
  step_id INT NOT NULL REFERENCES onboarding_steps(id),
  option_index INT NOT NULL,
  chosen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (salon_id, step_id, option_index)
);
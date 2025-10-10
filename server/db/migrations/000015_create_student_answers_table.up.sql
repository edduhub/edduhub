CREATE TABLE IF NOT EXISTS student_answers (
    id SERIAL PRIMARY KEY,
    attempt_id INTEGER NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    selected_option_id INTEGER REFERENCES answer_options(id) ON DELETE SET NULL,
    answer_text TEXT,
    is_correct BOOLEAN,
    marks_awarded INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(attempt_id, question_id)
);

CREATE INDEX idx_student_answers_attempt ON student_answers(attempt_id);
CREATE INDEX idx_student_answers_question ON student_answers(question_id);

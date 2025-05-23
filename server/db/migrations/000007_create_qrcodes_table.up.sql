BEGIN;

CREATE TABLE IF NOT EXISTS qrcodes (
    -- Added an ID PK for consistency, assuming qr_code_id might not be unique alone over time
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    student_id INT NOT NULL,
    qr_code_id VARCHAR(255) NOT NULL UNIQUE, -- Assuming the QR code string itself should be unique
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_qrcodes_student
        FOREIGN KEY (student_id)
        REFERENCES students(student_id)
        ON DELETE CASCADE -- If student is deleted, remove their QR codes
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_qrcodes_student_id ON qrcodes (student_id);
CREATE INDEX IF NOT EXISTS idx_qrcodes_qr_code_id ON qrcodes (qr_code_id); -- For unique constraint/lookups
CREATE INDEX IF NOT EXISTS idx_qrcodes_expires_at ON qrcodes (expires_at);

COMMIT;

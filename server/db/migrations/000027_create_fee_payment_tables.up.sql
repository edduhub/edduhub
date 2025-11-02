BEGIN;

-- Create fee_structures table
CREATE TABLE IF NOT EXISTS fee_structures (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    college_id INT NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    fee_type VARCHAR(50) NOT NULL, -- 'tuition', 'hostel', 'exam', 'library', 'misc'
    frequency VARCHAR(50) NOT NULL DEFAULT 'semester', -- 'semester', 'annual', 'monthly', 'one-time'
    academic_year VARCHAR(20),
    semester VARCHAR(20),
    department_id INT,
    course_id INT,
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    due_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_fee_structures_college FOREIGN KEY (college_id) REFERENCES colleges(id) ON DELETE CASCADE,
    CONSTRAINT fk_fee_structures_department FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL,
    CONSTRAINT fk_fee_structures_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE SET NULL
);

-- Create fee_assignments table (which students owe which fees)
CREATE TABLE IF NOT EXISTS fee_assignments (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    student_id INT NOT NULL,
    fee_structure_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL, -- Can be different from structure amount (scholarships, waivers)
    waiver_amount DECIMAL(10, 2) DEFAULT 0,
    waiver_reason TEXT,
    due_date TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'partial', 'paid', 'overdue', 'waived'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_fee_assignments_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    CONSTRAINT fk_fee_assignments_fee_structure FOREIGN KEY (fee_structure_id) REFERENCES fee_structures(id) ON DELETE CASCADE,
    CONSTRAINT unique_student_fee_structure UNIQUE(student_id, fee_structure_id)
);

-- Create fee_payments table
CREATE TABLE IF NOT EXISTS fee_payments (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    fee_assignment_id INT NOT NULL,
    student_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50) NOT NULL, -- 'card', 'bank_transfer', 'cash', 'cheque', 'online'
    payment_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed', 'refunded'
    transaction_id VARCHAR(255), -- External payment gateway transaction ID
    gateway VARCHAR(50), -- 'stripe', 'paypal', 'razorpay', etc.
    gateway_response JSONB, -- Store full gateway response
    receipt_number VARCHAR(100) UNIQUE,
    payment_date TIMESTAMPTZ,
    notes TEXT,
    processed_by INT, -- User ID who processed the payment
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_fee_payments_assignment FOREIGN KEY (fee_assignment_id) REFERENCES fee_assignments(id) ON DELETE CASCADE,
    CONSTRAINT fk_fee_payments_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    CONSTRAINT fk_fee_payments_processed_by FOREIGN KEY (processed_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Create fee_payment_reminders table
CREATE TABLE IF NOT EXISTS fee_payment_reminders (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    fee_assignment_id INT NOT NULL,
    student_id INT NOT NULL,
    reminder_date TIMESTAMPTZ NOT NULL,
    reminder_type VARCHAR(50) NOT NULL, -- 'email', 'sms', 'notification'
    message TEXT,
    sent BOOLEAN NOT NULL DEFAULT FALSE,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_reminders_assignment FOREIGN KEY (fee_assignment_id) REFERENCES fee_assignments(id) ON DELETE CASCADE,
    CONSTRAINT fk_reminders_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX idx_fee_structures_college_id ON fee_structures(college_id);
CREATE INDEX idx_fee_structures_department_id ON fee_structures(department_id);
CREATE INDEX idx_fee_structures_course_id ON fee_structures(course_id);
CREATE INDEX idx_fee_structures_due_date ON fee_structures(due_date);
CREATE INDEX idx_fee_assignments_student_id ON fee_assignments(student_id);
CREATE INDEX idx_fee_assignments_fee_structure_id ON fee_assignments(fee_structure_id);
CREATE INDEX idx_fee_assignments_status ON fee_assignments(status);
CREATE INDEX idx_fee_assignments_due_date ON fee_assignments(due_date);
CREATE INDEX idx_fee_payments_student_id ON fee_payments(student_id);
CREATE INDEX idx_fee_payments_assignment_id ON fee_payments(fee_assignment_id);
CREATE INDEX idx_fee_payments_status ON fee_payments(payment_status);
CREATE INDEX idx_fee_payments_transaction_id ON fee_payments(transaction_id);
CREATE INDEX idx_fee_payments_receipt_number ON fee_payments(receipt_number);
CREATE INDEX idx_fee_reminders_student_id ON fee_payment_reminders(student_id);
CREATE INDEX idx_fee_reminders_sent ON fee_payment_reminders(sent);

COMMIT;

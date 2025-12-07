-- ============================================
-- Achievement Schema for PostgreSQL
-- ============================================

-- Table: achievement_references
-- Menyimpan reference achievement dari MongoDB ke PostgreSQL
CREATE TABLE IF NOT EXISTS achievement_references (
    id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    mongo_achievement_id VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'draft',
    submitted_at TIMESTAMP,
    verified_at TIMESTAMP,
    verified_by UUID,
    rejection_note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
);

-- Index untuk performa query
CREATE INDEX IF NOT EXISTS idx_achievement_student_id ON achievement_references(student_id);
CREATE INDEX IF NOT EXISTS idx_achievement_status ON achievement_references(status);
CREATE INDEX IF NOT EXISTS idx_achievement_mongo_id ON achievement_references(mongo_achievement_id);

-- ============================================
-- Permissions untuk Achievement
-- ============================================

-- Insert permissions untuk achievements
INSERT INTO permissions (id, name, resource, action, description) VALUES
('p16', 'achievements.create', 'achievements', 'create', 'Submit prestasi baru'),
('p17', 'achievements.read', 'achievements', 'read', 'Melihat prestasi'),
('p18', 'achievements.update', 'achievements', 'update', 'Update prestasi'),
('p19', 'achievements.delete', 'achievements', 'delete', 'Hapus prestasi')
ON CONFLICT (id) DO NOTHING;

-- Assign permissions ke roles
-- Student: bisa create, read, update, delete prestasi sendiri
INSERT INTO role_permissions (role_id, permission_id) VALUES
('3', 'p16'), ('3', 'p17'), ('3', 'p18'), ('3', 'p19')
ON CONFLICT DO NOTHING;

-- Lecturer: bisa read prestasi mahasiswa bimbingannya
INSERT INTO role_permissions (role_id, permission_id) VALUES
('2', 'p17')
ON CONFLICT DO NOTHING;

-- Admin: full access
INSERT INTO role_permissions (role_id, permission_id) VALUES
('1', 'p16'), ('1', 'p17'), ('1', 'p18'), ('1', 'p19')
ON CONFLICT DO NOTHING;

-- ============================================
-- Sample Data (Optional)
-- ============================================

-- Contoh data akan dibuat melalui API
-- ============================================
-- RBAC Database Schema
-- ============================================

-- Table: roles (jika belum ada)
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: permissions
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100),
    action VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: role_permissions (junction table)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id VARCHAR(36) REFERENCES roles(id) ON DELETE CASCADE,
    permission_id VARCHAR(36) REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

-- ============================================
-- Sample Data
-- ============================================

-- Insert sample roles
INSERT INTO roles (id, name, description) VALUES
('1', 'admin', 'Administrator dengan akses penuh'),
('2', 'lecturer', 'Dosen/Pembimbing'),
('3', 'student', 'Mahasiswa/Alumni')
ON CONFLICT (id) DO NOTHING;

-- Insert sample permissions
INSERT INTO permissions (id, name, resource, action, description) VALUES
-- User permissions
('p1', 'users.read', 'users', 'read', 'Melihat daftar user'),
('p2', 'users.create', 'users', 'create', 'Membuat user baru'),
('p3', 'users.update', 'users', 'update', 'Mengupdate data user'),
('p4', 'users.delete', 'users', 'delete', 'Menghapus user'),

-- Student permissions
('p5', 'students.read', 'students', 'read', 'Melihat daftar mahasiswa'),
('p6', 'students.create', 'students', 'create', 'Membuat data mahasiswa'),
('p7', 'students.update', 'students', 'update', 'Mengupdate data mahasiswa'),
('p8', 'students.delete', 'students', 'delete', 'Menghapus data mahasiswa'),

-- Lecturer permissions
('p9', 'lecturers.read', 'lecturers', 'read', 'Melihat daftar dosen'),
('p10', 'lecturers.create', 'lecturers', 'create', 'Membuat data dosen'),
('p11', 'lecturers.update', 'lecturers', 'update', 'Mengupdate data dosen'),
('p12', 'lecturers.delete', 'lecturers', 'delete', 'Menghapus data dosen'),

-- Admin permissions
('p13', 'admin.dashboard', 'admin', 'read', 'Akses dashboard admin'),
('p14', 'admin.settings', 'admin', 'write', 'Mengubah pengaturan sistem')
ON CONFLICT (id) DO NOTHING;

-- Assign permissions to roles
-- Admin: Full access
INSERT INTO role_permissions (role_id, permission_id) VALUES
('1', 'p1'), ('1', 'p2'), ('1', 'p3'), ('1', 'p4'),
('1', 'p5'), ('1', 'p6'), ('1', 'p7'), ('1', 'p8'),
('1', 'p9'), ('1', 'p10'), ('1', 'p11'), ('1', 'p12'),
('1', 'p13'), ('1', 'p14')
ON CONFLICT DO NOTHING;

-- Lecturer: Read students, manage own data
INSERT INTO role_permissions (role_id, permission_id) VALUES
('2', 'p5'), ('2', 'p9'), ('2', 'p11')
ON CONFLICT DO NOTHING;

-- Student: Read only
INSERT INTO role_permissions (role_id, permission_id) VALUES
('3', 'p5'), ('3', 'p9')
ON CONFLICT DO NOTHING;
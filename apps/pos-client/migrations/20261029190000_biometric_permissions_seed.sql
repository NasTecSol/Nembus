-- +goose Up
-- Seed biometrics.enroll permission and grant to supervisor roles
INSERT INTO permissions (name, code, description)
VALUES ('Enroll Biometrics', 'biometrics.enroll', 'Enroll or deactivate user fingerprint templates')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE p.code = 'biometrics.enroll'
  AND r.code IN ('admin', 'super_admin', 'supervisor', 'store_manager')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- +goose Up
-- Rename permission biometrics.enroll -> users:enroll (existing databases keep their role grants)
UPDATE permissions
SET code = 'users:enroll'
WHERE code = 'biometrics.enroll'
  AND NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'users:enroll');

SELECT setval(
  pg_get_serial_sequence('permissions', 'id'),
  COALESCE((SELECT MAX(id) FROM permissions), 1),
  (SELECT MAX(id) IS NOT NULL FROM permissions)
);

INSERT INTO permissions (name, code, description)
VALUES ('Enroll Biometrics', 'users:enroll', 'Enroll or deactivate user fingerprint templates')
ON CONFLICT (code) DO NOTHING;

SELECT setval(
  pg_get_serial_sequence('role_permissions', 'id'),
  COALESCE((SELECT MAX(id) FROM role_permissions), 1),
  (SELECT MAX(id) IS NOT NULL FROM role_permissions)
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE p.code = 'users:enroll'
  AND r.code IN ('admin', 'super_admin', 'supervisor', 'store_manager')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- name: GetCashierWithLimits :one
SELECT 
    c.id,
    c.cashier_code,
    c.drawer_limit,
    c.discount_limit,
    c.is_active,
    u.first_name,
    u.last_name,
    u.email,
    s.name AS store_name
FROM cashiers c
JOIN users u ON c.user_id = u.id
JOIN stores s ON c.store_id = s.id
WHERE c.id = $1;

-- name: ListActiveCashiersInStore :many
SELECT 
    c.id,
    c.cashier_code,
    u.first_name || ' ' || u.last_name AS full_name,
    c.discount_limit,
    c.drawer_limit,
    COUNT(cs.id) FILTER (WHERE cs.status = 'open') AS active_sessions
FROM cashiers c
JOIN users u ON c.user_id = u.id
LEFT JOIN cashier_sessions cs ON cs.cashier_id = c.id AND cs.status = 'open'
WHERE c.store_id = $1
  AND c.is_active = true
GROUP BY c.id, u.first_name, u.last_name
ORDER BY u.first_name;
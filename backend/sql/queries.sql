-- name: GetWindowGeometry :one
SELECT id, width, height, x, y
FROM window
ORDER BY id ASC
LIMIT 1;

-- name: UpdateWindowGeometry :exec
UPDATE window
SET
    width  = sqlc.arg('width'),
    height = sqlc.arg('height'),
    x      = sqlc.arg('x'),
    y      = sqlc.arg('y')
WHERE id = 1;

-- name: InsertDefaultWindow :exec
INSERT INTO window (
    id,
    width,
    height,
    x,
    y
)
VALUES (
    1,
    sqlc.arg('width'),
    sqlc.arg('height'),
    sqlc.arg('x'),
    sqlc.arg('y')
);

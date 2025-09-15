-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ResetUsers :exec
delete from users;

-- name: GetUserByEmail :one
select * from users
where email = $1;

-- name: UpdateLoginDetails :one
update users
set updated_at = NOW(), email = $1, hashed_password = $2
where id = $3
returning *;
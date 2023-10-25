-- name: CreateEntry :one
insert into entries (account_id, amount)
values ($1, $2)
returning *
;

-- name: GetEntry :one
select id, account_id, amount
from entries
where id = $1
limit 1
;

-- name: ListEntries :many
select id, account_id, amount
from entries
where account_id = $1
order by id
limit $2
offset $3
;

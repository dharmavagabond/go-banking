-- name: CreateTransfer :one
insert into transfers (from_account_id, to_account_id, amount)
values ($1, $2, $3)
returning *
;

-- name: GetTransfer :one
select id, from_account_id, to_account_id, amount
from transfers
where id = $1
limit 1
;

-- name: ListTransfers :many
select id, from_account_id, to_account_id, amount
from transfers
where from_account_id = $1 or to_account_id = $2
order by id
limit $3
offset $4
;

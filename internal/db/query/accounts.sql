-- name: CreateAccount :one
insert into accounts (owner, balance, currency)
values ($1, $2, $3)
returning *
;

-- name: GetAccount :one
select id, owner, balance, currency
from accounts
where id = $1
limit 1
;

-- name: GetAccountForUpdate :one
select id, owner, balance, currency
from accounts
where id = $1
limit 1
for no key update
;

-- name: ListAccounts :many
select id, owner, balance, currency
from accounts
where owner = $1
order by id
limit $2
offset $3
;

-- name: UpdateAccount :one
update accounts
set balance = $2
where id = $1
returning *
;

-- name: AddAccountBalance :one
update accounts
set balance = balance + @amount_to_transfer
where id = @id
returning *
;

-- name: DeleteAccount :exec
delete from accounts
where id = $1
;

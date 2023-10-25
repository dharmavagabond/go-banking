create table "users" (
    "username" varchar primary key,
    "hashed_password" varchar not null,
    "full_name" varchar not null,
    "email" varchar unique not null,
    "password_changed_at" timestamptz not null default '0001-01-01 00:00:00Z',
    "created_at" timestamptz not null default 'now()'
)
;

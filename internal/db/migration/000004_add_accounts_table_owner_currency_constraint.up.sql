alter table "accounts"
add constraint accounts_owner_currency_key unique ("owner", "currency")
;

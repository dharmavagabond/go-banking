ALTER TABLE "accounts"
    ADD CONSTRAINT accounts_owner_currency_key UNIQUE ("owner", "currency");


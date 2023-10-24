ALTER TABLE "accounts"
    ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");


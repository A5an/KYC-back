ALTER TABLE products_kyc
    ADD COLUMN passport_verification_link TEXT,
    ADD COLUMN image_url TEXT,
    ADD COLUMN passport_number VARCHAR(255),
    ADD COLUMN passport_verification_status VARCHAR(255);
CREATE TABLE IF NOT EXISTS products
(
    id          VARCHAR(255) NOT NULL,
    name        VARCHAR(255) UNIQUE,
    description TEXT,
    image_url   TEXT,
    provider_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS risk_parameters
(
    id                varchar(255) PRIMARY KEY,
    country           varchar(20)  NOT NULL UNIQUE,
    account_Balance   FLOAT        NOT NULL,
    average_salary    FLOAT        NOT NULL,
    employment_status BOOLEAN      NOT NULL,
    provider_id       VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS products_kyc
(
    id                         varchar(255) PRIMARY KEY,
    link                       TEXT,
    product_id                 VARCHAR(255) NOT NULL,
    provider_id                VARCHAR(255) NOT NULL,
    first_name                 varchar(255),
    middle_name                varchar(255),
    last_name                  varchar(255),
    dob                        DATE,
    country                    varchar(30),
    gender                    varchar(30),
    account_balance            FLOAT,
    employment_status          BOOLEAN,
    average_salary             FLOAT,
    bank_verification_number   VARCHAR(255),
    id_type                    VARCHAR(255),
    mobile_number              VARCHAR(255),
    status                     VARCHAR(255),
    account_balance_risk_level VARCHAR(255),
    average_salary_risk_level  VARCHAR(255),
    employment_risk_level      VARCHAR(255),
    identity_response          jsonb,
    FOREIGN KEY (product_id) REFERENCES products (id)
);
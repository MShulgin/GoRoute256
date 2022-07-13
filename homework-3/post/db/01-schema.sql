CREATE TABLE "income_delivery"
(
    post_id      bigint    NOT NULL,
    shipment_id  uuid      NOT NULL UNIQUE,
    created_time timestamp NOT NULL
);

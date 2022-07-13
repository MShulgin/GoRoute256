CREATE TABLE "offer"
(
    id         text PRIMARY KEY,
    seller_id  bigint           NOT NULL,
    product_id bigint           NOT NULL,
    stock      bigint           NOT NULL,
    reserved   bigint           NOT NULL,
    price      numeric(1000, 2) NOT NULL,
    CHECK ( stock >= 0 ),
    CHECK ( reserved >= 0 )
);

CREATE SEQUENCE offer_id START 1;

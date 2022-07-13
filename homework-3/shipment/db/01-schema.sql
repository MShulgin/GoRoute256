CREATE TABLE "shipment"
(
    id             uuid PRIMARY KEY,
    order_id       text      NOT NULL,
    seller_id      bigint    NOT NULL,
    units          jsonb     NOT NULL,
    destination_id bigint    NOT NULL,
    status         integer   NOT NULL,
    created_time   timestamp NOT NULL
);

CREATE INDEX shipment_id_idx ON shipment (id);
CREATE INDEX shipment_order_id_idx ON shipment (order_id);

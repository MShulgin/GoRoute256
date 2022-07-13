INSERT INTO "offer" VALUES (nextval('offer_id') || '-1' || '-1', 1, 1, 10, 2, 100.0);
INSERT INTO "offer" VALUES (nextval('offer_id') || '-1' || '-2', 1, 2, 5, 0, 500.0);

CREATE OR REPLACE FUNCTION random_between(low INT ,high INT)
    RETURNS INT AS
$$
BEGIN
    RETURN floor(random()* (high-low + 1) + low);
END;
$$ language 'plpgsql' STRICT;

INSERT INTO "offer"
SELECT
    nextval('offer_id') || '-' || random_between(1, 10000) || '-' || random_between(1, 10000),
    random_between(1, 10000),
    random_between(1, 10000),
    random_between(1, 5000),
    random_between(1, 5000),
    random_between(1, 1000000)
FROM generate_series(1, 1000000);

DROP FUNCTION random_between;
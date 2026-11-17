CREATE TABLE IF NOT EXISTS menu_items (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    price NUMERIC(6,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    dish TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'new',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO menu_items (name, price)
SELECT * FROM (VALUES
    ('Chicken Karahi', 12.00),
    ('Daal Chawal', 7.00),
    ('Seekh Kebab', 9.00)
) AS v(name, price)
WHERE NOT EXISTS (SELECT 1 FROM menu_items);

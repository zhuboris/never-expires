CREATE TEMP TABLE temp_food_names (
    en TEXT,
    ru TEXT
);

COPY temp_food_names(en, ru)
FROM '/assets/food_names.csv'
DELIMITER ','
CSV HEADER;

INSERT INTO shared_types_of_items (name)
SELECT en FROM temp_food_names
UNION ALL
SELECT ru FROM temp_food_names
ORDER BY 1
ON CONFLICT DO NOTHING;

DROP TABLE temp_food_names;
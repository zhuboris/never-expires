FROM postgres:15.3-alpine3.18
COPY assets/food_names.csv /assets/food_names.csv
COPY init.sql /docker-entrypoint-initdb.d/01_init.sql
COPY fill_food_names.sql /docker-entrypoint-initdb.d/02_fill_food_names.sql

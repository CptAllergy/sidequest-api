#!/bin/bash
set -e

if [ -n "$APP_DB_NAME" ]; then
    echo "Creating additional database: $APP_DB_NAME"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        CREATE DATABASE $APP_DB_NAME OWNER $POSTGRES_USER;
EOSQL
fi
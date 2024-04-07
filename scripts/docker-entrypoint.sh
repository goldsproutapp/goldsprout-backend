#!/usr/bin/env bash

if [[ "${MYSQL_HOST:0:1}" != '/' ]]; then  # Only if we're not connecting over a unix socket
    echo 'Waiting for DB to start at network address.'
    while ! nc -z "$MYSQL_HOST" "$MYSQL_PORT" ; do
        sleep 0.1
    done
fi
echo 'DB started, continuing.'
/investment-tracker

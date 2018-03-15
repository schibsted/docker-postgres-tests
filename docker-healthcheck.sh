#!/bin/bash

# docker-entrypoint starts an postgres temporarily
# ensure the entrypoint script is over
ps aux | grep docker-entrypoint.sh | grep -v grep && exit 1

echo "Entrypoint has finished"

# copy-paste from docker-library: https://github.com/docker-library/postgres/blob/61e369ce3738e38386fa01ce4809c2304e76257c/10/docker-entrypoint.sh

# usage: file_env VAR [DEFAULT]
#    ie: file_env 'XYZ_DB_PASSWORD' 'example'
# (will allow for "$XYZ_DB_PASSWORD_FILE" to fill in the value of
#  "$XYZ_DB_PASSWORD" from a file, especially for Docker's secrets feature)
file_env() {
	local var="$1"
	local fileVar="${var}_FILE"
	local def="${2:-}"
	if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
		echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
		exit 1
	fi
	local val="$def"
	if [ "${!var:-}" ]; then
		val="${!var}"
	elif [ "${!fileVar:-}" ]; then
		val="$(< "${!fileVar}")"
	fi
	export "$var"="$val"
	unset "$fileVar"
}

file_env 'POSTGRES_USER' 'postgres'
file_env 'POSTGRES_DB' "$POSTGRES_USER"
file_env 'POSTGRES_PASSWORD'
file_env 'POSTGRES_HEALTH_QUERY' "SELECT 'uptime: ' ||  now() - pg_postmaster_start_time();"
file_env 'POSTGRES_NO_HEALTH_QUERY'

pg_isready=(pg_isready)

if [ "${POSTGRES_USER}" != "" ]; then
    pg_isready+=(--username "${POSTGRES_USER}")
fi

if [ "${POSTGRES_DB}" != "" ]; then
    pg_isready+=(--dbname "${POSTGRES_DB}")
fi

${pg_isready[@]} || exit 1

echo "Postgres accepts connections"

if [ "${POSTGRES_HEALTH_QUERY}" != "" ] && [ "${POSTGRES_NO_HEALTH_QUERY}" == "" ]; then
    health=(psql -t -v ON_ERROR_STOP=1)

    if [ "${POSTGRES_USER}" != "" ]; then
        health+=(--username "${POSTGRES_USER}")
    fi

    if [ "${POSTGRES_PASSWORD}" != "" ]; then
        export PGPASWORD=${POSTGRES_PASSWORD}
    fi
    
    if [ "${POSTGRES_DB}" != "" ]; then
        health+=(--dbname "${POSTGRES_DB}")
    fi
    echo ${POSTGRES_HEALTH_QUERY} | ${health[@]} || exit 1
    echo "Health query succeed"
fi

exit 0

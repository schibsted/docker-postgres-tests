#!/bin/sh -e

if ! [ -e ${PGDATA}/server.crt ] && ! [ -e ${PGDATA}/server.key ]; then
    openssl req -x509 -newkey rsa:${POSTGRES_SSL_KEY_LENGTH} -keyout ${PGDATA}/server.key -out ${PGDATA}/server.crt -days 3650 -nodes -subj '/'
fi
chmod og-rwx ${PGDATA}/server.key

echo 'ssl = on' >> ${PGDATA}/postgresql.conf

if [ "${POSTGRES_SSL_ONLY}" == "true" ]; then
    sed -i -E '/host[ \t]+all[ \t]+all[ \t]+all[ \t]+trust/d' ${PGDATA}/pg_hba.conf
    echo 'hostnossl    all  all  all  reject' >> ${PGDATA}/pg_hba.conf
    echo 'hostssl      all  all  all  trust' >> ${PGDATA}/pg_hba.conf
fi
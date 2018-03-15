FROM postgres
ADD docker-healthcheck.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-healthcheck.sh
# override devault values see https://docs.docker.com/engine/reference/builder/#healthcheck
HEALTHCHECK --interval=2s --timeout=2s --retries=10 --start-period=500ms CMD ["docker-healthcheck.sh"]

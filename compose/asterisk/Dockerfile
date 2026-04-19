FROM debian:bullseye-slim

RUN set -eux; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      asterisk \
      curl; \
    rm -rf /var/lib/apt/lists/*

COPY manager.conf /etc/asterisk/manager.conf
COPY extensions.conf /etc/asterisk/extensions.conf
COPY queues.conf /etc/asterisk/queues.conf

EXPOSE 5038

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD asterisk -rx 'core show uptime seconds' >/dev/null 2>&1 || exit 1

CMD ["/usr/sbin/asterisk", "-f", "-U", "asterisk", "-G", "asterisk", "-vvv"]

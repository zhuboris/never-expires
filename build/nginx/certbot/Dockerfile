FROM alpine:3.18

COPY crontab /var/spool/cron/crontabs/root
COPY init.sh /init.sh

ENV PATH="/opt/certbot/bin:${PATH}"

RUN apk --no-cache add python3 py3-pip augeas-libs && \
    python3 -m venv /opt/certbot/ && \
    /opt/certbot/bin/pip install --upgrade pip && \
    /opt/certbot/bin/pip install certbot certbot-regru && \
    chmod 0644 /var/spool/cron/crontabs/root && \
    chmod +x /init.sh

CMD ["/init.sh"]
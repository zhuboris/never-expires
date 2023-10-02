FROM grafana/grafana:main-ubuntu

COPY /etc/letsencrypt/live/never-expires.com/ /etc/grafana/ssl/

RUN chown 472:472 /etc/grafana/ssl/fullchain.pem && \
  chown 472:472 /etc/grafana/ssl/privkey.pem

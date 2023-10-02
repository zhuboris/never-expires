FROM alpine:3.18
RUN apk --no-cache add docker ca-certificates
WORKDIR /root/

ADD run_sender.sh /root/run_sender.sh
ADD crontab /var/spool/cron/crontabs/root
RUN chmod 0644 /var/spool/cron/crontabs/root
RUN chmod +x /root/run_sender.sh

CMD crond -f
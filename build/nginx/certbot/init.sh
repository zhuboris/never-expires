#!/bin/sh

chmod 0600 /etc/letsencrypt/regru.ini

/opt/certbot/bin/certbot certonly --cert-name never-expires.com -a dns --email as00as00@mail.ru --agree-tos --no-eff-email -d 'never-expires.com' -d '*.never-expires.com' --dns-propagation-seconds 900 --force-renewal

crond -f
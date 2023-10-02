#!/bin/sh

docker stop notification-notification_sender-1 || true
docker stop notification-redis_db-1 || true

docker start notification-notification_sender-1
docker start notification-redis_db-1

docker wait notification-notification_sender-1
docker stop notification-redis_db-1

#!/bin/sh

export MC_HOSTS_store=http://${MINIO_KEY}:${MINIO_SECRET}@${MINIO_HOST}:9000

mc cp store/builds/${NAME}-${VERSION}.tgz app.tgz
tar -xzf app.tgz -C /app

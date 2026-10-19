#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

RETENTION_DAYS="${RETENTION_DAYS:-14}"
STAMP="$(date +%Y%m%d_%H%M%S)"
FILE="mysql_${STAMP}.sql.gz"

mkdir -p backups

docker compose exec -T mysql sh -c \
    'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysqldump -uroot --single-transaction --routines --triggers --default-character-set=utf8mb4 "$MYSQL_DATABASE" | gzip > /backups/'"$FILE"

if [ ! -s "backups/$FILE" ]; then
    echo "备份失败: backups/$FILE 不存在或为空" >&2
    exit 1
fi

find backups -name 'mysql_*.sql.gz' -mtime "+${RETENTION_DAYS}" -delete

echo "备份完成: backups/$FILE"

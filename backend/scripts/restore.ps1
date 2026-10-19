[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$File,
    [string]$Database = "c5_navigation_restore",
    [string]$ComposeFile = "docker-compose.yml"
)

$ErrorActionPreference = "Stop"

if ($Database -notmatch '^[A-Za-z0-9_]+$') {
    throw "数据库名只允许字母、数字与下划线: $Database"
}

$root = Split-Path -Parent $PSScriptRoot
$composePath = Join-Path $root $ComposeFile
$source = (Resolve-Path $File).Path
if (!(Test-Path $source)) {
    throw "备份文件不存在: $source"
}

Push-Location $root
try {
    & docker compose -f $composePath cp $source mysql:/tmp/restore.sql.gz
    if ($LASTEXITCODE -ne 0) {
        throw "复制备份文件到容器失败"
    }

    $create = 'echo CREATE DATABASE IF NOT EXISTS $1 CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci | MYSQL_PWD=$MYSQL_ROOT_PASSWORD mysql -uroot'
    & docker compose -f $composePath exec -T mysql sh -c $create sh $Database
    if ($LASTEXITCODE -ne 0) {
        throw "创建目标数据库失败"
    }

    $restore = 'gunzip -c /tmp/restore.sql.gz | MYSQL_PWD=$MYSQL_ROOT_PASSWORD mysql -uroot $1'
    & docker compose -f $composePath exec -T mysql sh -c $restore sh $Database
    if ($LASTEXITCODE -ne 0) {
        throw "恢复数据失败"
    }

    & docker compose -f $composePath exec -T mysql rm -f /tmp/restore.sql.gz
}
finally {
    Pop-Location
}

Write-Host "恢复完成: $source -> $Database"
Write-Host "可使用以下命令核对："
Write-Host "  docker compose exec mysql mysql -uroot -p -e `"USE $Database; SHOW TABLES;`""

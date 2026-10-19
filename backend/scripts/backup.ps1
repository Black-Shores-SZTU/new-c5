[CmdletBinding()]
param(
    [string]$ComposeFile = "docker-compose.yml",
    [int]$RetentionDays = 14
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$composePath = Join-Path $root $ComposeFile
$backupDir = Join-Path $root "backups"
if (!(Test-Path $backupDir)) {
    New-Item -ItemType Directory -Path $backupDir | Out-Null
}

$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$fileName = "mysql_$stamp.sql.gz"
$command = 'MYSQL_PWD=$MYSQL_ROOT_PASSWORD mysqldump -uroot --single-transaction --routines --triggers --default-character-set=utf8mb4 $MYSQL_DATABASE | gzip > /backups/' + $fileName

Push-Location $root
try {
    & docker compose -f $composePath exec -T mysql sh -c $command
    if ($LASTEXITCODE -ne 0) {
        throw "mysqldump 失败，退出码 $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}

$filePath = Join-Path $backupDir $fileName
if (!(Test-Path $filePath)) {
    throw "备份文件未生成: $filePath"
}
$size = (Get-Item $filePath).Length
if ($size -le 0) {
    throw "备份文件为空: $filePath"
}

$expired = Get-ChildItem $backupDir -Filter "mysql_*.sql.gz" |
    Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-$RetentionDays) }
foreach ($item in $expired) {
    Remove-Item $item.FullName -Force
}

Write-Host ("备份完成: {0} ({1:N0} bytes)" -f $filePath, $size)
if ($expired.Count -gt 0) {
    Write-Host ("已清理过期备份 {0} 个（保留 {1} 天）" -f $expired.Count, $RetentionDays)
}

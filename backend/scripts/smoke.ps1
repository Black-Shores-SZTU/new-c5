[CmdletBinding()]
param(
    [string]$BaseUrl = "http://localhost:8080",
    [int]$TimeoutSec = 10
)

$ErrorActionPreference = "Stop"
$script:failures = 0

function Pass([string]$name) {
    Write-Host ("PASS  " + $name) -ForegroundColor Green
}

function Fail([string]$name, [string]$detail) {
    $script:failures++
    Write-Host ("FAIL  " + $name + " - " + $detail) -ForegroundColor Red
}

function Assert-True([string]$name, [bool]$condition, [string]$detail = "") {
    if ($condition) {
        Pass $name
    }
    else {
        Fail $name $detail
    }
}

Write-Host "冒烟目标: $BaseUrl"

try {
    $health = Invoke-RestMethod -Uri "$BaseUrl/health" -Method Get -TimeoutSec $TimeoutSec
    Assert-True "GET /health 返回 ok" ($health.status -eq "ok") ("status=" + $health.status)
    Assert-True "GET /health MySQL 正常" ($health.checks.mysql -eq "ok") ("mysql=" + $health.checks.mysql)
}
catch {
    Fail "GET /health" $_.Exception.Message
}

try {
    $headers = @{ "X-Request-ID" = "smoke-request-id" }
    $response = Invoke-WebRequest -Uri "$BaseUrl/buildings" -Method Get -Headers $headers -UseBasicParsing -TimeoutSec $TimeoutSec
    $buildings = $response.Content | ConvertFrom-Json
    $items = @($buildings.data.items)
    Assert-True "GET /buildings 成功" ($buildings.code -eq 0) ("code=" + $buildings.code)
    Assert-True "GET /buildings 包含 C5" ($items.Count -ge 1 -and $items[0].code -eq "C5") ("items=" + $items.Count)
    Assert-True "X-Request-ID 透传" ($response.Headers["X-Request-ID"] -eq "smoke-request-id") ("header=" + $response.Headers["X-Request-ID"])
}
catch {
    Fail "GET /buildings" $_.Exception.Message
}

try {
    $floors = Invoke-RestMethod -Uri "$BaseUrl/buildings/C5/floors" -Method Get -TimeoutSec $TimeoutSec
    $items = @($floors.data.items)
    Assert-True "GET /buildings/C5/floors 成功" ($floors.code -eq 0 -and $items.Count -ge 1) ("items=" + $items.Count)
}
catch {
    Fail "GET /buildings/C5/floors" $_.Exception.Message
}

try {
    $classrooms = Invoke-RestMethod -Uri "$BaseUrl/classrooms?keyword=101&limit=20" -Method Get -TimeoutSec $TimeoutSec
    Assert-True "GET /classrooms 搜索成功" ($classrooms.code -eq 0 -and $classrooms.data.total -ge 1) ("total=" + $classrooms.data.total)
}
catch {
    Fail "GET /classrooms" $_.Exception.Message
}

try {
    $body = @{
        from = @{ building = "C5"; floor = "F1"; room_no = "101" }
        to   = @{ building = "C5"; floor = "F2"; room_no = "201" }
    } | ConvertTo-Json -Depth 5
    $route = Invoke-RestMethod -Uri "$BaseUrl/api/v1/route" -Method Post -ContentType "application/json" -Body $body -TimeoutSec $TimeoutSec
    $segments = @($route.data.segments)
    $transitions = @($route.data.transitions)
    Assert-True "POST /api/v1/route 成功" ($route.code -eq 0) ("code=" + $route.code)
    Assert-True "路线距离大于 0" ($route.data.distance_m -gt 0) ("distance_m=" + $route.data.distance_m)
    Assert-True "路线包含同层分段" ($segments.Count -ge 1) ("segments=" + $segments.Count)
    Assert-True "跨层路线包含转层事件" ($transitions.Count -ge 1) ("transitions=" + $transitions.Count)
}
catch {
    Fail "POST /api/v1/route" $_.Exception.Message
}

try {
    $nearby = Invoke-RestMethod -Uri "$BaseUrl/api/v1/nearby?building=C5&floor=F1&room_no=101&types=restroom,exit" -Method Get -TimeoutSec $TimeoutSec
    $items = @($nearby.data.items)
    $allFound = $true
    foreach ($item in $items) {
        if (-not $item.found) {
            $allFound = $false
        }
    }
    Assert-True "GET /api/v1/nearby 成功" ($nearby.code -eq 0 -and $items.Count -eq 2) ("items=" + $items.Count)
    Assert-True "最近设施均命中" $allFound "存在 found=false"
}
catch {
    Fail "GET /api/v1/nearby" $_.Exception.Message
}

try {
    $swagger = Invoke-WebRequest -Uri "$BaseUrl/swagger/doc.json" -Method Get -UseBasicParsing -TimeoutSec $TimeoutSec
    Assert-True "Swagger 文档可访问" ($swagger.StatusCode -eq 200) ("status=" + $swagger.StatusCode)
}
catch {
    Fail "GET /swagger/doc.json" $_.Exception.Message
}

if ($script:failures -gt 0) {
    Write-Host ("冒烟失败: {0} 项" -f $script:failures) -ForegroundColor Red
    exit 1
}
Write-Host "冒烟全部通过" -ForegroundColor Green

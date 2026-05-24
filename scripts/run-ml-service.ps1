$ErrorActionPreference = "Stop"
$serviceDir = Join-Path $PSScriptRoot "..\services\ml-service"
Push-Location $serviceDir
try {
  py -3 -m uvicorn app:app --host 127.0.0.1 --port 8090
} finally {
  Pop-Location
}


$ErrorActionPreference = "Stop"
$serviceDir = Join-Path $PSScriptRoot "..\services\genai-service"
if (-not $env:GENAI_TIMEOUT_MS) {
  $env:GENAI_TIMEOUT_MS = "35000"
}
if (-not $env:GENAI_MAX_TOKENS) {
  $env:GENAI_MAX_TOKENS = "240"
}
Push-Location $serviceDir
try {
  py -3 -m uvicorn app:app --host 127.0.0.1 --port 8091
} finally {
  Pop-Location
}

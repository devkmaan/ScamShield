$ErrorActionPreference = "Stop"
if (-not $env:PORT) {
  $env:PORT = "8081"
}
if (-not $env:ML_SERVICE_URL) {
  $env:ML_SERVICE_URL = "http://localhost:8090"
}
if (-not $env:GENAI_SERVICE_URL) {
  $env:GENAI_SERVICE_URL = "http://localhost:8091"
}
go run ./cmd/scamshield

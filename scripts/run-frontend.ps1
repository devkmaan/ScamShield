$ErrorActionPreference = "Stop"
Push-Location (Join-Path $PSScriptRoot "..\web")
try {
  npm run dev
} finally {
  Pop-Location
}


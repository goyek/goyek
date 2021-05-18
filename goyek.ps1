$ErrorActionPreference = "Stop"

Push-Location "$PSScriptRoot\build"
& go run . -wd=".." $args
Pop-Location
exit $global:LASTEXITCODE

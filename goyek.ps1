$ErrorActionPreference = "Stop"

Push-Location "$PSScriptRoot\build"
& go run . $args
Pop-Location
exit $global:LASTEXITCODE

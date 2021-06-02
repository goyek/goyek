Push-Location "$PSScriptRoot\build" -ErrorAction Stop
& go run . -wd=".." $args
Pop-Location
exit $global:LASTEXITCODE

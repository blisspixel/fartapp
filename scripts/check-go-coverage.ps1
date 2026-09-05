[CmdletBinding()]
param(
    [string]$ProfilePath = "coverage.out",
    [double]$AggregateMinimum = 90,
    [double]$PackageMinimum = 80
)

$ErrorActionPreference = "Stop"
$repositoryRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repositoryRoot
& go run ./tools/repoquality coverage --profile $ProfilePath --aggregate $AggregateMinimum --package $PackageMinimum
exit $LASTEXITCODE

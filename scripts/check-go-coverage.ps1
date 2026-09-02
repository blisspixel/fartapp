[CmdletBinding()]
param(
    [string]$ProfilePath = "coverage.out",
    [double]$AggregateMinimum = 90,
    [double]$PackageMinimum = 80
)

$ErrorActionPreference = "Stop"
$repositoryRoot = Split-Path -Parent $PSScriptRoot
$resolvedProfile = Join-Path $repositoryRoot $ProfilePath

if (-not (Test-Path -LiteralPath $resolvedProfile -PathType Leaf)) {
    throw "coverage profile not found: $ProfilePath"
}

$modulePath = (& go list -m -f '{{.Path}}').Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($modulePath)) {
    throw "could not resolve the Go module path"
}

$packages = @{}
$profileLines = Get-Content -LiteralPath $resolvedProfile
if ($profileLines.Count -lt 2 -or $profileLines[0] -notmatch '^mode: ') {
    throw "invalid Go coverage profile: $ProfilePath"
}

foreach ($line in $profileLines | Select-Object -Skip 1) {
    if ($line -notmatch '^(?<file>.+):\d+\.\d+,\d+\.\d+\s+(?<statements>\d+)\s+(?<count>\d+)$') {
        throw "invalid Go coverage record: $line"
    }

    $profileFile = $Matches.file
    $statementCount = [long]$Matches.statements
    $executionCount = [long]$Matches.count
    if (-not $profileFile.StartsWith("$modulePath/", [System.StringComparison]::Ordinal)) {
        throw "coverage record is outside the module: $profileFile"
    }

    $relativeFile = $profileFile.Substring($modulePath.Length + 1)
    $localFile = Join-Path $repositoryRoot ($relativeFile.Replace('/', [System.IO.Path]::DirectorySeparatorChar))
    if (-not (Test-Path -LiteralPath $localFile -PathType Leaf)) {
        throw "covered source file not found: $relativeFile"
    }

    $header = (Get-Content -LiteralPath $localFile -TotalCount 5) -join "`n"
    if ($header -match '(?m)^// Code generated .* DO NOT EDIT\.$') {
        continue
    }

    $separator = $profileFile.LastIndexOf('/')
    $packagePath = $profileFile.Substring(0, $separator)
    if (-not $packages.ContainsKey($packagePath)) {
        $packages[$packagePath] = [PSCustomObject]@{
            Statements = 0L
            Covered = 0L
        }
    }

    $packages[$packagePath].Statements += $statementCount
    if ($executionCount -gt 0) {
        $packages[$packagePath].Covered += $statementCount
    }
}

if ($packages.Count -eq 0) {
    throw "coverage profile contains no non-generated package statements"
}

$failures = [System.Collections.Generic.List[string]]::new()
$totalStatements = 0L
$totalCovered = 0L

foreach ($packagePath in $packages.Keys | Sort-Object) {
    $package = $packages[$packagePath]
    if ($package.Statements -eq 0) {
        continue
    }
    $coverage = 100.0 * $package.Covered / $package.Statements
    Write-Output ("package coverage: {0} {1:N1}% ({2}/{3})" -f
        $packagePath,
        $coverage,
        $package.Covered,
        $package.Statements)
    if ($coverage -lt $PackageMinimum) {
        $failures.Add(("{0} is {1:N1}%, below {2:N1}%" -f
            $packagePath,
            $coverage,
            $PackageMinimum))
    }
    $totalStatements += $package.Statements
    $totalCovered += $package.Covered
}

$aggregateCoverage = 100.0 * $totalCovered / $totalStatements
Write-Output ("aggregate non-generated coverage: {0:N1}% ({1}/{2})" -f
    $aggregateCoverage,
    $totalCovered,
    $totalStatements)
if ($aggregateCoverage -lt $AggregateMinimum) {
    $failures.Add(("aggregate is {0:N1}%, below {1:N1}%" -f
        $aggregateCoverage,
        $AggregateMinimum))
}

if ($failures.Count -gt 0) {
    throw "coverage policy failed: $($failures -join '; ')"
}

$ErrorActionPreference = "Stop"

$repositoryRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$lockPath = Join-Path $repositoryRoot "package-lock.json"
$lock = Get-Content -LiteralPath $lockPath -Raw | ConvertFrom-Json -AsHashtable
$failures = [System.Collections.Generic.List[string]]::new()
$allowedLicenses = [System.Collections.Generic.HashSet[string]]::new(
    [System.StringComparer]::Ordinal
)

foreach ($license in @(
    "Apache-2.0",
    "BSD-2-Clause",
    "BSD-3-Clause",
    "ISC",
    "MIT",
    "Python-2.0"
)) {
    [void]$allowedLicenses.Add($license)
}

if ($lock.lockfileVersion -ne 3) {
    $failures.Add("unsupported npm lockfile version: $($lock.lockfileVersion)")
}

$packageCount = 0
$licenseCounts = @{}
foreach ($entry in $lock.packages.GetEnumerator()) {
    if ([string]::IsNullOrEmpty([string]$entry.Key)) {
        continue
    }

    $packageCount++
    $package = $entry.Value
    $license = [string]$package.license
    if (-not $allowedLicenses.Contains($license)) {
        $failures.Add("unreviewed npm license ${license}: $($entry.Key)")
    }
    elseif ($licenseCounts.ContainsKey($license)) {
        $licenseCounts[$license]++
    }
    else {
        $licenseCounts[$license] = 1
    }

    if (-not $package.dev) {
        $failures.Add("npm package is not development-only: $($entry.Key)")
    }
    if ($package.hasInstallScript) {
        $failures.Add("npm install script is not permitted: $($entry.Key)")
    }
    if ([string]$package.integrity -notmatch '^sha512-[A-Za-z0-9+/=]+$') {
        $failures.Add("missing or invalid npm integrity: $($entry.Key)")
    }
    if (-not ([string]$package.resolved).StartsWith(
        "https://registry.npmjs.org/",
        [System.StringComparison]::Ordinal
    )) {
        $failures.Add("unexpected npm source: $($entry.Key)")
    }
}

$goModules = @(go list -m all)
if ($LASTEXITCODE -ne 0) {
    $failures.Add("go list -m all failed")
}
elseif ($goModules.Count -ne 1 -or
    $goModules[0] -ne "github.com/blisspixel/fartapp") {
    $failures.Add(
        "external Go modules require a reviewed dependency and license manifest"
    )
}

if ($failures.Count -gt 0) {
    $failures | ForEach-Object { Write-Error $_ }
    exit 1
}

$licenseSummary = $licenseCounts.GetEnumerator() |
    Sort-Object Name |
    ForEach-Object { "$($_.Name)=$($_.Value)" }

Write-Output (
    "dependency policy verified: $packageCount npm development packages " +
    "($($licenseSummary -join ', ')); no external Go modules"
)

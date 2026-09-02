$ErrorActionPreference = "Stop"

$repositoryRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$mediaRoot = [System.IO.Path]::GetFullPath((Join-Path $repositoryRoot "docs/media"))
$brandRoot = [System.IO.Path]::GetFullPath((Join-Path $repositoryRoot "brand/source"))
$assetRoots = @($mediaRoot, $brandRoot)
$manifestPath = Join-Path $mediaRoot "manifest.json"
$manifest = Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
$failures = [System.Collections.Generic.List[string]]::new()
$registered = [System.Collections.Generic.HashSet[string]]::new(
    [System.StringComparer]::Ordinal
)

function Test-AssetRoot {
    param([string]$Path)

    foreach ($root in $assetRoots) {
        if ($Path.StartsWith(
            $root + [System.IO.Path]::DirectorySeparatorChar,
            [System.StringComparison]::OrdinalIgnoreCase
        )) {
            return $true
        }
    }
    return $false
}

function Get-UInt24LittleEndian {
    param([byte[]]$Bytes, [int]$Offset)

    return [int]$Bytes[$Offset] -bor
        ([int]$Bytes[$Offset + 1] -shl 8) -bor
        ([int]$Bytes[$Offset + 2] -shl 16)
}

function Get-AssetDimensions {
    param([string]$Path)

    $extension = [System.IO.Path]::GetExtension($Path).ToLowerInvariant()
    if ($extension -eq ".svg") {
        $header = Get-Content -LiteralPath $Path -Raw
        $widthMatch = [regex]::Match($header, '<svg[^>]*\bwidth="(?<value>\d+)"')
        $heightMatch = [regex]::Match($header, '<svg[^>]*\bheight="(?<value>\d+)"')
        if (-not $widthMatch.Success -or -not $heightMatch.Success) {
            throw "SVG has no integer width and height: $Path"
        }
        return @(
            [int]$widthMatch.Groups["value"].Value,
            [int]$heightMatch.Groups["value"].Value
        )
    }

    $bytes = [System.IO.File]::ReadAllBytes($Path)
    if ($extension -eq ".png") {
        if ($bytes.Length -lt 24 -or
            [System.Text.Encoding]::ASCII.GetString($bytes, 1, 3) -ne "PNG") {
            throw "invalid PNG header: $Path"
        }
        $width = [System.Net.IPAddress]::NetworkToHostOrder(
            [System.BitConverter]::ToInt32($bytes, 16)
        )
        $height = [System.Net.IPAddress]::NetworkToHostOrder(
            [System.BitConverter]::ToInt32($bytes, 20)
        )
        return @($width, $height)
    }

    if ($extension -eq ".webp") {
        if ($bytes.Length -lt 30 -or
            [System.Text.Encoding]::ASCII.GetString($bytes, 0, 4) -ne "RIFF" -or
            [System.Text.Encoding]::ASCII.GetString($bytes, 8, 4) -ne "WEBP") {
            throw "invalid WebP header: $Path"
        }

        $chunk = [System.Text.Encoding]::ASCII.GetString($bytes, 12, 4)
        if ($chunk -eq "VP8X") {
            $width = (Get-UInt24LittleEndian $bytes 24) + 1
            $height = (Get-UInt24LittleEndian $bytes 27) + 1
            return @($width, $height)
        }
        if ($chunk -eq "VP8 ") {
            if ($bytes[23] -ne 0x9d -or $bytes[24] -ne 0x01 -or
                $bytes[25] -ne 0x2a) {
                throw "invalid lossy WebP frame header: $Path"
            }
            $width = [System.BitConverter]::ToUInt16($bytes, 26) -band 0x3fff
            $height = [System.BitConverter]::ToUInt16($bytes, 28) -band 0x3fff
            return @($width, $height)
        }
        if ($chunk -eq "VP8L") {
            if ($bytes[20] -ne 0x2f) {
                throw "invalid lossless WebP frame header: $Path"
            }
            $width = 1 + $bytes[21] + (($bytes[22] -band 0x3f) -shl 8)
            $height = 1 + (($bytes[22] -band 0xc0) -shr 6) +
                ($bytes[23] -shl 2) + (($bytes[24] -band 0x0f) -shl 10)
            return @($width, $height)
        }
        throw "unsupported WebP chunk $chunk in $Path"
    }

    throw "unsupported media extension $extension in $Path"
}

if ($manifest.schemaVersion -ne 2) {
    $failures.Add("unsupported media manifest schema: $($manifest.schemaVersion)")
}

foreach ($asset in $manifest.assets) {
    $relativePath = [string]$asset.path
    $normalizedRelative = $relativePath.Replace("\", "/")
    if (-not $registered.Add($normalizedRelative)) {
        $failures.Add("duplicate media manifest path: $normalizedRelative")
        continue
    }

    $absolutePath = [System.IO.Path]::GetFullPath(
        (Join-Path $repositoryRoot $relativePath)
    )
    if (-not (Test-AssetRoot $absolutePath)) {
        $failures.Add("asset path escapes approved roots: $normalizedRelative")
        continue
    }

    if (-not (Test-Path -LiteralPath $absolutePath -PathType Leaf)) {
        $failures.Add("missing media file: $normalizedRelative")
        continue
    }

    $file = Get-Item -LiteralPath $absolutePath
    if ($file.Length -ne [long]$asset.bytes) {
        $failures.Add(
            "byte count drift for ${normalizedRelative}: expected $($asset.bytes), got $($file.Length)"
        )
    }
    if ($file.Length -gt [long]$asset.maxBytes) {
        $failures.Add(
            "byte budget exceeded for ${normalizedRelative}: $($file.Length) > $($asset.maxBytes)"
        )
    }

    try {
        $dimensions = Get-AssetDimensions $absolutePath
        if ($dimensions[0] -ne [int]$asset.width -or
            $dimensions[1] -ne [int]$asset.height) {
            $failures.Add(
                "dimension drift for ${normalizedRelative}: expected $($asset.width)x$($asset.height), got $($dimensions[0])x$($dimensions[1])"
            )
        }
    }
    catch {
        $failures.Add($_.Exception.Message)
    }

    $digest = (Get-FileHash -LiteralPath $absolutePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($digest -ne ([string]$asset.sha256).ToLowerInvariant()) {
        $failures.Add("SHA-256 drift for $normalizedRelative")
    }
    if ([string]$asset.sha256 -notmatch '^[0-9a-fA-F]{64}$') {
        $failures.Add("invalid SHA-256 format for $normalizedRelative")
    }

    if ([string]::IsNullOrWhiteSpace([string]$asset.alt)) {
        $failures.Add("missing alt text for $normalizedRelative")
    }
    if ($asset.status -notin @("current", "planned")) {
        $failures.Add("invalid status for ${normalizedRelative}: $($asset.status)")
    }
    foreach ($field in @(
        "origin",
        "license",
        "rightsStatus",
        "rightsBasis",
        "rightsReviewedAt",
        "reviewerRole"
    )) {
        if ([string]::IsNullOrWhiteSpace([string]$asset.$field)) {
            $failures.Add("missing $field for $normalizedRelative")
        }
    }
    if ($asset.rightsStatus -ne "approved-for-public-repository") {
        $failures.Add("asset is not approved for public repository: $normalizedRelative")
    }
    if ($asset.PSObject.Properties.Name -notcontains "replacementHistory") {
        $failures.Add("missing replacementHistory for $normalizedRelative")
    }
    if ($asset.origin -eq "project-generated-concept") {
        foreach ($field in @(
            "generator",
            "model",
            "generatedAt",
            "sanitizedPrompt",
            "sourceAssetSha256",
            "postProcessing"
        )) {
            if ([string]::IsNullOrWhiteSpace([string]$asset.provenance.$field)) {
                $failures.Add("missing provenance.$field for $normalizedRelative")
            }
        }
        if ([string]$asset.provenance.sourceAssetSha256 -notmatch
            '^[0-9a-fA-F]{64}$') {
            $failures.Add("invalid provenance source hash for $normalizedRelative")
        }
        if ($asset.provenance.PSObject.Properties.Name -notcontains
            "inputAssets") {
            $failures.Add("missing provenance.inputAssets for $normalizedRelative")
        }
        else {
            foreach ($inputAsset in @($asset.provenance.inputAssets)) {
                if ([string]$inputAsset.sha256 -notmatch '^[0-9a-fA-F]{64}$') {
                    $failures.Add("invalid provenance input hash for $normalizedRelative")
                }
                if ([string]::IsNullOrWhiteSpace([string]$inputAsset.role)) {
                    $failures.Add("missing provenance input role for $normalizedRelative")
                }
                $inputIdentity = [string]$inputAsset.repositoryPathAtGeneration
                if ([string]::IsNullOrWhiteSpace($inputIdentity)) {
                    $failures.Add("missing provenance input identity for $normalizedRelative")
                }
            }
        }
    }
    elseif ([string]::IsNullOrWhiteSpace([string]$asset.sourceRevision)) {
        $failures.Add("missing sourceRevision for $normalizedRelative")
    }

    foreach ($reference in $asset.referencedBy) {
        $referencePath = Join-Path $repositoryRoot ([string]$reference)
        if (-not (Test-Path -LiteralPath $referencePath -PathType Leaf)) {
            $failures.Add("missing reference file $reference for $normalizedRelative")
            continue
        }
        $referenceText = Get-Content -LiteralPath $referencePath -Raw
        if (-not $referenceText.Contains($normalizedRelative)) {
            $failures.Add("$reference does not reference $normalizedRelative")
        }
    }
}

$mediaFiles = foreach ($root in $assetRoots) {
    Get-ChildItem -LiteralPath $root -Recurse -File |
        Where-Object { $_.Name -notin @("README.md", "manifest.json") }
}

foreach ($file in $mediaFiles) {
    $relative = [System.IO.Path]::GetRelativePath(
        $repositoryRoot,
        $file.FullName
    ).Replace("\", "/")
    if (-not $registered.Contains($relative)) {
        $failures.Add("orphan media file: $relative")
    }
}

if ($failures.Count -gt 0) {
    $failures | ForEach-Object { Write-Error $_ }
    exit 1
}

Write-Output "media manifest verified: $($registered.Count) assets"

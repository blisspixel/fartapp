$ErrorActionPreference = "Stop"

$repositoryRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$failures = [System.Collections.Generic.List[string]]::new()
$checkedLinks = 0
$linkPattern = [regex]'!?\[[^\]]*\]\((?<target><[^>]+>|[^)\s]+)'
$pathComparison = if ($IsWindows) {
    [System.StringComparison]::OrdinalIgnoreCase
}
else {
    [System.StringComparison]::Ordinal
}

$markdownFiles = Get-ChildItem -LiteralPath $repositoryRoot -Recurse -File -Filter "*.md" |
    Where-Object {
        $relativePath = [System.IO.Path]::GetRelativePath(
            $repositoryRoot,
            $_.FullName
        ).Replace("\", "/")
        -not (
            $relativePath.StartsWith("node_modules/", [System.StringComparison]::Ordinal) -or
            $relativePath.StartsWith("vendor/", [System.StringComparison]::Ordinal) -or
            $relativePath.StartsWith(".git/", [System.StringComparison]::Ordinal) -or
            $relativePath.StartsWith(".steward/DECISIONS/", [System.StringComparison]::Ordinal)
        )
    }

foreach ($file in $markdownFiles) {
    $text = Get-Content -LiteralPath $file.FullName -Raw
    foreach ($match in $linkPattern.Matches($text)) {
        $target = $match.Groups["target"].Value.Trim("<", ">")
        if ($target -match '^(?:https?://|mailto:|#)') {
            continue
        }

        $pathPart = ($target -split "#", 2)[0]
        if ([string]::IsNullOrWhiteSpace($pathPart)) {
            continue
        }

        try {
            $decodedPath = [System.Uri]::UnescapeDataString($pathPart)
            $candidate = [System.IO.Path]::GetFullPath(
                (Join-Path $file.DirectoryName $decodedPath)
            )
        }
        catch {
            $failures.Add("invalid local link in $($file.Name): $target")
            continue
        }

        $checkedLinks++
        if (-not $candidate.StartsWith(
            $repositoryRoot + [System.IO.Path]::DirectorySeparatorChar,
            $pathComparison
        )) {
            $failures.Add("local link escapes repository: $($file.Name) -> $target")
            continue
        }
        if (-not (Test-Path -LiteralPath $candidate)) {
            $relativeFile = [System.IO.Path]::GetRelativePath(
                $repositoryRoot,
                $file.FullName
            ).Replace("\", "/")
            $failures.Add("missing local link: $relativeFile -> $target")
        }
    }
}

if ($failures.Count -gt 0) {
    $failures | ForEach-Object { Write-Error $_ }
    exit 1
}

Write-Output "local Markdown links verified: $checkedLinks links in $($markdownFiles.Count) files"

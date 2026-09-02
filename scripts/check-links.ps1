$ErrorActionPreference = "Stop"

$repositoryRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$failures = [System.Collections.Generic.List[string]]::new()
$checkedLinks = 0
$linkPattern = [regex]'!?\[[^\]]*\]\((?<target><[^>]+>|[^)\s]+)'

$markdownFiles = Get-ChildItem -LiteralPath $repositoryRoot -Recurse -File -Filter "*.md" |
    Where-Object {
        $_.FullName -notlike "*\node_modules\*" -and
        $_.FullName -notlike "*\.git\*" -and
        $_.FullName -notlike "*\.steward\DECISIONS\*"
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
            [System.StringComparison]::OrdinalIgnoreCase
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

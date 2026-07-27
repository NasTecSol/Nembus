# Script to install a Git post-commit hook for automatically keeping notebook_sources updated

$hookPath = ".git/hooks/post-commit"

$hookContent = @"
#!/bin/sh
# Automatically regenerate NotebookLM source bundles on git commit
echo "[Git Hook] Regenerating NotebookLM context sources..."
powershell -ExecutionPolicy Bypass -File scripts/build_notebook_sources.ps1
"@

Set-Content -Path $hookPath -Value $hookContent -Encoding utf8
Write-Host "Git post-commit hook installed successfully at $hookPath"

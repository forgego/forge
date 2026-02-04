#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
report_dir="$repo_root/reports"
report_file="$report_dir/unused-code.md"

audit_status=0

mkdir -p "$report_dir"

{
  echo "# Unused code audit"
  echo ""
  echo "Generated: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo ""
  echo "## Module checks"
} > "$report_file"

run_module_checks() {
  local module_dir="$1"
  local module_name

  module_name=$(basename "$module_dir")

  {
    echo ""
    echo "### ${module_name}"
  } >> "$report_file"

  for cmd in "go list ./..." "go vet ./..." "golangci-lint run --enable=unused ./..."; do
    {
      echo ""
      echo "#### ${cmd}"
      echo "\`\`\`"
    } >> "$report_file"

    if output=$(cd "$module_dir" && eval "$cmd" 2>&1); then
      echo "$output" >> "$report_file"
    else
      audit_status=1
      echo "$output" >> "$report_file"
    fi

    echo "\`\`\`" >> "$report_file"
  done
}

run_module_checks "$repo_root/forge"

for module_dir in "$repo_root/examples"/*; do
  if [[ -f "$module_dir/go.mod" ]]; then
    run_module_checks "$module_dir"
  fi
done

{
  echo ""
  echo "## Potential orphaned files (rg scan)"
  echo ""
  echo "Files listed below are not referenced by path or basename within the repository (excluding reports)."
  echo "Review before removal or refactor."
  echo ""
} >> "$report_file"

orphans=()
while IFS= read -r file_path; do
  rel_path=${file_path#"$repo_root/"}
  base_name=$(basename "$file_path")

  if rg -F --quiet "$rel_path" "$repo_root" --glob '!reports/**' --glob '!**/.git/**'; then
    continue
  fi

  if rg -F --quiet "$base_name" "$repo_root" --glob '!reports/**' --glob "!${rel_path}" --glob '!**/.git/**'; then
    continue
  fi

  orphans+=("$rel_path")
done < <(find "$repo_root/examples" -type f -not -path '*/.git/*' -not -name 'go.mod' -not -name 'go.sum')

if ((${#orphans[@]})); then
  audit_status=1
  for orphan in "${orphans[@]}"; do
    echo "- \
\`$orphan\` - No references found via ripgrep; verify if required by runtime or documentation." >> "$report_file"
  done
else
  echo "- No orphaned files detected." >> "$report_file"
fi

exit "$audit_status"

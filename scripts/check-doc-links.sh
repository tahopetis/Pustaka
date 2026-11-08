#!/usr/bin/env bash
# check-doc-links.sh - verify that relative Markdown links in the tracked
# documentation resolve to files that exist in this repository.
#
# Scope: README.md, CLAUDE.md, GLOSSARY.md and docs/**/*.md.
# docs/archive/** is deliberately excluded: archived documents are kept for
# provenance and are not maintained, so their links are allowed to rot.
#
# Ignored link targets: http(s) and other URL schemes, mailto:, bare
# "#anchor" links, and absolute paths. Anchors on local files are stripped
# before the file check, so "docs/status.md#implemented" only verifies
# "docs/status.md".
#
# Output: one "file:line: unresolved link: target" line per broken link.
# Exit status: 0 when every link resolves, 1 when at least one does not,
# 2 when the script cannot run (for example because no docs were found).

set -u

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)" || exit 2
cd "$repo_root" || exit 2

files=()
for f in README.md CLAUDE.md GLOSSARY.md; do
    [ -f "$f" ] && files+=("$f")
done
while IFS= read -r f; do
    files+=("$f")
done < <(find docs -type f -name '*.md' -not -path 'docs/archive/*' 2>/dev/null | sort)

if [ "${#files[@]}" -eq 0 ]; then
    echo "check-doc-links: no Markdown files found under $repo_root" >&2
    exit 2
fi

broken=0

for file in "${files[@]}"; do
    base_dir="$(dirname "$file")"

    # One record per line: "line-number<TAB>raw-target". Code fences are
    # skipped so shell snippets and JSON examples never look like links.
    while IFS=$'\t' read -r lineno target; do
        [ -n "$target" ] || continue

        # Bare anchors and external URLs are out of scope.
        case "$target" in
            \#*) continue ;;
        esac
        if [[ "$target" =~ ^[A-Za-z][A-Za-z0-9+.-]*: ]]; then
            continue
        fi
        # Absolute paths are host-specific and out of scope.
        case "$target" in
            /*) continue ;;
        esac

        path="$target"
        path="${path#<}"
        path="${path%>}"
        path="${path%%#*}"
        path="${path%%\?*}"
        [ -n "$path" ] || continue

        if [ ! -e "$base_dir/$path" ]; then
            printf '%s:%s: unresolved link: %s\n' "$file" "$lineno" "$target"
            broken=$((broken + 1))
        fi
    done < <(awk '
        function emit(target) {
            gsub(/^[[:space:]]+/, "", target)
            gsub(/[[:space:]]+$/, "", target)
            if (target != "") print NR "\t" target
        }
        /^[[:space:]]*(```|~~~)/ { in_fence = !in_fence; next }
        in_fence { next }
        {
            line = $0

            # Reference definitions: [label]: target
            if (match(line, /^[[:space:]]*\[[^]]+\]:[[:space:]]*/)) {
                emit(substr(line, RLENGTH + 1))
                next
            }

            # Inline links and images: ](target) or ](target "title")
            while (match(line, /\]\([^()]*\)/)) {
                target = substr(line, RSTART + 2, RLENGTH - 3)
                line = substr(line, RSTART + RLENGTH)
                sub(/^[[:space:]]*/, "", target)
                if (target ~ /^</) {
                    sub(/^</, "", target)
                    sub(/>.*$/, "", target)
                } else {
                    sub(/[[:space:]]+["\x27(].*$/, "", target)
                }
                emit(target)
            }
        }
    ' "$file")
done

if [ "$broken" -gt 0 ]; then
    printf '%s unresolved documentation link(s)\n' "$broken" >&2
    exit 1
fi

printf 'check-doc-links: all relative documentation links resolve (%s files)\n' "${#files[@]}"
exit 0

# Version management recipes
# Call as: just version::show, just version::bump minor, etc.

PROJECT_ROOT := justfile_directory()
CHART_FILE := PROJECT_ROOT + "/chart/Chart.yaml"
CHART_VALUES := PROJECT_ROOT + "/chart/values.yaml"

# Show current versions
show:
    @echo "Chart version:  $(grep '^version:' {{ CHART_FILE }} | awk '{print $2}')"
    @echo "App version:    $(grep '^appVersion:' {{ CHART_FILE }} | sed 's/appVersion: *"\(.*\)"/\1/')"
    @echo "DICEbear image: $(grep 'tag:' {{ CHART_VALUES }} | head -1 | awk '{print $2}' | tr -d '\"')"
    @echo "Shim image:     $(grep 'tag:' {{ CHART_VALUES }} | sed -n '2p' | awk '{print $2}' | tr -d '\"')"
    @echo "Git tag:        $(git -C {{ PROJECT_ROOT }} describe --tags --always 2>/dev/null || echo 'none')"

# Bump version, update chart, tag, and push (usage: just version::bump patch|minor|major)
bump level:
    #!/usr/bin/env bash
    set -euo pipefail
    CURRENT=$(grep '^version:' {{ CHART_FILE }} | awk '{print $2}')
    IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"
    case "{{ level }}" in
        major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
        minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
        patch) PATCH=$((PATCH + 1)) ;;
        *) echo "Usage: just version::bump [major|minor|patch]"; exit 1 ;;
    esac
    NEW="${MAJOR}.${MINOR}.${PATCH}"
    echo "Bumping: ${CURRENT} -> ${NEW}"
    sed -i "s/^version:.*/version: ${NEW}/" {{ CHART_FILE }}
    sed -i "s/^appVersion:.*/appVersion: \"${NEW}\"/" {{ CHART_FILE }}
    git -C {{ PROJECT_ROOT }} add chart/Chart.yaml
    git -C {{ PROJECT_ROOT }} commit -m "chore: bump version to v${NEW}"
    git -C {{ PROJECT_ROOT }} tag "v${NEW}"
    echo "Tagged v${NEW}. Push with: git push && git push --tags"

# Update DICEbear API image tag in chart defaults
set-dicebear tag:
    #!/usr/bin/env bash
    set -euo pipefail
    sed -i '0,/tag: "[^"]*"/{ /dicebear/,/tag:/{ s/tag: "[^"]*"/tag: "{{ tag }}"/ } }' {{ CHART_VALUES }}
    echo "DICEbear image tag updated to {{ tag }}"
    git -C {{ PROJECT_ROOT }} add chart/values.yaml
    git -C {{ PROJECT_ROOT }} commit -m "chore: update dicebear/api image to {{ tag }}"

# Update shim image tag in chart defaults
set-shim tag:
    #!/usr/bin/env bash
    set -euo pipefail
    awk '/^  shim:/,/tag:/{sub(/tag: "[^"]*"/, "tag: \"{{ tag }}\"")}1' {{ CHART_VALUES }} > {{ CHART_VALUES }}.tmp
    mv {{ CHART_VALUES }}.tmp {{ CHART_VALUES }}
    echo "Shim image tag updated to {{ tag }}"
    git -C {{ PROJECT_ROOT }} add chart/values.yaml
    git -C {{ PROJECT_ROOT }} commit -m "chore: update dicebear-shim image to {{ tag }}"

# Full release: bump version, update shim tag to match, tag and push
release level:
    #!/usr/bin/env bash
    set -euo pipefail
    CURRENT=$(grep '^version:' {{ CHART_FILE }} | awk '{print $2}')
    IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"
    case "{{ level }}" in
        major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
        minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
        patch) PATCH=$((PATCH + 1)) ;;
        *) echo "Usage: just version::release [major|minor|patch]"; exit 1 ;;
    esac
    NEW="${MAJOR}.${MINOR}.${PATCH}"
    echo "Releasing: ${CURRENT} -> ${NEW}"
    # Update chart version + appVersion
    sed -i "s/^version:.*/version: ${NEW}/" {{ CHART_FILE }}
    sed -i "s/^appVersion:.*/appVersion: \"${NEW}\"/" {{ CHART_FILE }}
    # Update shim image tag to match release
    awk "/^  shim:/,/tag:/{sub(/tag: \"[^\"]*\"/, \"tag: \\\"v${NEW}\\\"\")}" {{ CHART_VALUES }} > {{ CHART_VALUES }}.tmp
    mv {{ CHART_VALUES }}.tmp {{ CHART_VALUES }}
    git -C {{ PROJECT_ROOT }} add chart/Chart.yaml chart/values.yaml
    git -C {{ PROJECT_ROOT }} commit -m "chore: release v${NEW}"
    git -C {{ PROJECT_ROOT }} tag "v${NEW}"
    echo "Released v${NEW}. Push with: git push && git push --tags"

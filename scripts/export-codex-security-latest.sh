#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$REPO_ROOT/codex-security-output"
MODE="full"
FULL_KEEP_FILES=0

for arg in "$@"; do
  if [[ "$arg" == "--full" ]]; then
    MODE="full"
    FULL_KEEP_FILES=0
  elif [[ "$arg" == "--full-keep-files" ]]; then
    MODE="full"
    FULL_KEEP_FILES=1
  elif [[ "$arg" == "--compact" ]]; then
    MODE="compact"
    FULL_KEEP_FILES=0
  elif [[ "$arg" == "--help" || "$arg" == "-h" ]]; then
    cat <<'EOF'
Usage: export-codex-security-latest.sh [OUTPUT_DIR] [--compact|--full|--full-keep-files]

Default (base behavior):
  - OUTPUT_DIR: repository-root/codex-security-output
  - MODE: full

Modes:
  --compact         Save only latest-report.md, latest-report-ja.md, latest-findings.json, latest-scan-metadata.txt.
  --full             (default) Save all full artifacts in a single archive and a stable symlink `latest`.
  --full-keep-files
             Keep legacy full artifacts in a timestamped directory plus latest-* copies.
EOF
    exit 0
  else
    OUTPUT_DIR="$arg"
  fi
done

if [[ -z "${OUTPUT_DIR}" ]]; then
  echo "Output directory is required." >&2
  exit 1
fi

# base behavior is full (archive), compact is opt-in
if [[ "$MODE" != "compact" && "$MODE" != "full" ]]; then
  echo "Unknown mode: ${MODE}" >&2
  exit 1
fi

CLI_OUTPUT="$(corepack pnpm exec codex-security scans list --format json)"

SCAN_INFO="$(node -e '
const fs = require("fs");
const path = require("path");

const data = JSON.parse(process.argv[1] || "{}");
const mode = process.argv[2];
const requiredArtifacts = mode === "full"
  ? ["report.md", "findings.json", "coverage.json", "scan-manifest.json", "exports/results.sarif"]
  : ["report.md", "findings.json"];

const scans = Array.isArray(data.scans) ? data.scans : [];
const scan = scans
  .filter((candidate) =>
    candidate?.progress?.status === "complete" &&
    candidate.completedAt &&
    candidate.scanId &&
    candidate.scanDir &&
    requiredArtifacts.every((artifact) =>
      fs.existsSync(path.join(candidate.scanDir, artifact)),
    ),
  )
  .sort((a, b) => Date.parse(b.completedAt) - Date.parse(a.completedAt))[0];

if (!scan) {
  console.error(`No completed codex-security scan with ${mode} artifacts found.`);
  process.exit(1);
}

console.log(scan.scanId);
console.log(scan.scanDir);
' "$CLI_OUTPUT" "$MODE")"

SCAN_ID="$(printf '%s\n' "$SCAN_INFO" | sed -n '1p')"
SCAN_DIR="$(printf '%s\n' "$SCAN_INFO" | sed -n '2p')"

if [[ -z "$SCAN_DIR" || -z "$SCAN_ID" ]]; then
  echo "No completed codex-security scan found." >&2
  exit 1
fi

mkdir -p "$OUTPUT_DIR"

# Normalize OUTPUT_DIR across mode switches so intermediate/full leftovers do not accumulate.
# Optional $1: absolute path of the active temp dir to preserve (archive staging).
clean_hidden_temp_dirs() {
  local keep_path="${1:-}"
  local d
  for d in "$OUTPUT_DIR"/.full-export-*; do
    [[ -e "$d" ]] || continue
    if [[ -n "$keep_path" && "$d" == "$keep_path" ]]; then
      continue
    fi
    rm -rf "$d"
  done
}

clean_timestamped_dirs() {
  local keep_basename="${1:-}"
  local d
  for d in "$OUTPUT_DIR"/20??????_??????-*; do
    [[ -e "$d" ]] || continue
    if [[ -n "$keep_basename" && "$(basename "$d")" == "$keep_basename" ]]; then
      continue
    fi
    [[ -d "$d" ]] && rm -rf "$d"
  done
}

clean_archives() {
  if compgen -G "$OUTPUT_DIR"/*.tar.gz > /dev/null; then
    rm -f "$OUTPUT_DIR"/*.tar.gz
  fi
}

clean_root_intermediates() {
  rm -f \
    "$OUTPUT_DIR/report.md" \
    "$OUTPUT_DIR/findings.json" \
    "$OUTPUT_DIR/coverage.json" \
    "$OUTPUT_DIR/scan-manifest.json" \
    "$OUTPUT_DIR/results.sarif" \
    "$OUTPUT_DIR/scan-show.json" \
    "$OUTPUT_DIR/scan-metadata.txt" \
    "$OUTPUT_DIR/report-ja.md"
}

clean_all_latest_flats() {
  rm -f \
    "$OUTPUT_DIR/latest-report.md" \
    "$OUTPUT_DIR/latest-findings.json" \
    "$OUTPUT_DIR/latest-report-ja.md" \
    "$OUTPUT_DIR/latest-scan-metadata.txt" \
    "$OUTPUT_DIR/latest-coverage.json" \
    "$OUTPUT_DIR/latest-scan-manifest.json" \
    "$OUTPUT_DIR/latest-results.sarif" \
    "$OUTPUT_DIR/latest-scan-show.json"
}

clean_full_only_latest_flats() {
  rm -f \
    "$OUTPUT_DIR/latest-coverage.json" \
    "$OUTPUT_DIR/latest-scan-manifest.json" \
    "$OUTPUT_DIR/latest-results.sarif" \
    "$OUTPUT_DIR/latest-scan-show.json"
}

if [[ "$MODE" == "full" ]]; then
  TIMESTAMP="$(date '+%Y%m%d_%H%M%S')"
  if [[ "$FULL_KEEP_FILES" -eq 1 ]]; then
    DEST_DIR="${OUTPUT_DIR}/${TIMESTAMP}-${SCAN_ID}"
  else
    DEST_DIR="$(mktemp -d "${OUTPUT_DIR}/.full-export-${TIMESTAMP}-XXXXXX")"
  fi
  mkdir -p "$DEST_DIR"
else
  DEST_DIR="$OUTPUT_DIR"
fi

cp "$SCAN_DIR/report.md" "$DEST_DIR/report.md"
cp "$SCAN_DIR/findings.json" "$DEST_DIR/findings.json"
if [[ "$MODE" == "full" ]]; then
  cp "$SCAN_DIR/coverage.json" "$DEST_DIR/coverage.json"
  cp "$SCAN_DIR/scan-manifest.json" "$DEST_DIR/scan-manifest.json"
  cp "$SCAN_DIR/exports/results.sarif" "$DEST_DIR/results.sarif"
fi

SCAN_SHOW_PATH="${DEST_DIR}/scan-show.json"
corepack pnpm exec codex-security scans show "$SCAN_ID" --format json > "$SCAN_SHOW_PATH"

node - "$SCAN_SHOW_PATH" "$DEST_DIR/report-ja.md" <<'NODE'
const fs = require('fs');

const sourcePath = process.argv[2];
const outputPath = process.argv[3];

const scan = JSON.parse(fs.readFileSync(sourcePath, 'utf8'));
const severityLabel = { high: '高', medium: '中', low: '低', critical: '重大' };
const confidenceLabel = { high: '高', medium: '中', low: '低' };
const roleLabel = {
  entrypoint: 'エントリーポイント',
  source: 'ソース',
  sink: 'シンク',
  concrete_implementation: '実装',
  root_control: '制御点',
  evidence: '証跡',
  entrypoint_wrapper: 'ラッパー',
  concrete: '具体実装',
  'entrypoint/wrapper': 'エントリーポイント/ラッパー',
};

const lines = [];
const date = new Date(scan.updatedAt || scan.createdAt || new Date().toISOString());
lines.push(`# Codex Security レポート（日本語要約）`);
lines.push(`対象リポジトリ: ${scan.target?.displayName || scan.contract?.target?.displayName || 'N/A'}`);
lines.push(`対象パス: ${scan.targetPath || 'N/A'}`);
lines.push(`スキャンID: ${scan.scanId || 'N/A'}`);
lines.push(`実行日時: ${isNaN(date.getTime()) ? 'N/A' : date.toLocaleString('ja-JP', { timeZone: 'Asia/Tokyo' })}`);
lines.push(`進捗ステータス: ${scan.progress?.status || 'unknown'}`);
lines.push(`スコープ: ${scan.scope || '.'}`);
lines.push(`モード: ${scan.mode || 'N/A'}`);
lines.push('');
lines.push('## 要約');
lines.push(`- 検出件数: **${scan.findingCount ?? 0}件**`);
if (scan.severityCounts) {
  const sev = scan.severityCounts;
  lines.push(`- 深刻度別: ${Object.entries(sev).map(([k, v]) => `${severityLabel[k] || k}: ${v}件`).join(' / ')}`);
}
if (scan.cost) {
  lines.push(`- コスト（推定）: \$${scan.cost.estimatedUsd ?? 'N/A'}`);
}
if (scan.progress?.coverage) {
  const c = scan.progress.coverage;
  lines.push(`- カバレッジ: ${c.closedRows ?? 0}/${c.filesTotal ?? 0}（partial）`);
}
lines.push('');
lines.push('## 検出内容');

const findings = Array.isArray(scan.findings) ? scan.findings : [];
if (!findings.length) {
  lines.push('検出なし');
} else {
  findings.forEach((f, idx) => {
    const sev = f.severity?.level || f.severity || 'low';
    const conf = f.confidence?.level || 'low';
    lines.push('');
    lines.push(`### [${idx + 1}] ${f.title || '（タイトル不明）'}`);
    lines.push(`- 重大度: ${severityLabel[sev] || sev}`);
    lines.push(`- 信頼度: ${confidenceLabel[conf] || conf}`);
    lines.push(`- CWE: ${(f.taxonomy?.cwe || []).join(', ') || 'N/A'}`);
    lines.push(`- 影響: ${f.impact || 'N/A'}`);
    if (f.summary) {
      lines.push('');
      lines.push('#### 概要');
      lines.push(f.summary);
    }
    if (f.rootCause) {
      lines.push('');
      lines.push('#### 根本原因');
      lines.push(typeof f.rootCause === 'string' ? f.rootCause : f.rootCause.summary || 'N/A');
    }
    if (Array.isArray(f.remediation)) {
      lines.push('');
      lines.push('#### 推奨対応');
      lines.push(...f.remediation.map((r) => `- ${r}`));
    } else if (typeof f.remediation === 'string' && f.remediation.trim()) {
      lines.push('');
      lines.push('#### 推奨対応');
      lines.push(f.remediation);
    }
    if (Array.isArray(f.locations) && f.locations.length > 0) {
      lines.push('');
      lines.push('#### 対象箇所');
      f.locations.forEach((loc) => {
        const role = roleLabel[loc.role] || loc.role || 'unknown';
        const pathText = loc.absolutePath || loc.path || '';
        const lineText = [loc.startLine, loc.endLine].filter(Boolean).join('-');
        lines.push(`- ${role}: \`${pathText}\`${lineText ? ` (${lineText})` : ''}`);
      });
    }
  });
}

lines.push('');
lines.push('## 補足（原文レポート）');
lines.push('- 英語版: `report.md`');
lines.push(`- Scan show JSON: \`scan-show.json\``);

fs.writeFileSync(outputPath, `${lines.join('\n')}\n`, 'utf8');
NODE

cat > "$DEST_DIR/scan-metadata.txt" <<EOF
scanId: ${SCAN_ID}
scanDir: ${SCAN_DIR}
exportedAt: $(date '+%Y-%m-%d %H:%M:%S%z')
EOF

if [[ "$MODE" == "full" ]]; then
  if [[ "$FULL_KEEP_FILES" -eq 1 ]]; then
    # Drop archives, prior keep dirs, and root intermediates from other modes.
    clean_hidden_temp_dirs
    clean_archives
    clean_timestamped_dirs "$(basename "$DEST_DIR")"
    clean_root_intermediates
    rm -f "$OUTPUT_DIR/latest"
    ln -sfn "$(basename "$DEST_DIR")" "$OUTPUT_DIR/latest"

    # fixed-path latest files for easy access
    cp "$DEST_DIR/report.md" "$OUTPUT_DIR/latest-report.md"
    cp "$DEST_DIR/findings.json" "$OUTPUT_DIR/latest-findings.json"
    cp "$DEST_DIR/coverage.json" "$OUTPUT_DIR/latest-coverage.json"
    cp "$DEST_DIR/scan-manifest.json" "$OUTPUT_DIR/latest-scan-manifest.json"
    cp "$DEST_DIR/results.sarif" "$OUTPUT_DIR/latest-results.sarif"
    cp "$DEST_DIR/scan-show.json" "$OUTPUT_DIR/latest-scan-show.json"
    cp "$DEST_DIR/scan-metadata.txt" "$OUTPUT_DIR/latest-scan-metadata.txt"
    cp "$DEST_DIR/report-ja.md" "$OUTPUT_DIR/latest-report-ja.md"
    echo "Exported full codex-security scan ($SCAN_ID) to $DEST_DIR"
    echo "Latest snapshot dir: $OUTPUT_DIR/latest"
    echo "Latest flat files:"
    echo "  - $OUTPUT_DIR/latest-report.md"
    echo "  - $OUTPUT_DIR/latest-findings.json"
    echo "  - $OUTPUT_DIR/latest-report-ja.md"
  else
    # Pure full surface: one timestamped archive + latest symlink only.
    # Preserve active staging dir while dropping orphaned .full-export-* leftovers.
    clean_hidden_temp_dirs "$DEST_DIR"
    clean_timestamped_dirs
    clean_all_latest_flats
    clean_root_intermediates
    clean_archives
    rm -f "$OUTPUT_DIR/latest"

    FULL_ARCHIVE="$OUTPUT_DIR/${TIMESTAMP}-${SCAN_ID}.tar.gz"
    tar -czf "$FULL_ARCHIVE" -C "$DEST_DIR" .
    rm -rf "$DEST_DIR"
    ln -sfn "$(basename "$FULL_ARCHIVE")" "$OUTPUT_DIR/latest"
    echo "Exported full codex-security scan archive ($SCAN_ID) to $FULL_ARCHIVE"
    echo "Latest archive: $OUTPUT_DIR/latest"
  fi
else
  # compact mode: keep only minimal artifacts and overwrite each run
  cp "$DEST_DIR/report.md" "$OUTPUT_DIR/latest-report.md"
  cp "$DEST_DIR/findings.json" "$OUTPUT_DIR/latest-findings.json"
  cp "$DEST_DIR/report-ja.md" "$OUTPUT_DIR/latest-report-ja.md"
  cp "$DEST_DIR/scan-metadata.txt" "$OUTPUT_DIR/latest-scan-metadata.txt"
  # remove verbose artifacts and stale full-mode outputs
  clean_root_intermediates
  clean_full_only_latest_flats
  clean_hidden_temp_dirs
  clean_timestamped_dirs
  clean_archives
  rm -f "$OUTPUT_DIR/latest"
  echo "Exported compact codex-security scan ($SCAN_ID) to $OUTPUT_DIR"
  echo "Compact files:"
  echo "  - $OUTPUT_DIR/latest-report.md"
  echo "  - $OUTPUT_DIR/latest-findings.json"
  echo "  - $OUTPUT_DIR/latest-report-ja.md"
  echo "  - $OUTPUT_DIR/latest-scan-metadata.txt"
fi

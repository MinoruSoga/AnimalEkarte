#!/usr/bin/env python3
"""Static regression guard for research-cloudflare.html.

This is the durable, version-controlled replacement for the throwaway
`scratchpad/verify_html.py` used while authoring the STG-AWS -> Cloudflare
migration research deliverable. A scratchpad script lives in a session-isolated,
ephemeral directory and protects nothing on the next edit; this one ships in the
repo so any future change to the deliverable is re-checked.

It does not run a browser or a network call. It reads the HTML as text and
asserts the load-bearing facts of the deliverable:

- HTML structure is intact (div / table / tr tags balanced).
- All nine document sections are present.
- The executive-summary cost figures and the verdict pill are unchanged.
- The summary carries the HA-parity caveat (the fact that the headline -45%
  production saving narrows once the external Postgres is brought to HA parity),
  and that caveat stays consistent with the detailed cost section.
- Every inventory row cites a source file.
- The FX assumption and official pricing source URLs are present.
- The headline STG-AWS cost reconciles to the Cost-Explorer doc total via the
  stated FX rate, and every cost-summary total row is internally FX-consistent
  (guards against fabricated JPY figures / silent FX drift).
- The "Security Group ×N" inventory count matches the number of
  aws_security_group resources actually declared in the Terraform IaC
  (guards against the undercount type where conditional resources are missed).
- No secrets or infrastructure PII leaked into the deliverable.

Exit code is 0 only when every check passes, non-zero otherwise, so it can be
wired into CI or a pre-edit guard.

Usage:
    python3 scripts/verify_research_html.py path/to/research-cloudflare.html
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import re
import sys


ROOT = Path(__file__).resolve().parents[1]


# IaC + cost truth sources the deliverable's figures must reconcile against.
TERRAFORM_DIR = ROOT / "infra" / "terraform"
COST_DOC = ROOT / "docs" / "infra" / "STG_AWS_COST_REDUCTION.md"

# Section anchors the deliverable must contain (id="..." on the <h2> headings).
REQUIRED_SECTION_IDS = (
    "summary",
    "inventory",
    "mapping",
    "detail",
    "db",
    "cost",
    "verdict",
    "sources",
    "appendix",
)

# Executive-summary cost figures that must never silently drift.
SUMMARY_FIGURES = ("6,424", "1,455", "52,700", "29,100")

# Tags whose open/close counts must balance.
BALANCED_TAGS = ("div", "table", "tr")

# A source citation looks like `foo/bar.go`, `001_init.sql:95`, or `Dockerfile`.
CITATION_RE = re.compile(
    r"(?:[\w./-]+\.(?:go|tf|sql|ts|tsx|js|yml|yaml|md)|Dockerfile)(?::\d+)?"
)

# Official pricing/source hosts that should back the cost claims.
OFFICIAL_SOURCE_RE = re.compile(
    r"https?://[^\"'\s]*"
    r"(?:cloudflare\.com|developers\.cloudflare\.com|neon\.com|neon\.tech|"
    r"supabase\.com|aws\.amazon\.com|amazon\.com/[^\"'\s]*pricing)"
)

# --- Cost reconciliation ---------------------------------------------------
# FX rate as stated in the deliverable ("1 USD = 161.7 JPY").
FX_RE = re.compile(r"1\s*USD\s*=\s*([\d.]+)\s*JPY")
# Cost-doc total row: `| **合計（税前）** | **$21.17** | **$39.73** | ...`.
# Two USD columns (before / after optimization); the second is the live total.
COST_DOC_TOTAL_RE = re.compile(r"合計（税前）.*?\$[\d.]+.*?\$([\d.]+)")
# Each cost-summary total row carries a USD cell then a JPY cell.
TOTALROW_RE = re.compile(r'<tr class="totalrow">(.*?)</tr>', re.DOTALL)
NUM_CELL_RE = re.compile(r'<td class="num">(.*?)</td>')
# Tolerance: total rows reconcile within max(¥10, 0.5%) to absorb per-row
# rounding and the explicitly approximate (~) production projections.
RECONCILE_FLOOR_JPY = 10
RECONCILE_PCT = 0.005

# --- Security Group inventory <-> IaC ---------------------------------------
# A Terraform security-group *declaration* (counted regardless of any `count`
# meta-argument, so conditional `count = ... ? 1 : 0` resources still count 1).
SG_RESOURCE_RE = re.compile(r'resource\s+"aws_security_group"')
# Inventory / mapping text: "Security Group ×3" (× is U+00D7; allow ascii x too).
HTML_SG_COUNT_RE = re.compile(r"Security Group\s*[×x]\s*(\d+)")

# Secrets / infrastructure PII that must NOT appear in a shareable research doc.
SECRET_PATTERNS: tuple[tuple[str, re.Pattern[str]], ...] = (
    ("aws_account_id", re.compile(r"(?<!\d)\d{12}(?!\d)")),
    ("aws_access_key", re.compile(r"AKIA[0-9A-Z]{16}")),
    ("aws_arn", re.compile(r"arn:aws:[a-z0-9-]+:")),
    ("resource_id", re.compile(r"\b(?:subnet|sg|vpc|eni|eip|igw|nat)-[0-9a-f]{8,}\b")),
    ("rds_endpoint", re.compile(r"[a-z0-9-]+\.rds\.amazonaws\.com")),
    ("elb_endpoint", re.compile(r"[a-z0-9-]+\.elb\.amazonaws\.com")),
    # A private host IP is a leak; a CIDR block (x.x.x.x/nn) is documented
    # network design, so the negative lookahead excludes CIDR notation.
    ("private_ip", re.compile(
        r"(?:10\.\d{1,3}\.\d{1,3}\.\d{1,3}"
        r"|172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}"
        r"|192\.168\.\d{1,3}\.\d{1,3})(?!/\d)\b"
    )),
    ("private_key", re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----")),
    ("password_literal", re.compile(r"(?i)password\s*[:=]\s*\S")),
)


@dataclass(frozen=True)
class CheckResult:
    name: str
    passed: bool
    detail: str


def _section(html: str, section_id: str) -> str:
    """Return the slice of html from the given anchor to the next id= anchor."""
    start = html.find(f'id="{section_id}"')
    if start == -1:
        return ""
    nxt = html.find('id="', start + len(section_id) + 5)
    return html[start:] if nxt == -1 else html[start:nxt]


def check_tag_balance(html: str) -> CheckResult:
    parts: list[str] = []
    passed = True
    for tag in BALANCED_TAGS:
        opens = len(re.findall(rf"<{tag}(?:\s|>)", html))
        closes = len(re.findall(rf"</{tag}>", html))
        ok = opens == closes
        passed = passed and ok
        parts.append(f"{tag} {opens}/{closes} {'ok' if ok else 'MISMATCH'}")
    return CheckResult("tag_balance", passed, ", ".join(parts))


def check_sections_present(html: str) -> CheckResult:
    missing = [s for s in REQUIRED_SECTION_IDS if f'id="{s}"' not in html]
    return CheckResult(
        "sections_present",
        not missing,
        "all 9 sections" if not missing else f"missing: {missing}",
    )


def check_summary_figures(html: str) -> CheckResult:
    summary = _section(html, "summary")
    missing = [f for f in SUMMARY_FIGURES if f not in summary]
    return CheckResult(
        "summary_figures_intact",
        not missing,
        "¥6,424/¥1,455/¥52,700/¥29,100 present"
        if not missing
        else f"missing figures: {missing}",
    )


def check_summary_pill(html: str) -> CheckResult:
    summary = _section(html, "summary")
    present = 'pill good">CF が安価' in summary
    return CheckResult(
        "summary_verdict_pill",
        present,
        'pill "CF が安価" present' if present else "verdict pill changed/removed",
    )


def check_summary_ha_caveat(html: str) -> CheckResult:
    """The regression guard the scratchpad SUM block was meant to make durable."""
    summary = _section(html, "summary")
    has_asymmetry = "非対称比較" in summary
    has_range = "150" in summary
    has_parity = "拮抗" in summary
    passed = has_asymmetry and has_range and has_parity
    return CheckResult(
        "summary_ha_caveat",
        passed,
        "summary carries HA-parity caveat (非対称比較 + 150 + 拮抗)"
        if passed
        else f"caveat incomplete: 非対称比較={has_asymmetry} 150={has_range} 拮抗={has_parity}",
    )


def check_cost_section_consistency(html: str) -> CheckResult:
    cost = _section(html, "cost")
    has_range = "150" in cost
    has_parity = "拮抗" in cost
    passed = has_range and has_parity
    return CheckResult(
        "cost_ha_consistency",
        passed,
        "detail cost section matches summary caveat ($150 + 拮抗)"
        if passed
        else f"detail drift: 150={has_range} 拮抗={has_parity}",
    )


def check_fx_assumption(html: str) -> CheckResult:
    present = "161.7" in html
    return CheckResult(
        "fx_assumption",
        present,
        "FX 1USD=161.7JPY stated" if present else "FX assumption missing",
    )


def check_source_urls(html: str) -> CheckResult:
    count = len(set(OFFICIAL_SOURCE_RE.findall(html)))
    # The deliverable cites Cloudflare + AWS + external-Postgres pricing pages.
    passed = count >= 4
    return CheckResult(
        "official_source_urls",
        passed,
        f"{count} official pricing/source hosts cited (need >= 4)",
    )


def check_inventory_citations(html: str) -> CheckResult:
    inventory = _section(html, "inventory")
    tbody_match = re.search(r"<tbody>(.*?)</tbody>", inventory, re.DOTALL)
    if not tbody_match:
        return CheckResult("inventory_citations", False, "no inventory <tbody> found")
    rows = re.findall(r"<tr[^>]*>(.*?)</tr>", tbody_match.group(1), re.DOTALL)
    if not rows:
        return CheckResult("inventory_citations", False, "no inventory rows found")
    uncited = sum(1 for row in rows if not CITATION_RE.search(row))
    passed = uncited == 0
    return CheckResult(
        "inventory_citations",
        passed,
        f"{len(rows)} inventory rows, all cite a source file"
        if passed
        else f"{uncited}/{len(rows)} inventory rows lack a source citation",
    )


def check_no_secrets(html: str) -> CheckResult:
    hits: list[str] = []
    for name, pattern in SECRET_PATTERNS:
        found = pattern.findall(html)
        if found:
            hits.append(f"{name}={found[:2]}")
    passed = not hits
    return CheckResult(
        "no_secrets_or_pii",
        passed,
        "no secrets / infra PII detected" if passed else f"LEAK: {hits}",
    )


def _money_to_float(cell: str) -> float:
    """Strip ~, ¥, $, commas and whitespace from a money cell -> float."""
    return float(re.sub(r"[~¥$,\s]", "", cell))


def check_cost_reconciliation(html: str) -> CheckResult:
    """Headline cost must trace to the Cost-Explorer doc total via the FX rate,
    and every cost-summary total row must be internally FX-consistent.

    This is the regression guard against fabricated JPY figures and silent FX
    drift: a JPY edited without a matching USD/FX change is caught here.
    """
    fx_match = FX_RE.search(html)
    if not fx_match:
        return CheckResult("cost_reconciliation", False, "FX rate not found in HTML")
    fx = float(fx_match.group(1))

    if not COST_DOC.is_file():
        return CheckResult(
            "cost_reconciliation", False, f"cost doc not found: {COST_DOC}"
        )
    doc_match = COST_DOC_TOTAL_RE.search(COST_DOC.read_text(encoding="utf-8"))
    if not doc_match:
        return CheckResult(
            "cost_reconciliation", False, "AWS total USD not parseable from cost doc"
        )
    doc_usd = float(doc_match.group(1))
    expected_stg_jpy = round(doc_usd * fx)

    problems: list[str] = []
    anchor_ok = False
    for row in TOTALROW_RE.findall(html):
        cells = NUM_CELL_RE.findall(row)
        if len(cells) < 2:
            continue
        try:
            usd = _money_to_float(cells[0])
            jpy = _money_to_float(cells[1])
        except ValueError:
            continue
        expected = round(usd * fx)
        tolerance = max(RECONCILE_FLOOR_JPY, jpy * RECONCILE_PCT)
        if abs(expected - jpy) > tolerance:
            problems.append(f"{usd}USD->¥{int(jpy)} (expected ~¥{expected})")
        # Anchor: the headline STG-AWS row must trace to the cost-doc total.
        # Use the same tolerance as the row check so a legitimate rounding
        # difference (e.g. ¥6,421 vs ¥6,424) doesn't raise a false failure.
        if abs(int(jpy) - expected_stg_jpy) <= tolerance:
            anchor_ok = True

    if not anchor_ok:
        problems.append(
            f"no total row equals cost-doc-derived ¥{expected_stg_jpy} "
            f"(${doc_usd}×{fx})"
        )
    passed = not problems
    return CheckResult(
        "cost_reconciliation",
        passed,
        f"cost doc ${doc_usd}×{fx}=¥{expected_stg_jpy} matches headline; "
        f"all total rows reconcile within tolerance"
        if passed
        else f"reconcile FAIL: {problems}",
    )


def check_security_group_count(html: str) -> CheckResult:
    """The "Security Group ×N" inventory count must equal the number of
    aws_security_group resources declared in the Terraform IaC.

    Robust to conditional resources: it counts static `resource
    "aws_security_group"` declarations, not runtime instances, so a
    `count = var.x ? 1 : 0` security group still counts as one. This catches
    the undercount type where a conditional SG (e.g. fck-nat) is overlooked.
    """
    tf_files = [
        f
        for f in TERRAFORM_DIR.rglob("*.tf")
        if ".terraform" not in f.parts
    ]
    if not tf_files:
        return CheckResult(
            "security_group_count",
            False,
            f"no terraform files under {TERRAFORM_DIR}",
        )
    iac_count = sum(
        len(SG_RESOURCE_RE.findall(f.read_text(encoding="utf-8"))) for f in tf_files
    )
    html_counts = [int(n) for n in HTML_SG_COUNT_RE.findall(html)]
    if not html_counts:
        return CheckResult(
            "security_group_count", False, "no 'Security Group ×N' mention in HTML"
        )
    mismatched = [n for n in html_counts if n != iac_count]
    passed = not mismatched
    return CheckResult(
        "security_group_count",
        passed,
        f"HTML SG counts {html_counts} match IaC aws_security_group resources "
        f"({iac_count})"
        if passed
        else f"SG mismatch: HTML says {html_counts}, IaC declares {iac_count}",
    )


CHECKS = (
    check_tag_balance,
    check_sections_present,
    check_summary_figures,
    check_summary_pill,
    check_summary_ha_caveat,
    check_cost_section_consistency,
    check_fx_assumption,
    check_source_urls,
    check_inventory_citations,
    check_cost_reconciliation,
    check_security_group_count,
    check_no_secrets,
)


def run(target: Path) -> int:
    if not target.is_file():
        print(f"FAIL: target not found: {target}", file=sys.stderr)
        return 2
    html = target.read_text(encoding="utf-8")
    results = [check(html) for check in CHECKS]

    # Label relative to ROOT when the target lives inside the repo; fall back to
    # the absolute path for an arbitrary out-of-tree target (the usage docstring
    # allows any path, so relative_to must not raise here).
    try:
        label: Path = target.relative_to(ROOT)
    except ValueError:
        label = target
    print(f"verify_research_html: {label}")
    for r in results:
        status = "PASS" if r.passed else "FAIL"
        print(f"  [{status}] {r.name}: {r.detail}")

    failed = [r for r in results if not r.passed]
    print(f"\n{len(results) - len(failed)}/{len(results)} checks passed")
    if failed:
        print("RESULT: FAIL", file=sys.stderr)
        return 1
    print("RESULT: ALL CHECKS PASS")
    return 0


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print("usage: verify_research_html.py path/to/research-cloudflare.html", file=sys.stderr)
        return 2
    return run(Path(argv[1]).resolve())


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))

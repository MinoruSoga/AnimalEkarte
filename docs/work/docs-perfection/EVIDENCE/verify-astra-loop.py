"""Read-only checks for the authorized investigation/repair loop, not runtime proof."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys
from urllib.parse import unquote

ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "docs/work/docs-perfection"
FINAL = "--final" in sys.argv
BASE = json.loads((OUT / "EVIDENCE/astra-loop-baseline.json").read_text())


def git(*args):
    return subprocess.check_output(["git", "-C", str(ROOT), *args])


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest() if path.is_file() else None


def check(ok, message):
    if not ok:
        raise SystemExit("FAIL " + message)
    print("PASS " + message)


def allowed(path):
    return path.startswith(("docs/architecture/", "docs/spec/", "docs/ops/", "docs/delivery/", "docs/work/docs-perfection/", "docs/work/decisions/")) or path in {
        "docs/README.md", "docs/product-philosophy.md", "docs/work/README.md",
        "docs/work/phase2-deferred.md", "docs/work/linear-f1-f6-mapping.md",
        "docs/work/skill-reeval-2026-09-06.md",
    }


def local_links(path):
    content = re.sub(r"(?ms)^\s*(```|~~~).*?^\s*\1\s*$", "", path.read_text())
    targets = []
    for target in re.findall(r"\[[^\]\n]*\]\(([^)\n]+)\)", content):
        if re.match(r"^[a-zA-Z][a-zA-Z0-9+.-]*:", target) or target.startswith("#"):
            continue
        target = unquote(target.split("#", 1)[0].split(' "', 1)[0].strip("<>"))
        targets.append((target, (path.parent / target).exists()))
    return targets


check(git("rev-parse", "HEAD").decode().strip() == BASE["head"], "HEAD unchanged")
check(hashlib.sha256(git("diff", "--cached", "--binary")).hexdigest() == BASE["index_sha256"], "staging index unchanged")
inventory = (OUT / "INVENTORY.md").read_text()
rows = [line.strip("| ").split(" | ") for line in inventory.splitlines() if line.startswith("| docs/")]
paths = [row[0] for row in rows]
actual = {str(p.relative_to(ROOT)) for p in (ROOT / "docs").rglob("*.md")}
check(set(paths) == actual and len(paths) == len(actual), f"inventory={len(paths)} actual={len(actual)} missing=0 extra=0 duplicate=0")
check(all(len(r) == 5 and all(r) for r in rows), "inventory roles/classifications/evidence present")
missing = [(p, target) for p in sorted(actual) for target, exists in local_links(ROOT / p) if not exists]
check(not missing, "all docs local file links missing=" + str(len(missing)) + " " + str(missing))
for target in ("README.md", "REPAIR-QUEUE.md", "ROLE-MAP.md"):
    check("docs-perfection/" + target in (ROOT / "docs/work/README.md").read_text(), "work entry " + target)

children = {}
for path in sorted((OUT / "CHILD-GOALS").glob("*.md")):
    front = path.read_text().split("---", 2)[1]
    fields = {k: json.loads(v) for k, v in (line.split(": ", 1) for line in front.strip().splitlines())}
    required = {"id", "allowlist", "done_condition", "verification", "acceptance_tests", "forbidden", "depends_on"}
    check(required <= fields.keys() and all(fields[k] for k in required), "child contract " + path.stem)
    if FINAL:
        check(fields["status"] == "COMPLETE", "child docs acceptance " + path.stem)
    children[fields["id"]] = fields
check(len(children) == 6, "six child contracts retained")
seen = {"DOCS-PERFECT-COORDINATOR"}
while set(children) - seen:
    ready = {k for k, v in children.items() if k not in seen and set(v["depends_on"]) <= seen}
    check(bool(ready), "child DAG step")
    seen |= ready

queue = (OUT / "REPAIR-QUEUE.md").read_text()
rq = {r.split(" | ")[0][2:]: r for r in queue.splitlines() if r.startswith("| RQ-")}
check(len(rq) == 11, "eleven RQ entries retained")
for n in (1, 3, 4, 6, 7, 8, 9, 10, 11):
    check(" / resolved |" in rq[f"RQ-{n:03}"], f"RQ-{n:03} docs repair resolved")
check("external-scope" in rq["RQ-002"] and "別" in rq["RQ-002"], "RQ-002 external-scope retained")
check("index-resolved" in rq["RQ-005"], "RQ-005 index-only boundary retained")
check(re.search(r"^status: active$", (OUT / "GOAL.yaml").read_text(), re.M), "parent status active")
report = (ROOT / "docs/ops/testing/UAT-DOMAIN-STATUS.md").read_text()
check(not re.search(r"\]\([^)]*reports/uat-", report), "UAT unavailable report identifiers are not local links")
check("現在の所在" in report and "再確認していない" in report, "UAT original evidence remains unverified")
for name in ("31-lstep-integration.md", "34-lstep-delivery-monitor.md"):
    text = (ROOT / "docs/spec/screens" / name).read_text()
    check("bulk-read を必須" in text and "fallback" in text, name + " normative bulk and observed fallback separated")
manual = "frontend/src/features/manual/content/workflows/01-new-owner-first-visit.md"
check(digest(ROOT / manual) == BASE["tracked_sha256"][manual], "RQ-002 manual source unchanged")

tracked_delta = {p for p, sha in BASE["tracked_sha256"].items() if digest(ROOT / p) != sha}
check(all(allowed(p) for p in tracked_delta if p.startswith("docs/")), "new tracked docs changes within authorized allowlists")
ownership = json.loads((OUT / "EVIDENCE/astra-loop-ownership.json").read_text())
foreign_docs = ownership["foreign_docs_observed_changes"]
check({p for p in tracked_delta if p.startswith("docs/")} == set(ownership["owned_tracked_edits"]) | set(foreign_docs), "tracked docs delta matches owned writes plus separately observed foreign docs")
check(all(digest(ROOT / p) == sha for p, sha in foreign_docs.items()), "foreign docs match observed hashes; no loop overwrite")
check(not any(p.startswith("docs/architecture/adr/") for p in ownership["owned_tracked_edits"]), "owned ADR edits zero (shared-tree ADR changes recorded separately)")
foreign = sorted(p for p in tracked_delta if not p.startswith("docs/"))
print("OBSERVED outside-docs concurrent deltas=" + json.dumps(foreign, ensure_ascii=False))
check({p: digest(ROOT / p) for p in foreign} == ownership["outside_docs_observed_changes"], "outside-docs deltas match separately recorded observation")
print("NOTE outside-docs deltas are not counted as loop edits; inspect owned-write receipt, never restore foreign WIP")
untracked = {p for p in git("ls-files", "--others", "--exclude-standard", "-z").decode().split("\0") if p}
new_untracked = untracked - set(BASE["untracked_sha256"])
check(all(allowed(p) for p in new_untracked if p.startswith("docs/")), "new untracked docs only authorized paths")
print("OBSERVED new outside-docs untracked=" + json.dumps(sorted(p for p in new_untracked if not p.startswith("docs/"))))
check(not any(not p.startswith("docs/") for p in new_untracked), "no unclassified new outside-docs untracked paths")
check(subprocess.run(["git", "-C", str(ROOT), "diff", "--check", "--", "docs"], capture_output=True).returncode == 0, "git diff --check -- docs")
validation = json.loads((OUT / "EVIDENCE/astra-loop-validation.json").read_text())
check(validation["ok"] and validation["harness"]["ok"] and validation["receiver"]["ok"], "saved prompt harness/receiver PASS")
pattern = re.compile(r"AKIA[0-9A-Z]{16}|sk-[a-zA-Z0-9]{24,}|-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----")
changed_docs = {p for p in tracked_delta | untracked if allowed(p) and (ROOT / p).is_file()}
check(not any(pattern.search((ROOT / p).read_text()) for p in changed_docs), "changed docs secret-pattern scan (manual review also required)")
if FINAL:
    for label in ("B", "C"):
        review = json.loads((OUT / f"EVIDENCE/astra-loop-review-{label}.json").read_text())
        check(review["verdict"] == "PASS" and not review["critical_issues"], "independent review " + label)
    drift = (OUT / "EVIDENCE/astra-loop-drift.txt").read_text()
    check("$ bash scripts/check-docs-symbol-drift.sh" in drift and "exit=0" in drift, "terminal drift output/exit recorded")
    ledger = (OUT / "LEDGER.md").read_text()
    check(all(re.search(r"^\| A" + str(i) + r" \|.*\| PASS \|", ledger, re.M) for i in range(1, 9)), "A1-A8 reconciliation PASS")
print("RESULT PASS phase=" + ("final" if FINAL else "draft"))

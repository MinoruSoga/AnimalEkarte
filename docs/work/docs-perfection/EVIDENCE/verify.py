"""Replay the coordinator's docs-only artifact checks; no app or network calls."""
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


def git(*args):
    return subprocess.check_output(["git", "-C", str(ROOT), *args])


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def check(condition, message):
    if not condition:
        raise SystemExit("FAIL " + message)
    print("PASS " + message)


base = json.loads((OUT / "EVIDENCE/BASELINE.json").read_text())
inventory = (OUT / "INVENTORY.md").read_text()
rows = [line.split(" | ") for line in inventory.splitlines() if line.startswith("| docs/")]
paths = [row[0][2:] for row in rows]
actual = {str(p.relative_to(ROOT)) for p in (ROOT / "docs").rglob("*.md")}
check(set(paths) == actual and len(paths) == len(actual), f"inventory={len(paths)} find={len(actual)} missing=0 extra=0 duplicate=0")
enums = {"current", "stale-suspect", "conflict-suspect", "duplicate", "obsolete-suspect", "structure-debt", "skip-with-reason"}
check(all(len(row) == 5 and row[2] in enums and row[3] and row[4] for row in rows), "inventory classification/freshness/reason all rows")
scan = json.loads((OUT / "EVIDENCE/INVENTORY-SCAN.json").read_text())
original = {p for p in base["docs_sha256"] if p.endswith(".md")}
check({r["path"] for r in scan["rows"]} == original, f"baseline Markdown={len(original)} scanned={len(scan['rows'])}")

role = (OUT / "ROLE-MAP.md").read_text()
check(all("## " + r in role for r in ("Constitution", "Map", "Status", "History")), "four role headings")
queue = (OUT / "REPAIR-QUEUE.md").read_text()
items = [line.split(" | ") for line in queue.splitlines() if line.startswith("| RQ-")]
scores = [int(row[3]) for row in items]
check(scores == sorted(scores, reverse=True) and all(int(r[1][1:]) * int(r[2]) == int(r[3]) for r in items), f"repair queue={len(items)} Severity*Impact descending")
check(all(r[5] in {"update", "delete", "merge-link", "rewrite-structure", "adr-correction"} for r in items), "repair actions valid")
queue_ids = {r[0][2:] for r in items}
citations = re.findall(r"`((?:docs|backend|frontend|\.github)/[^`:]+):(\d+)`", queue)
check(bool(citations) and all((ROOT / p).is_file() and 0 < int(n) <= len((ROOT / p).read_text().splitlines()) for p, n in citations), f"repair evidence references={len(citations)} exist with valid lines")

children = {}
required = {"id", "objective", "status", "scope", "allowlist", "done_condition", "verification", "forbidden", "adr_policy", "priority", "depends_on", "repair_queue_refs", "acceptance_tests", "risk_level", "rollback_plan"}
for p in sorted((OUT / "CHILD-GOALS").glob("*.md")):
    front = p.read_text().split("---", 2)[1]
    fields = dict((k, json.loads(v)) for k, v in (line.split(": ", 1) for line in front.strip().splitlines()))
    check(required <= fields.keys() and fields["status"] == "pending" and bool(fields["done_condition"]) and bool(fields["verification"]), "child schema " + p.name)
    check(set(fields["repair_queue_refs"]) <= queue_ids, "child repair refs " + p.name)
    children[fields["id"]] = fields
check({p.stem for p in (OUT / "CHILD-GOALS").glob("*.md")} == {"product-philosophy", "architecture", "spec", "ops", "delivery", "work"}, "six required categories")
seen = {"DOCS-PERFECT-COORDINATOR"}
while set(children) - seen:
    ready = {k for k, v in children.items() if k not in seen and set(v["depends_on"]) <= seen}
    check(bool(ready), "DAG dependency step")
    seen |= ready

goal = (OUT / "GOAL.yaml").read_text()
check(re.search(r"^status: active$", goal, re.M) is not None, "parent status active")
check(all(re.search(r"^" + key + ":", goal, re.M) for key in ("objective", "scope", "done_condition", "budget", "checkpoints", "validation_commands", "forbidden_actions", "approval_required_for", "progress_log_ref")), "goal durable fields")
check("max_steps: 24" in goal and "max_wall_time: 1 session" in goal, "budget limits")

allowed_indexes = {"docs/work/README.md", "docs/ops/deploy/README.md"}
changed = {p for p, old in base["docs_sha256"].items() if not (ROOT / p).is_file() or digest(ROOT / p) != old}
check(changed <= allowed_indexes, "existing docs changed only two allowed index READMEs")
new_docs = {str(p.relative_to(ROOT)) for p in (ROOT / "docs").rglob("*") if p.is_file()} - set(base["docs_sha256"])
check(all(p.startswith("docs/work/docs-perfection/") for p in new_docs), "new docs only coordinator allowlist")
tracked = git("ls-files", "-z").decode().split("\0")
foreign = {p: digest(ROOT / p) for p in tracked if p and not p.startswith("docs/") and (ROOT / p).is_file()}
check(hashlib.sha256(json.dumps(foreign, sort_keys=True).encode()).hexdigest() == base["outside_docs_tracked_digest"], "outside docs tracked content unchanged from baseline")
check(hashlib.sha256(git("diff", "--cached", "--binary")).hexdigest() == base["index_sha256"], "staging index diff unchanged")
new_status = set(git("status", "--short").decode().splitlines()) - set(base["git_status_short"].splitlines())
check(all(line[3:] in allowed_indexes or line[3:].startswith("docs/work/docs-perfection/") for line in new_status), "tracked/untracked status delta within allowlist")
check(subprocess.run(["git", "-C", str(ROOT), "diff", "--check", "--", "docs"], capture_output=True).returncode == 0, "git diff --check -- docs")

missing = []
for p in OUT.rglob("*.md"):
    for target in re.findall(r"\[[^\]\n]*\]\(([^)\n]+)\)", p.read_text()):
        if re.match(r"^[a-zA-Z][a-zA-Z0-9+.-]*:", target) or target.startswith("#"):
            continue
        if not (p.parent / unquote(target.split("#")[0])).exists():
            missing.append((str(p), target))
check(not missing, "coordinator local link targets exist " + str(missing))
secret_pattern = re.compile(r"AKIA[0-9A-Z]{16}|sk-[a-zA-Z0-9]{24,}|-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----")
check(not any(secret_pattern.search(p.read_text()) for p in OUT.rglob("*") if p.is_file()), "new artifact secret-pattern scan (manual PHI review also required)")
validation = json.loads((OUT / "EVIDENCE/delivery-validation.json").read_text())
check(validation["ok"] and validation["harness"]["ok"] and validation["receiver"]["ok"], "saved prompt harness and receiver")
drift = (OUT / "EVIDENCE/docs-symbol-drift.txt").read_text()
check("$ bash scripts/check-docs-symbol-drift.sh" in drift and "exit=" in drift, "drift command output and exit recorded")
if FINAL:
    check("coordinator-complete" in goal and "coordinator-complete" in (OUT / "PROGRESS.md").read_text(), "coordinator-complete checkpoint")
    for label in ("B", "C"):
        review = json.loads((OUT / ("EVIDENCE/review-" + label + ".json")).read_text())
        check(review["verdict"] == "PASS" and not review["critical_issues"], "independent review " + label)
    ledger = (OUT / "LEDGER.md").read_text()
    check(all(re.search(r"^\| AC" + str(i) + r" \|.*\| PASS \|", ledger, re.M) for i in range(1, 10)), "AC1-AC9 completion audit PASS")
print("RESULT PASS phase=" + ("final" if FINAL else "draft"))

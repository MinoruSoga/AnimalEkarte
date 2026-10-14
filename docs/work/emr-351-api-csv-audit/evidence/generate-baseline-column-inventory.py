"""Verify the checked-in inventory against the pinned 21-table contract."""
import csv
import json
from pathlib import Path
import re
import subprocess

BASE = "ed04c0237f81ddc2890d39b528d3608a58180c8b"
SOURCE = "backend/internal/csvimport/cutover_contract.go"
root = Path(subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip())
source = subprocess.check_output(["git", "show", f"{BASE}:{SOURCE}"], cwd=root, text=True)
block = source.split("func CutoverTableSpecs() []CutoverTableSpec {", 1)[1].split("return cloneCutoverSpecs(specs)", 1)[0]
pattern = re.compile(r'^\s*\{"([^"]+)", \[\]string\{([^}]*)\}, \[\]string\{([^}]*)\}, (?:\[\]string\{([^}]*)\}|(nil))\},?$', re.M)
def strings(value):
    return re.findall(r'"([^"]+)"', value or "")
tables = [dict(name=m.group(1), columns=strings(m.group(2)), band_columns=strings(m.group(3)), force_not_null_columns=strings(m.group(4))) for m in pattern.finditer(block)]
pairs = [(t["name"], c) for t in tables for c in t["columns"]]
assert len(tables) == 21 and len(pairs) == 236 and len(set(pairs)) == 236
inventory = json.loads(Path(__file__).with_name("baseline-column-inventory.json").read_text())
assert tables == inventory["tables"]
with Path(__file__).parent.parent.joinpath("column-matrix.csv").open(newline="") as handle:
    matrix = [(r["table"], r["column"]) for r in csv.DictReader(handle)]
assert matrix == pairs
print(json.dumps(tables, ensure_ascii=False, indent=2))

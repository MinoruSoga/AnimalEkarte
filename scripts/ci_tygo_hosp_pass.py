#!/usr/bin/env python3
"""Build a tygo config that only contains the hospitalization-responses package.

tygo may emit only one output per Go package path per run. AnimalEkarte's
tygo.yaml lists internal/medicalrecord twice (medicalrecord-responses +
hospitalization-responses). When the primary generate omits hospitalization,
CI calls this helper then runs tygo again.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path


def main() -> int:
    src_path = Path(sys.argv[1] if len(sys.argv) > 1 else "tygo-ci.yaml")
    out_path = Path(sys.argv[2] if len(sys.argv) > 2 else "tygo-hosp.yaml")
    src = src_path.read_text(encoding="utf-8")
    parts = re.split(r"(?=  - path:)", src)
    if not parts:
        print("empty tygo config", file=sys.stderr)
        return 1
    head, packages = parts[0], parts[1:]
    hosp = [p for p in packages if "hospitalization-responses.ts" in p]
    if not hosp:
        print("hospitalization package not found", file=sys.stderr)
        return 1
    out_path.write_text(head + hosp[0], encoding="utf-8")
    print(f"wrote {out_path} ({len(hosp[0])} chars package body)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

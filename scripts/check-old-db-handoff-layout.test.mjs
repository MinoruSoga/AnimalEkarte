import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";

const stage = fs.readFileSync(new URL("./stage-old-db-handoff.sh", import.meta.url), "utf8");
const importer = fs.readFileSync(
  new URL("./import-old-db-handoffs-on-reset.sh", import.meta.url),
  "utf8",
);
const makefile = fs.readFileSync(new URL("../Makefile", import.meta.url), "utf8");

test("stages CSV under _old_db_handoff/<clinic>/ without a run subdirectory", () => {
  assert.match(
    stage,
    /DEST="\$ROOT\/backend\/migrations\/seeds\/_old_db_handoff\/\$CLINIC_CODE"/,
  );
  assert.doesNotMatch(
    stage,
    /_old_db_handoff\/\$CLINIC_CODE\/\$MIGRATION_RUN_ID/,
  );
});

test("reset import finds clinic/manifest.json rather than clinic/run/manifest.json", () => {
  assert.match(importer, /root\.glob\("\*\/manifest\.json"\)/);
  assert.doesNotMatch(importer, /root\.glob\("\*\/\*\/manifest\.json"\)/);
});

test("make old-db-handoff-check looks for clinic/manifest.json", () => {
  assert.match(
    makefile,
    /_old_db_handoff\/\$\$\{CLINIC_CODE\}\/manifest\.json/,
  );
  assert.doesNotMatch(
    makefile,
    /_old_db_handoff\/\$\$\{CLINIC_CODE\}\/\$\$\{MIGRATION_RUN_ID\}\/manifest\.json/,
  );
});

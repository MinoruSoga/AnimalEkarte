import { expect, test } from "@playwright/test";

import { parseV04Fixture, readV04FixtureFromEnv } from "./v04-fixture";

const valid = {
  clinicId: 991234,
  ownerName: "e2e-owner-991234",
  ownerSearch: "e2e-owner",
  petId: 44,
  petName: "e2e-pet-991234",
  outsideFirstPagePetId: 66,
  outsideFirstPagePetName: "e2e-zebra-991234",
  estimateTitle: "e2e-est-991234",
  medicalRecordCount: 21,
  v04: {
    viewOnlyEmail: "e2e-v04-view-991234@example.test",
    viewOnlyResources: ["master-hospitalization", "lab-import"],
    cageName: "V04-e2e-cage-991234",
    labDeviceName: "V04-e2e-device-991234",
  },
};

function encode(overrides: { clinicId?: unknown; v04?: unknown }): string {
  const { clinicId = valid.clinicId, v04 = valid.v04 } = overrides;
  return JSON.stringify({ ...valid, clinicId, v04 });
}

test.describe("v04 view-only fixture contract", () => {
  test("parses a complete fixture", () => {
    const fixture = parseV04Fixture(JSON.stringify(valid));
    expect(fixture.clinicId).toBe(991234);
    expect(fixture.viewOnlyEmail).toBe("e2e-v04-view-991234@example.test");
    expect(fixture.viewOnlyResources).toEqual(["master-hospitalization", "lab-import"]);
    expect(fixture.cageName).toBe("V04-e2e-cage-991234");
    expect(fixture.labDeviceName).toBe("V04-e2e-device-991234");
  });

  test("rejects missing JSON", () => {
    expect(() => parseV04Fixture(undefined)).toThrow(/missing/);
    expect(() => parseV04Fixture("")).toThrow(/missing/);
  });

  test("rejects reserved clinic ids", () => {
    expect(() => parseV04Fixture(encode({ clinicId: 1 }))).toThrow(/reserved/);
    expect(() => parseV04Fixture(encode({ clinicId: 2 }))).toThrow(/reserved/);
  });

  test("rejects a missing v04 block", () => {
    const { v04: _v04, ...rest } = valid;
    expect(() => parseV04Fixture(JSON.stringify(rest))).toThrow(/v04/);
    expect(() => parseV04Fixture(encode({ v04: null }))).toThrow(/v04/);
  });

  test("rejects a non-example.test view-only email", () => {
    expect(() =>
      parseV04Fixture(
        encode({ v04: { ...valid.v04, viewOnlyEmail: "staff@hospital.example.com" } }),
      ),
    ).toThrow(/example\.test/);
  });

  test("rejects empty or non-string viewOnlyResources", () => {
    expect(() => parseV04Fixture(encode({ v04: { ...valid.v04, viewOnlyResources: [] } }))).toThrow(
      /viewOnlyResources/,
    );
    expect(() =>
      parseV04Fixture(encode({ v04: { ...valid.v04, viewOnlyResources: [42] } })),
    ).toThrow(/viewOnlyResources/);
  });

  test("rejects row names without the V04-e2e- prefix", () => {
    expect(() =>
      parseV04Fixture(encode({ v04: { ...valid.v04, cageName: "cage-991234" } })),
    ).toThrow(/cageName/);
    expect(() => parseV04Fixture(encode({ v04: { ...valid.v04, labDeviceName: "" } }))).toThrow(
      /labDeviceName/,
    );
  });

  test("readV04FixtureFromEnv returns null only when the env is unset or blank", () => {
    const saved = process.env.E2E_CLINICAL_FIXTURE;
    try {
      delete process.env.E2E_CLINICAL_FIXTURE;
      expect(readV04FixtureFromEnv()).toBeNull();
      process.env.E2E_CLINICAL_FIXTURE = "   ";
      expect(readV04FixtureFromEnv()).toBeNull();
      process.env.E2E_CLINICAL_FIXTURE = JSON.stringify(valid);
      expect(readV04FixtureFromEnv()?.viewOnlyEmail).toBe(valid.v04.viewOnlyEmail);
    } finally {
      if (saved === undefined) {
        delete process.env.E2E_CLINICAL_FIXTURE;
      } else {
        process.env.E2E_CLINICAL_FIXTURE = saved;
      }
    }
  });
});

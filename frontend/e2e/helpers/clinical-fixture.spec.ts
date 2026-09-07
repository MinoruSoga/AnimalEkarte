import { expect, test } from "@playwright/test";

import { parseClinicalFixture } from "./clinical-fixture";

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
};

test.describe("clinical e2e fixture contract", () => {
  test("parses a complete fixture and stringifies ids", () => {
    const fixture = parseClinicalFixture(JSON.stringify(valid));
    expect(fixture.clinicId).toBe(991234);
    expect(fixture.petId).toBe("44");
    expect(fixture.outsideFirstPagePet).toEqual({ id: "66", name: "e2e-zebra-991234" });
    expect(fixture.ownerSearch).toBe("e2e-owner");
  });

  test("rejects missing JSON", () => {
    expect(() => parseClinicalFixture(undefined)).toThrow(/missing/);
    expect(() => parseClinicalFixture("")).toThrow(/missing/);
  });

  test("rejects reserved clinic ids", () => {
    expect(() => parseClinicalFixture(JSON.stringify({ ...valid, clinicId: 1 }))).toThrow(
      /reserved/,
    );
    expect(() => parseClinicalFixture(JSON.stringify({ ...valid, clinicId: 2 }))).toThrow(
      /reserved/,
    );
  });

  test("rejects incomplete owner graph", () => {
    expect(() => parseClinicalFixture(JSON.stringify({ ...valid, ownerName: "" }))).toThrow(
      /incomplete/,
    );
    expect(() => parseClinicalFixture(JSON.stringify({ ...valid, medicalRecordCount: 0 }))).toThrow(
      /incomplete|missing/,
    );
  });
});

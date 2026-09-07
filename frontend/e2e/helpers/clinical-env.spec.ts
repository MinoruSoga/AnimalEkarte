import { expect, test } from "@playwright/test";

import {
  assertClinicalAppEnv,
  assertClinicalBaseURL,
  assertClinicalTeardownRegistered,
} from "./clinical-env";

test.describe("clinical e2e env guards", () => {
  test("accepts APP_ENV=test", () => {
    expect(() => assertClinicalAppEnv("test")).not.toThrow();
    expect(() => assertClinicalAppEnv("TEST")).not.toThrow();
  });

  test("rejects non-test APP_ENV", () => {
    expect(() => assertClinicalAppEnv("development")).toThrow(/APP_ENV=test/);
    expect(() => assertClinicalAppEnv("staging")).toThrow(/APP_ENV=test/);
    expect(() => assertClinicalAppEnv(undefined)).toThrow(/APP_ENV=test/);
  });

  test("accepts local compose origins", () => {
    expect(assertClinicalBaseURL("http://localhost:3003")).toBe("http://localhost:3003");
    expect(assertClinicalBaseURL("http://127.0.0.1:3003/login")).toBe("http://127.0.0.1:3003");
    expect(assertClinicalBaseURL("http://host.docker.internal:3003")).toBe(
      "http://host.docker.internal:3003",
    );
    expect(assertClinicalBaseURL(undefined)).toBe("http://localhost:3003");
  });

  test("rejects remote base URLs", () => {
    expect(() => assertClinicalBaseURL("https://stg.example.test")).toThrow(/local compose/);
  });

  test("requires teardown registration", () => {
    expect(() => assertClinicalTeardownRegistered("registered")).not.toThrow();
    expect(() => assertClinicalTeardownRegistered(undefined)).toThrow(/teardown/);
  });
});

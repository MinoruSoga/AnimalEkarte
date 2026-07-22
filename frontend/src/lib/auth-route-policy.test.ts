import { describe, expect, it } from "vitest";

import { paths } from "@/config/paths";

import { isPasswordRecoveryPublicPath } from "./auth-route-policy";

describe("isPasswordRecoveryPublicPath", () => {
  it("derives both recovery routes from the centralized path config", () => {
    expect(paths.auth.forgotPassword.path).toBe("/forgot-password");
    expect(paths.auth.resetPassword.path).toBe("/reset-password");
    expect(isPasswordRecoveryPublicPath(paths.auth.forgotPassword.path)).toBe(true);
    expect(isPasswordRecoveryPublicPath(paths.auth.resetPassword.path)).toBe(true);
  });

  it.each([
    "/forgot-password",
    "/forgot-password/",
    "/reset-password",
    "/reset-password/",
  ])(
    "allows the password recovery path %s with an optional trailing slash",
    (pathname) => {
      expect(isPasswordRecoveryPublicPath(pathname)).toBe(true);
    },
  );

  it("uses pathname only, so a reset token query remains on the public route", () => {
    const url = new URL("/reset-password?token=test-token", "http://localhost");

    expect(isPasswordRecoveryPublicPath(url.pathname)).toBe(true);
  });

  it.each([
    "/login",
    "/forgot-password//",
    "/forgot-password/extra",
    "/reset-password//",
    "/reset-password/extra",
    "//forgot-password",
    "/Forgot-Password",
  ])("rejects non-exact or protected path %s", (pathname) => {
    expect(isPasswordRecoveryPublicPath(pathname)).toBe(false);
  });
});

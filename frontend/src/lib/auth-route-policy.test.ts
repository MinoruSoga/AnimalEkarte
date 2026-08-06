import { describe, expect, it } from "vitest";

import { paths } from "@/config/paths";

import {
  isAuthPublicPath,
  isLoginPublicPath,
  isPasswordRecoveryPublicPath,
} from "./auth-route-policy";

describe("isPasswordRecoveryPublicPath", () => {
  it("derives both recovery routes from the centralized path config", () => {
    expect(paths.auth.forgotPassword.path).toBe("/forgot-password");
    expect(paths.auth.resetPassword.path).toBe("/reset-password");
    expect(isPasswordRecoveryPublicPath(paths.auth.forgotPassword.path)).toBe(true);
    expect(isPasswordRecoveryPublicPath(paths.auth.resetPassword.path)).toBe(true);
  });

  it("treats login and password recovery as public auth routes", () => {
    expect(isAuthPublicPath(paths.auth.login.path)).toBe(true);
    expect(isAuthPublicPath(paths.auth.forgotPassword.path)).toBe(true);
    expect(isAuthPublicPath(paths.auth.resetPassword.path)).toBe(true);
    expect(isAuthPublicPath(paths.home.path)).toBe(false);
  });

  it("identifies login for BUG-031 session hydrate without treating recovery as login", () => {
    expect(isLoginPublicPath(paths.auth.login.path)).toBe(true);
    expect(isLoginPublicPath("/login/")).toBe(true);
    expect(isLoginPublicPath(paths.auth.forgotPassword.path)).toBe(false);
    expect(isLoginPublicPath(paths.home.path)).toBe(false);
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

import { paths } from "@/config/paths";

const PASSWORD_RECOVERY_PUBLIC_PATHS = [
  paths.auth.forgotPassword.path,
  paths.auth.resetPassword.path,
] as const;

const AUTH_PUBLIC_PATHS = [paths.auth.login.path, ...PASSWORD_RECOVERY_PUBLIC_PATHS] as const;

function matchesExactPath(pathname: string, path: string): boolean {
  return pathname === path || pathname === `${path}/`;
}

/** Password recovery routes that must remain usable without session restoration. */
export function isPasswordRecoveryPublicPath(pathname: string): boolean {
  return PASSWORD_RECOVERY_PUBLIC_PATHS.some((path) => matchesExactPath(pathname, path));
}

/** Public auth routes whose expected 401s must not start session refresh. */
export function isAuthPublicPath(pathname: string): boolean {
  return AUTH_PUBLIC_PATHS.some((path) => matchesExactPath(pathname, path));
}

/** Login route only — BUG-031: hydrate session so authenticated users redirect home. */
export function isLoginPublicPath(pathname: string): boolean {
  return matchesExactPath(pathname, paths.auth.login.path);
}

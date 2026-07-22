import { paths } from "@/config/paths";

const PASSWORD_RECOVERY_PUBLIC_PATHS = [
  paths.auth.forgotPassword.path,
  paths.auth.resetPassword.path,
] as const;

/** Password recovery routes that must remain usable without session restoration. */
export function isPasswordRecoveryPublicPath(pathname: string): boolean {
  return PASSWORD_RECOVERY_PUBLIC_PATHS.some(
    (path) => pathname === path || pathname === `${path}/`,
  );
}

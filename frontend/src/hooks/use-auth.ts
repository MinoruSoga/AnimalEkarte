import { useContext } from "react";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";

export type { AuthContextValue };

/**
 * Returns the current auth context value.
 * Must be used within an AuthProvider.
 */
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}

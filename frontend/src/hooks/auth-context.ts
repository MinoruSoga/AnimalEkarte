import { createContext } from "react";
import type { AuthContextValue } from "@/types/auth";

/**
 * AuthContext - shared authentication context.
 * Provided by AuthProvider (features/auth), consumed by useAuth (@/hooks/use-auth).
 */
export const AuthContext = createContext<AuthContextValue | null>(null);

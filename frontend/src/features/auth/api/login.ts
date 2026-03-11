import type { AuthUser } from "../types";
import { MOCK_USERS } from "./mock-data";

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: AuthUser;
}

const SIMULATED_DELAY_MS = 600;

export function login(email: string, password: string): Promise<LoginResponse> {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      const credential = MOCK_USERS.find(
        (c) => c.email === email && c.password === password,
      );
      if (!credential) {
        reject(new Error("メールアドレスまたはパスワードが正しくありません"));
        return;
      }
      resolve({
        accessToken: `mock-access-${credential.user.id}-${Date.now()}`,
        refreshToken: `mock-refresh-${credential.user.id}-${Date.now()}`,
        user: credential.user,
      });
    }, SIMULATED_DELAY_MS);
  });
}

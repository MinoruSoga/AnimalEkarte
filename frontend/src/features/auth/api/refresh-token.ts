import type { AuthUser } from "../types";
import { MOCK_USERS } from "./mock-data";

export interface RefreshResponse {
  accessToken: string;
  user: AuthUser;
}

/* client-localstorage-schema: useAuth.tsx と同じバージョン付きキーを使用 */
const STORAGE_KEY_USER = "auth_user:v1";

export function refreshToken(): Promise<RefreshResponse | null> {
  return new Promise((resolve) => {
    setTimeout(() => {
      /* client-localstorage-schema: getItem を try-catch で保護 (Safari incognito 対策) */
      try {
        const stored = localStorage.getItem(STORAGE_KEY_USER);
        if (!stored) {
          resolve(null);
          return;
        }
        const parsed = JSON.parse(stored) as Partial<AuthUser>;
        if (!parsed?.id) {
          resolve(null);
          return;
        }
        /* 権限情報はモックDBから再取得して最新状態を保証 */
        const matched = MOCK_USERS.find((m) => m.user.id === parsed.id);
        const user = matched ? matched.user : (parsed as AuthUser);
        resolve({
          accessToken: `mock-access-${user.id}-${Date.now()}`,
          user,
        });
      } catch {
        resolve(null);
      }
    }, 300);
  });
}

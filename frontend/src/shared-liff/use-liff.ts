import { useState, useEffect } from "react";
import liff from "@line/liff";

// liff/ と line-reserve/ の両アプリで共有する LIFF 初期化フック（R-F21）。
// LIFF_MOCK は各アプリの lib/liff-config.ts と同じロジック（VITE_LIFF_MOCK 環境変数）で
// 直接判定し、アプリ固有の設定モジュールへの依存を避ける。
const LIFF_MOCK = import.meta.env.VITE_LIFF_MOCK === "true";

export interface UseLiffReturn {
  idToken: string | null;
  displayName: string;
  pictureUrl: string | null;
  isReady: boolean;
  initError: boolean;
}

export function useLiff(liffId: string): UseLiffReturn {
  const [idToken, setIdToken] = useState<string | null>(() => (LIFF_MOCK ? "mock-token" : null));
  const [displayName, setDisplayName] = useState<string>(() => (LIFF_MOCK ? "テストユーザー" : ""));
  const [pictureUrl, setPictureUrl] = useState<string | null>(null);
  const [isReady, setIsReady] = useState<boolean>(() => LIFF_MOCK);
  const [initError, setInitError] = useState<boolean>(false);

  useEffect(() => {
    if (!liffId) return;
    if (LIFF_MOCK) return;

    liff
      .init({ liffId })
      .then(async () => {
        if (!liff.isLoggedIn()) {
          liff.login();
          return;
        }
        const token = liff.getIDToken();
        setIdToken(token);

        try {
          const profile = await liff.getProfile();
          setDisplayName(profile.displayName);
          setPictureUrl(profile.pictureUrl ?? null);
        } catch {
          // プロフィール取得失敗は無視
        }

        setIsReady(true);
      })
      .catch((err: unknown) => {
        if (import.meta.env.DEV) {
          console.error("[useLiff] init failed:", err);
        }
        setInitError(true);
        setIsReady(true); // ローディングを解除して ErrorPage へ遷移させる
      });
  }, [liffId]);

  return { idToken, displayName, pictureUrl, isReady, initError };
}

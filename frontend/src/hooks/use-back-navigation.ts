import { useCallback } from "react";
import { useLocation, useNavigate } from "react-router";
import { parseInternalPath } from "@/lib/internal-navigation";

interface BackNavigate {
  (to: string, options?: { state?: unknown }): void;
}

function listReturnState(locationState: unknown): unknown {
  if (locationState === null || typeof locationState !== "object") return undefined;
  if (!("listReturnState" in locationState)) return undefined;
  return (locationState as { listReturnState?: unknown }).listReturnState;
}

/** 詳細の戻る・保存が、一覧 URL と一覧が持っていた戻り先を復元する。 */
export function performBackNavigation(
  navigate: BackNavigate,
  fallbackPath: string,
  locationState: unknown,
): void {
  const from = (locationState as { from?: unknown } | null)?.from;
  const to = parseInternalPath(from) ?? fallbackPath;
  const restored = listReturnState(locationState);
  if (restored !== undefined) {
    navigate(to, { state: restored });
    return;
  }
  navigate(to);
}

/**
 * 画面遷移元が `location.state.from` に渡した復帰先パスを返す。
 * `from` が無い・内部パスでない場合は `fallbackPath`（通常は機能の一覧ページ）を返す。
 * 保存後リダイレクトなど「戻るボタン以外の復帰」にも使う。
 *
 * `from` は `parseInternalPath` で same-origin の内部パスに限定する
 * （open redirect 防止。LoginForm / AuthProvider と同じ検証）。
 */
export function useBackPath(fallbackPath: string): string {
  const location = useLocation();
  const from = (location.state as { from?: unknown } | null)?.from;
  return parseInternalPath(from) ?? fallbackPath;
}

/**
 * 画面遷移元が `location.state.from` に渡した復帰先パスへ戻る onBack ハンドラを返す。
 * `from` が無い・内部パスでない場合は `fallbackPath`（通常は機能の一覧ページ）へ戻る。
 */
export function useBackNavigation(fallbackPath: string): () => void {
  const navigate = useNavigate();
  const location = useLocation();

  return useCallback(() => {
    performBackNavigation(navigate, fallbackPath, location.state);
  }, [navigate, fallbackPath, location.state]);
}

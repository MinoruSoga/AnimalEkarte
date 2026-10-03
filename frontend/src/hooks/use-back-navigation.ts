import { useCallback } from "react";
import { useLocation, useNavigate } from "react-router";
import { parseInternalPath } from "@/lib/internal-navigation";

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
  const backPath = useBackPath(fallbackPath);

  return useCallback(() => {
    navigate(backPath);
  }, [navigate, backPath]);
}

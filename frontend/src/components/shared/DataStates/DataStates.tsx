import { C } from "@/lib/design-tokens";
import type { ReactNode } from "react";

/**
 * ローディング・エラー・空状態の共有フォールバックコンポーネント。
 * 各リストページで重複していたインライン実装を統合。
 */

/** スピナーを表示するローディングフォールバック */
export function LoadingFallback({ className }: { className?: string }) {
  return (
    <div className={`flex justify-center items-center p-8 ${className ?? ""}`}>
      <div className={`inline-block animate-spin rounded-full h-8 w-8 border-b-2 ${C.borderPrimary}`} />
    </div>
  );
}

/** エラーメッセージを表示するエラーフォールバック */
export function ErrorFallback({
  message = "データの取得に失敗しました",
  className,
}: {
  message?: string;
  className?: string;
}) {
  return (
    <div className={`p-4 ${C.danger} ${className ?? ""}`}>
      {message}
    </div>
  );
}

/** アイコン + テキストを表示する空状態フォールバック */
export function EmptyStateFallback({
  icon,
  message,
  className,
}: {
  icon?: ReactNode;
  message: string;
  className?: string;
}) {
  return (
    <div className={`flex flex-col items-center gap-2 p-8 ${C.text50} ${className ?? ""}`}>
      {icon ? <div className="opacity-40">{icon}</div> : null}
      <span className="text-sm">{message}</span>
    </div>
  );
}

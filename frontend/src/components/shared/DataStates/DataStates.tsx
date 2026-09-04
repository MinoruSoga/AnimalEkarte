import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { C } from "@/lib/design-tokens";

/**
 * ローディング・エラー・空状態の共有フォールバックコンポーネント。
 * 各リストページで重複していたインライン実装を統合。
 */

/** スピナーを表示するローディングフォールバック */
export function LoadingFallback({ className }: { className?: string }) {
  return (
    <div className={`flex justify-center items-center p-8 ${className ?? ""}`}>
      <div
        className={`inline-block animate-spin rounded-full h-8 w-8 border-b-2 ${C.borderPrimary}`}
      />
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
  return <div className={`p-4 ${C.danger} ${className ?? ""}`}>{message}</div>;
}

/**
 * ページ/セクション単位の空状態カード（DESIGN.md ex-empty-state-card 準拠:
 * canvas-soft 背景 + rounded-xl + 余裕のあるパディング）。
 * テーブル内の空行には代わりに STYLE.tableEmpty を使うこと（カード内カードを避ける）。
 *
 * `children` はメッセージ下のアクションスロット（例: 「記録を追加」ボタン）。
 * `data-testid` 等の残り div 属性は型安全に pass-through される。
 */
interface EmptyStateProps extends ComponentPropsWithoutRef<"div"> {
  icon?: ReactNode;
  message: string;
  description?: string;
}

export function EmptyState({
  icon,
  message,
  description,
  className,
  children,
  ...rest
}: EmptyStateProps) {
  return (
    <div
      {...rest}
      className={`flex flex-col items-center justify-center gap-2 rounded-xl ${C.bgPage} py-8 px-8 text-center ${className ?? ""}`}
    >
      {icon ? <div className={C.text30}>{icon}</div> : null}
      <p className={`text-base ${C.text60}`}>{message}</p>
      {description ? <p className={`text-sm ${C.text40}`}>{description}</p> : null}
      {children}
    </div>
  );
}

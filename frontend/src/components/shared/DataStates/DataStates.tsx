import { C } from "@/lib/design-tokens";

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

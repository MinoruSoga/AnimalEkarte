import type { ReactNode } from "react";

interface FilteringIndicatorProps {
  /** useDeferredValue による遅延中かどうか（true のとき opacity を落とす） */
  isFiltering: boolean;
  children: ReactNode;
  className?: string;
}

/**
 * `useDeferredValue` による検索フィルタリング中に opacity トランジションを適用する
 * 共有ラッパーコンポーネント。各リストページのインライン実装を統一する。
 */
export function FilteringIndicator({ isFiltering, children, className }: FilteringIndicatorProps) {
  return (
    <div
      className={`transition-opacity duration-150 ${isFiltering ? "opacity-60" : ""} ${className ?? ""}`.trimEnd()}
    >
      {children}
    </div>
  );
}

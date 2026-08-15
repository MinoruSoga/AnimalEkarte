import { useEffect, useState } from "react";

/**
 * 値の変化を delayMs だけ遅らせて返す。
 *
 * 用途は「入力のたびにサーバへ問い合わせない」こと。検索条件を直接クエリへ渡すと
 * 1文字ごとにリクエストが飛ぶため、入力が止まってから1回だけ反映させる。
 *
 * 連続変更ではタイマーを張り直すので、中間値は反映されず最後の値だけが通る。
 *
 * `value` は呼び出し側で参照が安定していること（`useState` / `useMemo` 由来）。
 * 毎レンダー新規生成されるオブジェクトや配列を直接渡すと、依存が毎回変化して
 * タイマーが張り直され続け、値が永久に確定しない。
 */
export function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const timerId = setTimeout(() => setDebouncedValue(value), delayMs);
    return () => clearTimeout(timerId);
  }, [value, delayMs]);

  return debouncedValue;
}

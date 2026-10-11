import { useEffect, useState } from "react";

import { todayJSTISO } from "@/lib/jst-date";

/** 暦日切替えの検知間隔。same-value setState は React が bailout するため毎分 poll で十分。 */
const TODAY_JST_TICK_MS = 60_000;

/**
 * Codex PR606 #13: マウント中に JST 暦日を跨いでも「今日」を更新し続ける hook。
 * 期限超過 tentative の判定は日単位（JST 暦日 < 今日）なので、
 * 一度 render された境界が古いまま残ると expired 行が画面から消える。
 * 分単位の interval で todayJSTISO() を再評価する。
 */
export function useTodayJst(): string {
  const [today, setToday] = useState(() => todayJSTISO());
  useEffect(() => {
    const id = setInterval(() => setToday(todayJSTISO()), TODAY_JST_TICK_MS);
    return () => clearInterval(id);
  }, []);
  return today;
}

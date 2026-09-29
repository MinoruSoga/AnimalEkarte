/**
 * HelpChatHistoryProvider — ヘルプチャットの会話履歴を保持する Provider
 *
 * SupportWidget（Layout 常駐）から使い、パネルの開閉・タブ切替・記事遷移で
 * HelpChat がアンマウントされても履歴を残す。Layout アンマウント（ログアウト等）
 * で自然にリセットされる。
 *
 * hydrated はサーバー履歴をローカルへ一度だけ反映したかを記録し、
 * 再マウント時の二重適用を防ぐ。
 */
import { useCallback, useMemo, useState, type ReactNode } from "react";

import { HelpChatHistoryContext, type HelpChatHistoryValue } from "../lib/help-chat-history";
import type { SupportChatTurn } from "../types";

export function HelpChatHistoryProvider({ children }: { children: ReactNode }) {
  const [turns, setTurns] = useState<SupportChatTurn[]>([]);
  const [hydrated, setHydrated] = useState(false);
  const markHydrated = useCallback(() => setHydrated(true), []);
  const value = useMemo<HelpChatHistoryValue>(
    () => ({ turns, setTurns, hydrated, markHydrated }),
    [turns, hydrated, markHydrated],
  );
  return (
    <HelpChatHistoryContext.Provider value={value}>{children}</HelpChatHistoryContext.Provider>
  );
}

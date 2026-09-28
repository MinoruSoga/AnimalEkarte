/**
 * help-chat-history.ts — ヘルプチャットの会話履歴コンテキストとフック
 *
 * SupportWidget（Layout 常駐）配下に Provider を置くことで、パネルの開閉・
 * タブ切替・マニュアル記事への遷移で HelpChat がアンマウントされても
 * 会話が残る。Layout のアンマウント（ログアウト等）で自然にリセットされる。
 *
 * hydrated は「サーバー保存済み履歴をローカルへ一度だけ反映したか」の旗で、
 * HelpChat の再マウント時に古いキャッシュでローカル履歴を上書きしないためのもの。
 *
 * Provider 外で呼ばれた場合は useState にフォールバックし、
 * 従来どおりコンポーネント内完結の履歴として動く（テスト容易性のため）。
 */
import {
  createContext,
  useContext,
  useMemo,
  useState,
  type Dispatch,
  type SetStateAction,
} from "react";

import type { SupportChatTurn } from "../types";

export interface HelpChatHistoryValue {
  turns: SupportChatTurn[];
  setTurns: Dispatch<SetStateAction<SupportChatTurn[]>>;
  /** サーバー履歴の初回反映が済んだか（済んでいれば再適用しない） */
  hydrated: boolean;
  markHydrated: () => void;
}

export const HelpChatHistoryContext = createContext<HelpChatHistoryValue | null>(null);

export function useHelpChatHistory(): HelpChatHistoryValue {
  const context = useContext(HelpChatHistoryContext);
  const [turns, setTurns] = useState<SupportChatTurn[]>([]);
  const [hydrated, setHydrated] = useState(false);
  const fallback = useMemo<HelpChatHistoryValue>(
    () => ({ turns, setTurns, hydrated, markHydrated: () => setHydrated(true) }),
    [turns, hydrated],
  );
  return context ?? fallback;
}

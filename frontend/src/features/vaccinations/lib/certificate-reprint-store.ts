import type { CertificateReprintRequest } from "../api/certificate-api";
import type { IssueNoticeCode } from "./certificate-issue-reducer";

/**
 * 再印刷の operation_key と進行状態を、履歴行（CertificateEditionRow）より上で保持する feature 内 store。
 * 履歴の様式タブ・ページ送り・context 再取得で行が unmount されても、未確定の key を失わない
 * （失うと新しい key で二重に再印刷しうる）。edition id ごとに 1 エントリ。
 * 画面（CertificateHistory）が所有し、接種記録が変わったら一緒に破棄される。永続化はしない。
 */

export type ReprintState =
  | { phase: "idle" }
  | { phase: "running" }
  | { phase: "unknown"; notice: IssueNoticeCode }
  | { phase: "not-issued"; notice: IssueNoticeCode }
  | { phase: "stale"; notice: IssueNoticeCode }
  | { phase: "blocked"; notice: IssueNoticeCode }
  | { phase: "done"; editionId: string; artifactId: string; replayed: boolean };

export interface ReprintPending {
  readonly editionId: string;
  readonly request: CertificateReprintRequest;
}

export interface ReprintEntry {
  readonly state: ReprintState;
  /** 結果が確定するまで保持する操作。null なら新しい操作を始められる。 */
  readonly pending: ReprintPending | null;
}

export interface ReprintStore {
  get(editionId: number): ReprintEntry;
  set(editionId: number, entry: ReprintEntry): void;
  subscribe(listener: () => void): () => void;
}

const IDLE_ENTRY: ReprintEntry = { state: { phase: "idle" }, pending: null };

export function createReprintStore(): ReprintStore {
  const entries = new Map<number, ReprintEntry>();
  const listeners = new Set<() => void>();
  return {
    get: (editionId) => entries.get(editionId) ?? IDLE_ENTRY,
    set(editionId, entry) {
      entries.set(editionId, entry);
      for (const listener of [...listeners]) listener();
    },
    subscribe(listener) {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
  };
}

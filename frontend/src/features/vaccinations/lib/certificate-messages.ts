import type { IssueNoticeCode } from "./certificate-issue-reducer";

/** 画面に出す文言。機械判定は code / status で行い、BE の本文（ローカライズ可）は表示に使わない。 */
export const ISSUE_NOTICE: Record<
  IssueNoticeCode,
  { tone: "error" | "warning" | "info"; title: string; detail: string }
> = {
  "dependency-changed": {
    tone: "warning",
    title: "証明書の元情報が変わりました",
    detail:
      "接種の訂正・削除、発行日の日跨ぎなどで確認済みの内容が古くなりました。最新の内容を取得し、もう一度プレビューで確認してください。",
  },
  "version-conflict": {
    tone: "warning",
    title: "他の端末で更新されました",
    detail: "接種または発行履歴が更新されています。最新の状態を取得したので、プレビューからやり直してください。",
  },
  "operation-key-conflict": {
    tone: "error",
    title: "同じ操作番号で内容が異なります",
    detail: "古い送信と内容が一致しませんでした。結果は表示しません。もう一度プレビューから発行してください。",
  },
  validation: {
    tone: "error",
    title: "入力内容を受け付けられませんでした",
    detail: "様式の選択または入力欄を確認し、もう一度プレビューしてください。",
  },
  forbidden: {
    tone: "error",
    title: "この操作を行う権限がありません",
    detail: "権限が変更された可能性があります。管理者に確認してください。",
  },
  unauthenticated: {
    tone: "error",
    title: "ログインの有効期限が切れました",
    detail: "再ログイン後にやり直してください。",
  },
  "not-found": {
    tone: "error",
    title: "対象が見つかりません",
    detail: "接種記録が削除された可能性があります。一覧から開き直してください。",
  },
  "preview-failed": {
    tone: "error",
    title: "プレビューを作成できませんでした",
    detail: "通信状態を確認して、もう一度お試しください。",
  },
  "render-unavailable": {
    tone: "warning",
    title: "証明書の作成が完了したか確認できません",
    detail: "発行済みの可能性があります。結果を照会してください。新しく発行し直さないでください。",
  },
  busy: {
    tone: "info",
    title: "同じ発行の処理中です",
    detail: "処理が終わるまで待ってから、結果を照会してください。別の発行は始めないでください。",
  },
  unavailable: {
    tone: "warning",
    title: "発行結果を確認できません",
    detail: "発行済みの可能性があります。結果を照会してください。新しく発行し直さないでください。",
  },
  "not-succeeded": {
    tone: "warning",
    title: "この操作では発行されていません",
    detail: "同じ内容で再試行するか、内容を直してプレビューからやり直せます。",
  },
  "not-claimed": {
    tone: "warning",
    title: "この操作は受け付けられていません",
    detail: "発行されていません。同じ内容で再試行するか、内容を直してプレビューからやり直せます。",
  },
  malformed: {
    tone: "warning",
    title: "応答の形式が想定と異なります",
    detail: "発行済みの可能性があります。結果を照会してください。",
  },
  "contract-violation": {
    tone: "warning",
    title: "発行結果の参照が不足しています",
    detail: "発行済みの可能性があります。結果を照会してください。",
  },
};

export const VALIDITY_LABEL = {
  current: "現在有効",
  superseded: "訂正版あり（旧版）",
  source_changed: "接種が訂正済み（旧版）",
  source_deleted: "接種が削除済み（旧版）",
} as const;

export const BODY_SIZE_LABEL = { small: "小", medium: "中", large: "大" } as const;

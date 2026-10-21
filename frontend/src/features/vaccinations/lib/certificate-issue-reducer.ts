import type { CertificateIssueRequest } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import type {
  CertificateKind,
  CertificateLayout,
  CertificatePreviewResponse,
} from "../api/certificate-schemas";
import { emptyDraft, type CertificateInputDraft } from "./certificate-inputs";

/**
 * EMR-487: 発行フローの純粋な状態遷移。
 *
 * operation_key の規則（契約 C06 節）:
 * - key は「確認済み preview」1 つに束縛して発行時に 1 度だけ作る。
 * - 結果不明（通信断・503・operation-in-progress・サーバ異常）では同じ key・同じ request で照会する。
 *   新しい key を作るのは再 preview を経たときだけ。
 * - 入力・kind・layout のどれかが変わったら preview / fingerprint / key を破棄し、再 preview を要求する。
 * - 結果不明の間は編集不可（旧結果を確定してから別 key へ進む）。
 */

export type IssuePhase =
  | "editing" // 入力中。preview 未取得または無効化済み
  | "previewing"
  | "previewed" // 確認 preview あり。発行可
  | "issuing"
  | "unknown" // 結果不明。同 key で照会する
  | "not-issued" // 照会で「成功なし」が確定。同 key 同内容の再試行か、修正して再 preview
  | "issued"
  | "stale" // 前提変更。再取得・再 preview が必要
  | "blocked"; // 権限なし・未認証。操作不可

export type IssueNoticeCode =
  | "dependency-changed"
  | "version-conflict"
  | "operation-key-conflict"
  | "validation"
  | "forbidden"
  | "unauthenticated"
  | "not-found"
  | "preview-failed"
  | "render-unavailable"
  | "busy"
  | "unavailable"
  | "not-succeeded"
  | "not-claimed"
  | "malformed"
  | "contract-violation";

export interface IssuePending {
  readonly key: string;
  readonly request: CertificateIssueRequest;
}

export interface IssueResultRefs {
  readonly editionId: string;
  readonly artifactId: string;
  readonly replayed: boolean;
}

export interface IssueState {
  readonly phase: IssuePhase;
  readonly kind: CertificateKind;
  readonly layout: CertificateLayout;
  readonly draft: CertificateInputDraft;
  readonly preview: CertificatePreviewResponse | null;
  /** preview を取得した request（発行 body の元。inputs の再計算を避ける） */
  readonly previewRequest: Omit<CertificateIssueRequest, "operation_key" | "preview_fingerprint" | "preview_display_issue_date"> | null;
  readonly pending: IssuePending | null;
  readonly issued: IssueResultRefs | null;
  readonly notice: IssueNoticeCode | null;
}

export function initialIssueState(init: {
  kind: CertificateKind;
  layout: CertificateLayout;
  draft?: CertificateInputDraft;
}): IssueState {
  return {
    phase: "editing",
    kind: init.kind,
    layout: init.layout,
    draft: init.draft ?? emptyDraft(),
    preview: null,
    previewRequest: null,
    pending: null,
    issued: null,
    notice: null,
  };
}

export type IssueAction =
  | { type: "edit"; patch: { kind?: CertificateKind; layout?: CertificateLayout; draft?: Partial<CertificateInputDraft> } }
  | { type: "replaceDraft"; draft: CertificateInputDraft }
  | { type: "previewStart" }
  | {
      type: "previewOk";
      preview: CertificatePreviewResponse;
      request: NonNullable<IssueState["previewRequest"]>;
    }
  | { type: "previewFail"; notice: IssueNoticeCode }
  | { type: "issueStart"; pending: IssuePending }
  | { type: "issueOk"; refs: IssueResultRefs }
  | { type: "issueUnknown"; notice: IssueNoticeCode }
  | { type: "lookupNotSucceeded"; notice: "not-succeeded" | "not-claimed" }
  | { type: "retryStart" }
  | { type: "stale"; notice: IssueNoticeCode }
  | { type: "blocked"; notice: IssueNoticeCode }
  | { type: "unblock" }
  | { type: "acknowledgeResolved" }
  | { type: "startNew" };

const LOCKED: readonly IssuePhase[] = ["issuing", "unknown", "issued", "blocked"];

/** 編集可能か。結果不明・発行中・発行済み・権限なしでは触れない。 */
export function isEditable(phase: IssuePhase): boolean {
  return !LOCKED.includes(phase);
}

export function issueReducer(state: IssueState, action: IssueAction): IssueState {
  switch (action.type) {
    case "edit": {
      if (!isEditable(state.phase)) return state;
      const kind = action.patch.kind ?? state.kind;
      const kindChanged = kind !== state.kind;
      const layout = action.patch.layout ?? state.layout;
      // kind 切替で rabies 専用値を持ち越さない（normal は非 null で 422）。
      const baseDraft = kindChanged ? emptyDraft() : state.draft;
      return {
        ...state,
        phase: "editing",
        kind,
        layout,
        draft: { ...baseDraft, ...action.patch.draft },
        preview: null,
        previewRequest: null,
        pending: null,
        notice: null,
      };
    }
    case "replaceDraft":
      if (!isEditable(state.phase)) return state;
      return { ...state, phase: "editing", draft: action.draft, preview: null, previewRequest: null, pending: null, notice: null };
    case "previewStart":
      return { ...state, phase: "previewing", preview: null, previewRequest: null, pending: null, notice: null };
    case "previewOk":
      return { ...state, phase: "previewed", preview: action.preview, previewRequest: action.request, pending: null, notice: null };
    case "previewFail":
      return { ...state, phase: "editing", preview: null, previewRequest: null, pending: null, notice: action.notice };
    case "issueStart":
      return { ...state, phase: "issuing", pending: action.pending, notice: null };
    case "issueOk":
      return { ...state, phase: "issued", issued: action.refs, notice: null };
    case "issueUnknown":
      return { ...state, phase: "unknown", notice: action.notice };
    case "lookupNotSucceeded":
      return { ...state, phase: "not-issued", notice: action.notice };
    case "retryStart":
      return state.pending === null ? state : { ...state, phase: "issuing", notice: null };
    case "stale":
      // 前提変更: 旧 preview / key を破棄。次の発行は再 preview 後の新 key。
      return { ...state, phase: "stale", preview: null, previewRequest: null, pending: null, notice: action.notice };
    case "blocked":
      return { ...state, phase: "blocked", notice: action.notice };
    case "unblock":
      // 権限が戻ったことを呼び出し側が確認した後の解除。未確定の key があれば捨てずに結果不明へ戻して同 key で照会し、
      // なければ preview を作り直す編集状態へ戻る（確認済みだった preview は引き継がない）。
      if (state.phase !== "blocked") return state;
      return state.pending !== null
        ? { ...state, phase: "unknown", notice: "unavailable" }
        : { ...state, phase: "editing", preview: null, previewRequest: null, notice: null };
    case "startNew":
      // 発行済みから新しい発行へ。preview / pending を残さない（新 key は次の preview の後）。
      return state.phase === "issued"
        ? initialIssueState({ kind: state.kind, layout: state.layout })
        : state;
    case "acknowledgeResolved":
      return { ...state, phase: "editing", preview: null, previewRequest: null, pending: null, notice: null };
  }
}

export type IssueFailureDecision =
  | { type: "unknown"; notice: IssueNoticeCode }
  | { type: "stale"; notice: IssueNoticeCode; refetchContext: boolean }
  | { type: "blocked"; notice: IssueNoticeCode };

/** POST /certificate-editions の失敗を次の遷移へ写す。 */
export function decideIssueFailure(error: unknown): IssueFailureDecision {
  if (!(error instanceof CertificateApiError)) return { type: "unknown", notice: "unavailable" };
  switch (error.kind) {
    case "forbidden":
      return { type: "blocked", notice: "forbidden" };
    case "unauthenticated":
      return { type: "blocked", notice: "unauthenticated" };
    case "dependency-changed":
      return { type: "stale", notice: "dependency-changed", refetchContext: true };
    case "version-conflict":
      return { type: "stale", notice: "version-conflict", refetchContext: true };
    case "operation-key-conflict":
      return { type: "stale", notice: "operation-key-conflict", refetchContext: false };
    case "validation":
      return { type: "stale", notice: "validation", refetchContext: false };
    case "not-found":
      return { type: "stale", notice: "not-found", refetchContext: true };
    case "operation-in-progress":
      return { type: "unknown", notice: "busy" };
    case "render-unavailable":
      return { type: "unknown", notice: "render-unavailable" };
    case "malformed":
      return { type: "unknown", notice: "malformed" };
    default:
      return { type: "unknown", notice: "unavailable" };
  }
}

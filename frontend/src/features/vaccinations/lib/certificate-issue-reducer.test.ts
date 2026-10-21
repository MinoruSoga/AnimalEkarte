import { describe, expect, it } from "vitest";
import { CertificateApiError } from "../api/certificate-errors";
import type { CertificatePreviewResponse } from "../api/certificate-schemas";
import { emptyDraft } from "./certificate-inputs";
import {
  decideIssueFailure,
  initialIssueState,
  isEditable,
  issueReducer,
  restoredIssueState,
  type IssuePending,
} from "./certificate-issue-reducer";

const preview = { preview_fingerprint: "a".repeat(64) } as CertificatePreviewResponse;
const request = {
  kind: "rabies" as const,
  layout: "rabies" as const,
  inputs: {},
  preconditions: { source_version: 7, latest_edition_id: null },
};
const pending: IssuePending = {
  key: "11111111-1111-4111-8111-111111111111",
  request: { ...request, operation_key: "11111111-1111-4111-8111-111111111111", preview_fingerprint: "a".repeat(64), preview_display_issue_date: "2026-10-20" },
};

function previewed() {
  const s = initialIssueState({ kind: "rabies", layout: "rabies" });
  return issueReducer(s, { type: "previewOk", preview, request });
}

describe("発行フローの状態遷移", () => {
  it("入力・kind・layout の変更で preview と key を破棄し再 preview を要求する", () => {
    const edited = issueReducer(previewed(), { type: "edit", patch: { draft: { total: "100" } } });
    expect(edited.phase).toBe("editing");
    expect(edited.preview).toBeNull();
    expect(edited.pending).toBeNull();
    const layoutChanged = issueReducer(
      issueReducer(initialIssueState({ kind: "normal", layout: "portrait" }), { type: "previewOk", preview, request }),
      { type: "edit", patch: { layout: "landscape" } },
    );
    expect(layoutChanged.preview).toBeNull();
  });

  it("kind を切り替えると rabies 専用の入力値を持ち越さない", () => {
    const s = issueReducer(initialIssueState({ kind: "rabies", layout: "rabies", draft: { ...emptyDraft(), total: "500" } }), {
      type: "edit",
      patch: { kind: "normal", layout: "portrait" },
    });
    expect(s.draft).toEqual(emptyDraft());
  });

  it("結果不明の間は編集できず、key を保持する", () => {
    let s = issueReducer(previewed(), { type: "issueStart", pending });
    s = issueReducer(s, { type: "issueUnknown", notice: "unavailable" });
    expect(isEditable(s.phase)).toBe(false);
    const attempt = issueReducer(s, { type: "edit", patch: { draft: { total: "1" } } });
    expect(attempt).toBe(s);
    expect(attempt.pending?.key).toBe(pending.key);
  });

  it("照会で成功なしが確定したら編集でき、編集すると key を破棄する", () => {
    let s = issueReducer(previewed(), { type: "issueStart", pending });
    s = issueReducer(s, { type: "lookupNotSucceeded", notice: "not-succeeded" });
    expect(s.pending?.key).toBe(pending.key); // 同 key 再試行ができる
    s = issueReducer(s, { type: "edit", patch: { draft: { total: "1" } } });
    expect(s.pending).toBeNull();
  });

  it("前提変更(stale)では preview と key を破棄する", () => {
    const s = issueReducer(issueReducer(previewed(), { type: "issueStart", pending }), {
      type: "stale",
      notice: "dependency-changed",
    });
    expect(s.phase).toBe("stale");
    expect(s.preview).toBeNull();
    expect(s.pending).toBeNull();
  });
});

describe("発行済みからの新規発行", () => {
  it("issued から startNew で editing に戻り、preview / pending / issued を空にする", () => {
    let s = issueReducer(previewed(), { type: "issueStart", pending });
    s = issueReducer(s, { type: "issueOk", refs: { editionId: "1", artifactId: "2", replayed: false } });
    const next = issueReducer(s, { type: "startNew" });
    expect(next.phase).toBe("editing");
    expect(next.preview).toBeNull();
    expect(next.pending).toBeNull();
    expect(next.issued).toBeNull();
  });
  it("issued 以外では startNew は何もしない", () => {
    const s = previewed();
    expect(issueReducer(s, { type: "startNew" })).toBe(s);
  });
});

describe("発行失敗の分類", () => {
  const err = (kind: ConstructorParameters<typeof CertificateApiError>[0]) => new CertificateApiError(kind);
  it("結果不明の失敗は同 key の照会へ回す", () => {
    for (const kind of ["network", "result-unavailable", "render-unavailable", "operation-in-progress", "server"] as const) {
      expect(decideIssueFailure(err(kind)).type).toBe("unknown");
    }
  });
  it("前提変更系は stale（再 preview・新 key）", () => {
    expect(decideIssueFailure(err("dependency-changed"))).toMatchObject({ type: "stale", refetchContext: true });
    expect(decideIssueFailure(err("version-conflict"))).toMatchObject({ type: "stale", refetchContext: true });
    expect(decideIssueFailure(err("operation-key-conflict"))).toMatchObject({ type: "stale" });
    expect(decideIssueFailure(err("validation"))).toMatchObject({ type: "stale" });
  });
  it("権限・認証の失敗は blocked", () => {
    expect(decideIssueFailure(err("forbidden")).type).toBe("blocked");
    expect(decideIssueFailure(err("unauthenticated")).type).toBe("blocked");
  });
});

describe("権限なしからの再開", () => {
  it("未確定の key があれば捨てず、結果不明へ戻して同じ key で照会できる状態にする", () => {
    let s = issueReducer(previewed(), { type: "issueStart", pending });
    s = issueReducer(s, { type: "blocked", notice: "forbidden" });
    const resumed = issueReducer(s, { type: "unblock" });
    expect(resumed.phase).toBe("unknown");
    expect(resumed.pending).toBe(pending);
    expect(resumed.notice).toBe("unavailable");
  });

  it("key がなければ preview を引き継がず編集へ戻る", () => {
    const blocked = issueReducer(previewed(), { type: "blocked", notice: "forbidden" });
    const resumed = issueReducer(blocked, { type: "unblock" });
    expect(resumed.phase).toBe("editing");
    expect(resumed.preview).toBeNull();
    expect(resumed.previewRequest).toBeNull();
    expect(resumed.pending).toBeNull();
    expect(isEditable(resumed.phase)).toBe(true);
  });

  it("blocked 以外では何もしない", () => {
    const s = previewed();
    expect(issueReducer(s, { type: "unblock" })).toBe(s);
  });
});

describe("再 mount での復元", () => {
  it("未確定の key と request だけを結果不明として復元し、preview は持ち越さない", () => {
    const restored = restoredIssueState(pending);
    expect(restored.phase).toBe("unknown");
    expect(restored.pending).toBe(pending);
    expect(restored.kind).toBe("rabies");
    expect(restored.preview).toBeNull();
    expect(restored.previewRequest).toBeNull();
    expect(isEditable(restored.phase)).toBe(false);
  });
});

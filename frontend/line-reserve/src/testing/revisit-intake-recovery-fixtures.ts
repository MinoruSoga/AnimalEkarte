// EMR-474 operation-recovery Slice A: 契約どおりの合成 fixture / msw handler（テスト専用）。
// 本番コードから import しない（testing/revisit-intake-recovery-fixtures.guard.test.ts が検査する）。
// 実 API の失敗時に mock へフォールバックする経路は存在しない。値はすべて合成で、実在の対象ではない。
// 各 fixture は api/revisit-intake-recovery-schemas.ts の closed schema を満たす（required field を省かない）。
import { http, HttpResponse } from "msw";

export const RECOVERY_QUERY_PATH = "/api/liff/:clinicId/revisit-intake/operation-recovery/query";

export const KEY_A = "00000000-0000-4000-8000-0000000000a1";
export const KEY_B = "00000000-0000-4000-8000-0000000000b2";
export const KEY_C = "00000000-0000-4000-8000-0000000000c3";
export const DIGEST = "a".repeat(64);

type State = "succeeded" | "unresolved" | "closed_without_effect";

export interface FixtureItem {
  operation_key: string;
  intent: "revisit-intake.submit";
  claimed_at: string;
  discovery_state: "result" | "unavailable";
  execution_state: State;
  receipt_digest: string | null;
  result: unknown;
}

export function unresolvedResult(key: string, state: "claimed_uncommitted" | "result_unavailable") {
  return {
    operation_result: { operation_id: key, intent: "revisit-intake.submit", state },
    current_state: null,
    content_available: false,
  };
}

/** 現在本文を読めない本人の成功は既存 redacted 200 形（target / answer / 版 / 本文を足さない）。 */
export function redactedSuccessResult(key: string) {
  return {
    operation_result: {
      operation_id: key,
      intent: "revisit-intake.submit",
      state: "succeeded",
      committed_at: "2026-10-20T04:00:00Z",
      result_refs: null,
      committed_versions: null,
    },
    current_state: null,
    content_available: false,
  };
}

export function visibleSuccessResult(key: string) {
  return {
    operation_result: {
      operation_id: key,
      intent: "revisit-intake.submit",
      state: "succeeded",
      committed_at: "2026-10-20T04:00:00Z",
      result_refs: { appointment_id: "12001", pet_id: "801", target_id: "5401", answer_id: "9101" },
      committed_versions: { target_generation: "3", answer_revision: "1" },
    },
    current_state: null,
    content_available: true,
  };
}

export function succeededItem(key: string, claimedAt = "2026-10-20T04:00:00Z"): FixtureItem {
  return {
    operation_key: key,
    intent: "revisit-intake.submit",
    claimed_at: claimedAt,
    discovery_state: "result",
    execution_state: "succeeded",
    receipt_digest: DIGEST,
    result: visibleSuccessResult(key),
  };
}

export function unresolvedItem(key: string, claimedAt = "2026-10-20T03:00:00Z"): FixtureItem {
  return {
    operation_key: key,
    intent: "revisit-intake.submit",
    claimed_at: claimedAt,
    discovery_state: "result",
    execution_state: "unresolved",
    receipt_digest: null,
    result: unresolvedResult(key, "claimed_uncommitted"),
  };
}

export function unavailableItem(key: string, claimedAt = "2026-10-20T03:00:00Z"): FixtureItem {
  return {
    operation_key: key,
    intent: "revisit-intake.submit",
    claimed_at: claimedAt,
    discovery_state: "unavailable",
    execution_state: "unresolved",
    receipt_digest: null,
    result: null,
  };
}

export function closedItem(key: string, claimedAt = "2026-10-20T02:00:00Z"): FixtureItem {
  return {
    operation_key: key,
    intent: "revisit-intake.submit",
    claimed_at: claimedAt,
    discovery_state: "result",
    execution_state: "closed_without_effect",
    receipt_digest: null,
    result: null,
  };
}

export interface FixturePage {
  items: FixtureItem[];
  next_cursor: string | null;
  complete: boolean;
  recovery_generation: string | null;
  active_operation_key: string | null;
}

/** target scope: 世代は required 非 null（control 不存在は論理 "1"）。 */
export function targetPage(
  items: FixtureItem[] = [],
  extra: Partial<FixturePage> = {},
): FixturePage {
  return {
    items,
    next_cursor: null,
    complete: true,
    recovery_generation: "1",
    active_operation_key: null,
    ...extra,
  };
}

/** self scope: recovery_generation / active_operation_key は required null。 */
export function selfPage(items: FixtureItem[] = [], extra: Partial<FixturePage> = {}): FixturePage {
  return {
    items,
    next_cursor: null,
    complete: true,
    recovery_generation: null,
    active_operation_key: null,
    ...extra,
  };
}

export function noStoreJson(body: unknown, status = 200) {
  return HttpResponse.json(body, { status, headers: { "Cache-Control": "no-store" } });
}

export function recoveryError(status: number, code: string) {
  return noStoreJson({ error: "safe message", code }, status);
}

/** 受信した request body を記録しつつ、順に応答する msw handler。 */
export function recoveryHandler(
  respond: (body: Record<string, unknown>, index: number) => Response | Promise<Response>,
) {
  const bodies: Record<string, unknown>[] = [];
  const headers: Headers[] = [];
  const handler = http.post(RECOVERY_QUERY_PATH, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    bodies.push(body);
    headers.push(request.headers);
    return respond(body, bodies.length - 1);
  });
  return { handler, bodies, headers };
}

/**
 * 既存の問診ページテスト用の既定: 発見が 0 件で終わる応答（scope に合わせた形）。
 * 0 件・complete は保留理由が無いだけで、送信許可の根拠ではない（ページ側もそう扱う）。
 */
export function emptyRecoveryHandler() {
  return recoveryHandler((body) =>
    noStoreJson(body.scope === "self" ? selfPage() : targetPage()),
  ).handler;
}

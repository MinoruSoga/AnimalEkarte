import { http, HttpResponse } from "msw";
import type { ResponseResolverReturnType } from "msw";
import type {
  ShiftContractError,
  ShiftContractErrorCode,
  ShiftEntry,
  ShiftMutationIntent,
  ShiftMutationReceipt,
} from "./contract-types";

/** msw resolverが返せる応答（HttpResponse.error() / HttpResponse.json() の共用型） */
type MockResponse = Exclude<ResponseResolverReturnType, undefined | void>;
/**
 * EMR-514 固定API契約の MSW モック（テスト専用）。
 *
 * - app本体は axios の実経路のまま。このモックは `server.use(...)` で明示接続し、
 *   モックから本番通信へ黙って fallback しない（逆も同様）。
 * - 既定動作は in-memory store への stateful 適用（create/update/delete が receipt を
 *   記録し、GET /shifts/operations/{key} で照会できる）。
 * - シナリオ差し替えは scenario の該当endpointに関数を渡す（409の6code・400/401/403/404/500・通信断を含む）。
 */

/** 409 契約エラー応答（codeで分岐される対象） */
export function shiftContractErrorResponse(body: ShiftContractError): MockResponse {
  return HttpResponse.json(body, { status: 409 });
}

export type ShiftMockScenario = {
  listShifts?: () => MockResponse | undefined;
  createShift?: (body: unknown) => MockResponse | undefined;
  getShift?: (id: string) => MockResponse | undefined;
  updateShift?: (id: string, body: unknown) => MockResponse | undefined;
  deleteShift?: (id: string, query: URLSearchParams) => MockResponse | undefined;
  getOperation?: (key: string, intent: string) => MockResponse | undefined;
};

export interface ShiftMockStore {
  shifts: Map<string, ShiftEntry>;
  /** operation_key → 確定receipt（照会の正本） */
  receipts: Map<string, ShiftMutationReceipt>;
}

export function createShiftMockStore(initial: ShiftEntry[] = []): ShiftMockStore {
  const store: ShiftMockStore = {
    shifts: new Map(initial.map((s) => [s.id, s])),
    receipts: new Map(),
  };
  return store;
}

let mockSeq = 0;
export function nextMockId(): string {
  mockSeq += 1;
  return String(900000 + mockSeq);
}

export function makeMockShift(overrides: Partial<ShiftEntry> = {}): ShiftEntry {
  return {
    id: nextMockId(),
    clinic_id: "1",
    staff_id: "2",
    staff_name: "モックスタッフ",
    date: "2026-10-21",
    shift_type: "full",
    starts_at: "2026-10-21T09:00:00+09:00",
    ends_at: "2026-10-21T18:00:00+09:00",
    version: 1,
    interval_state: "known",
    notes: "",
    breaks: [],
    created_at: "2026-10-20T00:00:00+09:00",
    updated_at: "2026-10-20T00:00:00+09:00",
    ...overrides,
  };
}

function receipt(
  store: ShiftMockStore,
  intent: ShiftMutationIntent,
  shift: ShiftEntry | null,
  resultingVersion: number | null,
  operationKey: string,
  shiftId = shift?.id ?? "",
): ShiftMutationReceipt {
  const r: ShiftMutationReceipt = {
    operation_key: operationKey,
    intent,
    shift_id: shiftId,
    resulting_version: resultingVersion,
    shift,
  };
  store.receipts.set(operationKey, r);
  return r;
}

/**
 * 契約endpoint一式の MSW handler を返す。
 * scenario の該当関数が値を返したらそれを優先し、undefined なら既定（stateful成功）へ落ちる。
 * 通信断は `networkError: true` で作れる（HttpResponse.error → axiosは応答なしとして分類）。
 */
export function createShiftContractHandlers(
  store: ShiftMockStore,
  scenario: ShiftMockScenario = {},
) {
  return [
    http.get("*/v1/shifts", () => {
      const overridden = scenario.listShifts?.();
      if (overridden) return overridden;
      return HttpResponse.json([...store.shifts.values()]);
    }),

    http.post("*/v1/shifts", async ({ request }) => {
      const body = await request.json();
      const overridden = scenario.createShift?.(body);
      if (overridden) return overridden;
      const req = body as Record<string, unknown>;
      const key = String(req.operation_key ?? "");
      // 同じkeyの確定済み操作は元receiptを返す（重複効果を作らない）
      const existing = store.receipts.get(key);
      if (existing) return HttpResponse.json(existing, { status: 201 });
      const breaksReq = (req.breaks ?? []) as { starts_at: string; ends_at: string }[];
      const shift = makeMockShift({
        id: nextMockId(),
        staff_id: String(req.staff_id ?? ""),
        date: String(req.date ?? ""),
        shift_type: (req.shift_type ?? "full") as ShiftEntry["shift_type"],
        starts_at: (req.starts_at ?? null) as string | null,
        ends_at: (req.ends_at ?? null) as string | null,
        notes: String(req.notes ?? ""),
        breaks: breaksReq.map((b) => ({
          interval_state: "known" as const,
          starts_at: b.starts_at,
          ends_at: b.ends_at,
        })),
      });
      shift.version = 1;
      store.shifts.set(shift.id, shift);
      return HttpResponse.json(receipt(store, "create", shift, shift.version, key), {
        status: 201,
      });
    }),

    http.get("*/v1/shifts/operations/:operationKey", ({ params, request }) => {
      const key = String(params.operationKey ?? "");
      const intent = new URL(request.url).searchParams.get("intent") ?? "";
      const overridden = scenario.getOperation?.(key, intent);
      if (overridden) return overridden;
      const found = store.receipts.get(key);
      if (!found) {
        // 未記録のkeyは「証明できない」= unknown（不存在を未実行と推測しない）
        return HttpResponse.json({
          operation_key: key,
          state: "unknown",
          receipt: null,
        });
      }
      return HttpResponse.json({
        operation_key: key,
        state: "succeeded",
        receipt: found,
      });
    }),

    http.get("*/v1/shifts/:id", ({ params }) => {
      const id = String(params.id ?? "");
      const overridden = scenario.getShift?.(id);
      if (overridden) return overridden;
      const shift = store.shifts.get(id);
      if (!shift) return HttpResponse.json({ error: "not found" }, { status: 404 });
      return HttpResponse.json(shift);
    }),

    http.patch("*/v1/shifts/:id", async ({ params, request }) => {
      const id = String(params.id ?? "");
      const body = await request.json();
      const overridden = scenario.updateShift?.(id, body);
      if (overridden) return overridden;
      const current = store.shifts.get(id);
      if (!current) return HttpResponse.json({ error: "not found" }, { status: 404 });
      const req = body as Record<string, unknown>;
      const key = String(req.operation_key ?? "");
      // 同じkeyの確定済み操作は元receiptを返す（重複効果を作らない）
      const existing = store.receipts.get(key);
      if (existing) return HttpResponse.json(existing);
      // 期待版（aggregate版・休憩変更を含む）不一致は409 version_conflict
      const expectedVersion = Number(req.expected_version);
      if (Number.isInteger(expectedVersion) && expectedVersion !== current.version) {
        return shiftContractErrorResponse({
          error: "version mismatch",
          code: "version_conflict",
        });
      }
      const breaksPatchReq = (req.breaks ?? []) as { starts_at: string; ends_at: string }[];
      const updated: ShiftEntry = {
        ...current,
        shift_type: (req.shift_type ?? current.shift_type) as ShiftEntry["shift_type"],
        starts_at:
          "starts_at" in req ? ((req.starts_at ?? null) as string | null) : current.starts_at,
        ends_at: "ends_at" in req ? ((req.ends_at ?? null) as string | null) : current.ends_at,
        notes: "notes" in req ? String(req.notes ?? "") : current.notes,
        breaks:
          "breaks" in req
            ? breaksPatchReq.map((b) => ({
                interval_state: "known" as const,
                starts_at: b.starts_at,
                ends_at: b.ends_at,
              }))
            : current.breaks,
        version: current.version + 1,
        updated_at: "2026-10-20T12:00:00+09:00",
      };
      store.shifts.set(id, updated);
      return HttpResponse.json(receipt(store, "update", updated, updated.version, key));
    }),

    http.delete("*/v1/shifts/:id", ({ params, request }) => {
      const id = String(params.id ?? "");
      const query = new URL(request.url).searchParams;
      const overridden = scenario.deleteShift?.(id, query);
      if (overridden) return overridden;
      const current = store.shifts.get(id);
      if (!current) return HttpResponse.json({ error: "not found" }, { status: 404 });
      const key = query.get("operation_key") ?? "";
      // 同じkeyの確定済み操作は元receiptを返す（重複効果を作らない）
      const existing = store.receipts.get(key);
      if (existing) return HttpResponse.json(existing);
      // 期待版不一致（休憩変更を含む）は409 version_conflict
      const expectedVersion = Number(query.get("expected_version"));
      if (Number.isInteger(expectedVersion) && expectedVersion !== current.version) {
        return shiftContractErrorResponse({
          error: "version mismatch",
          code: "version_conflict",
        });
      }
      store.shifts.delete(id);
      // 削除の成功契約: shift=NULL・resulting_version=NULL。shift_idは対象id（pattern準拠）
      return HttpResponse.json(receipt(store, "delete", null, null, key, id));
    }),
  ];
}

/** 409シナリオの簡易ビルダ（6code分岐テスト用） */
export function conflictScenario(
  code: ShiftContractErrorCode,
  extra: Partial<ShiftContractError> = {},
): (body: unknown) => MockResponse {
  return () => shiftContractErrorResponse({ error: `conflict: ${code}`, code, ...extra });
}

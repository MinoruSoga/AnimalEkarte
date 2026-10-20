import { afterEach, describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import {
  conflictScenario,
  createShiftContractHandlers,
  createShiftMockStore,
  makeMockShift,
} from "./shift-contract-mock";
import { createShift } from "./create-shift";
import { updateShift } from "./update-shift";
import { deleteShift } from "./delete-shift";
import {
  getShiftOperationResult,
  classifyMutationError,
  isUnknownOutcome,
} from "./shift-operation";
import { extractShiftContractError } from "./contract-types";

function httpErrorStatus(err: unknown): number | undefined {
  const e = err as { response?: { status?: number } };
  return e?.response?.status;
}

afterEach(() => {
  server.resetHandlers();
});

describe("shift contract mock — stateful success", () => {
  it("createは201 receiptを返し、同keyの結果照会でsucceededになる", async () => {
    const store = createShiftMockStore();
    server.use(...createShiftContractHandlers(store));

    const receipt = await createShift({
      operation_key: "11111111-1111-4111-8111-111111111111",
      staff_id: "2",
      date: "2026-10-21",
      shift_type: "full",
      starts_at: "2026-10-21T09:00:00+09:00",
      ends_at: "2026-10-21T18:00:00+09:00",
      breaks: [{ starts_at: "2026-10-21T12:00:00+09:00", ends_at: "2026-10-21T13:00:00+09:00" }],
      notes: "",
    });

    expect(receipt.intent).toBe("create");
    expect(receipt.shift).not.toBeNull();
    expect(receipt.resulting_version).toBe(1);
    expect(receipt.shift?.breaks).toHaveLength(1);

    const result = await getShiftOperationResult(receipt.operation_key, "create");
    expect(result.state).toBe("succeeded");
    expect(result.receipt?.shift_id).toBe(receipt.shift_id);
  });

  it("create off/paid_leave は明示NULL・breaks[]をそのまま保存する", async () => {
    const store = createShiftMockStore();
    server.use(...createShiftContractHandlers(store));

    const receipt = await createShift({
      operation_key: "22222222-2222-4222-8222-222222222222",
      staff_id: "2",
      date: "2026-10-21",
      shift_type: "off",
      starts_at: null,
      ends_at: null,
      breaks: [],
      notes: "",
    });

    expect(receipt.shift?.shift_type).toBe("off");
    expect(receipt.shift?.starts_at).toBeNull();
    expect(receipt.shift?.ends_at).toBeNull();
    expect(receipt.shift?.breaks).toStrictEqual([]);
  });

  it("updateは配列全置換と版を進め、receiptのshiftは確定後のDTO", async () => {
    const initial = makeMockShift({
      id: "10",
      version: 2,
      breaks: [
        {
          interval_state: "known",
          starts_at: "2026-10-21T12:00:00+09:00",
          ends_at: "2026-10-21T13:00:00+09:00",
        },
      ],
    });
    const store = createShiftMockStore([initial]);
    server.use(...createShiftContractHandlers(store));

    const receipt = await updateShift("10", {
      operation_key: "33333333-3333-4333-8333-333333333333",
      expected_version: 2,
      breaks: [],
      notes: "",
    });

    expect(receipt.intent).toBe("update");
    expect(receipt.resulting_version).toBe(3);
    expect(receipt.shift?.breaks).toStrictEqual([]);
    expect(receipt.shift?.notes).toBe("");
    // 未指定の starts_at は維持
    expect(receipt.shift?.starts_at).toBe(initial.starts_at);
  });

  it("deleteはshift=NULL・resulting_version=NULLの成功契約のreceiptを返す", async () => {
    const store = createShiftMockStore([makeMockShift({ id: "20" })]);
    server.use(...createShiftContractHandlers(store));

    const receipt = await deleteShift({
      shiftId: "20",
      expectedVersion: 1,
      operationKey: "44444444-4444-4444-8444-444444444444",
    });

    expect(receipt.intent).toBe("delete");
    expect(receipt.shift).toBeNull();
    expect(receipt.resulting_version).toBeNull();
    // 一覧から消え、照会は同じ確定receiptを返す
    expect(store.shifts.has("20")).toBe(false);
    const result = await getShiftOperationResult("44444444-4444-4444-8444-444444444444", "delete");
    expect(result.state).toBe("succeeded");
    expect(result.receipt?.shift).toBeNull();
  });

  it("未記録keyの結果照会はunknown（不存在を未実行と推測しない）", async () => {
    const store = createShiftMockStore();
    server.use(...createShiftContractHandlers(store));

    const result = await getShiftOperationResult("55555555-5555-4555-8555-555555555555", "create");
    expect(result.state).toBe("unknown");
    expect(result.receipt).toBeNull();
  });
});

describe("shift contract mock — error & unknown scenarios", () => {
  it("409 version_conflict / protected_bookings / cross_clinic_overlap をcodeで取り出す", async () => {
    const store = createShiftMockStore();
    server.use(
      ...createShiftContractHandlers(store, {
        createShift: conflictScenario("version_conflict"),
      }),
    );
    try {
      await createShift({
        operation_key: "66666666-6666-4666-8666-666666666666",
        staff_id: "2",
        date: "2026-10-21",
        shift_type: "full",
        starts_at: "2026-10-21T09:00:00+09:00",
        ends_at: "2026-10-21T18:00:00+09:00",
        breaks: [],
        notes: "",
      });
      throw new Error("should not reach");
    } catch (err) {
      expect(httpErrorStatus(err)).toBe(409);
      const payload = extractShiftContractError(err);
      expect(payload?.code).toBe("version_conflict");
      const classified = classifyMutationError(err);
      expect(classified).toStrictEqual({
        kind: "contract",
        status: 409,
        payload,
      });
      expect(isUnknownOutcome(classified)).toBe(false);
    }

    server.use(
      ...createShiftContractHandlers(store, {
        createShift: conflictScenario("protected_bookings", {
          protected_bookings: [
            {
              appointment_id: "5",
              starts_at: "2026-10-21T10:00:00+09:00",
              ends_at: "2026-10-21T11:00:00+09:00",
              patient_label: "ハチ",
              reservation_type_label: "再診",
            },
          ],
        }),
      }),
    );
    try {
      await createShift({
        operation_key: "66666666-6666-4666-8666-666666666667",
        staff_id: "2",
        date: "2026-10-21",
        shift_type: "full",
        starts_at: "2026-10-21T09:00:00+09:00",
        ends_at: "2026-10-21T18:00:00+09:00",
        breaks: [],
        notes: "",
      });
      throw new Error("should not reach");
    } catch (err) {
      const payload = extractShiftContractError(err);
      expect(payload?.code).toBe("protected_bookings");
      expect(payload?.protected_bookings?.[0]?.patient_label).toBe("ハチ");
    }

    server.use(
      ...createShiftContractHandlers(store, {
        createShift: conflictScenario("cross_clinic_overlap", {
          conflicting_clinic_name: "B動物病院",
          overlap_starts_at: "2026-10-21T09:00:00+09:00",
          overlap_ends_at: "2026-10-21T10:00:00+09:00",
        }),
      }),
    );
    try {
      await createShift({
        operation_key: "66666666-6666-4666-8666-666666666668",
        staff_id: "2",
        date: "2026-10-21",
        shift_type: "full",
        starts_at: "2026-10-21T09:00:00+09:00",
        ends_at: "2026-10-21T18:00:00+09:00",
        breaks: [],
        notes: "",
      });
      throw new Error("should not reach");
    } catch (err) {
      const payload = extractShiftContractError(err);
      expect(payload?.code).toBe("cross_clinic_overlap");
      expect(payload?.conflicting_clinic_name).toBe("B動物病院");
    }
  });

  it("409の残りcode（operation_key_reused / shift_exists / source_unresolved）も分岐できる", async () => {
    const store = createShiftMockStore();
    for (const code of ["operation_key_reused", "shift_exists", "source_unresolved"] as const) {
      server.use(...createShiftContractHandlers(store, { createShift: conflictScenario(code) }));
      try {
        await createShift({
          operation_key: "77777777-7777-4777-8777-777777777777",
          staff_id: "2",
          date: "2026-10-21",
          shift_type: "full",
          starts_at: "2026-10-21T09:00:00+09:00",
          ends_at: "2026-10-21T18:00:00+09:00",
          breaks: [],
          notes: "",
        });
        throw new Error("should not reach");
      } catch (err) {
        expect(extractShiftContractError(err)?.code).toBe(code);
      }
    }
  });

  it("400/401/403/404はhttpとして分類され、409契約外shapeはcontractにならない", async () => {
    const store = createShiftMockStore();
    // 401はグローバルaxiosインタセプタがauth/refreshを呼ぶため、その応答も固定する
    server.use(
      http.post("*/v1/auth/refresh", () =>
        HttpResponse.json({ error: "unauthorized" }, { status: 401 }),
      ),
    );
    for (const status of [400, 401, 403, 404]) {
      server.use(
        ...createShiftContractHandlers(store, {
          createShift: () => HttpResponseJsonError(status),
        }),
      );
      try {
        await createShift({
          operation_key: "88888888-8888-4888-8888-888888888888",
          staff_id: "2",
          date: "2026-10-21",
          shift_type: "full",
          starts_at: "2026-10-21T09:00:00+09:00",
          ends_at: "2026-10-21T18:00:00+09:00",
          breaks: [],
          notes: "",
        });
        throw new Error("should not reach");
      } catch (err) {
        expect(httpErrorStatus(err)).toBe(status);
        expect(classifyMutationError(err)).toStrictEqual({ kind: "http", status });
      }
    }
  });

  it("500と通信断はunknown扱い（isUnknownOutcome）になり、同key照会へ回せる", async () => {
    const store = createShiftMockStore();
    server.use(
      ...createShiftContractHandlers(store, {
        createShift: () => HttpResponseJsonError(500),
      }),
    );
    try {
      await createShift({
        operation_key: "99999999-9999-4999-8999-999999999999",
        staff_id: "2",
        date: "2026-10-21",
        shift_type: "full",
        starts_at: "2026-10-21T09:00:00+09:00",
        ends_at: "2026-10-21T18:00:00+09:00",
        breaks: [],
        notes: "",
      });
      throw new Error("should not reach");
    } catch (err) {
      const classified = classifyMutationError(err);
      expect(isUnknownOutcome(classified)).toBe(true);
    }

    server.resetHandlers();
    server.use(
      ...createShiftContractHandlers(store, {
        createShift: () => HttpResponse.error(),
      }),
    );
    try {
      await createShift({
        operation_key: "99999999-9999-4999-8999-99999999999a",
        staff_id: "2",
        date: "2026-10-21",
        shift_type: "full",
        starts_at: "2026-10-21T09:00:00+09:00",
        ends_at: "2026-10-21T18:00:00+09:00",
        breaks: [],
        notes: "",
      });
      throw new Error("should not reach");
    } catch (err) {
      expect(classifyMutationError(err).kind).toBe("network");
    }
  });
});

function HttpResponseJsonError(status: number) {
  return HttpResponse.json({ error: `http ${status}` }, { status });
}

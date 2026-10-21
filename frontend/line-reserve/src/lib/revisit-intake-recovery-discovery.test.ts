import { describe, expect, it, vi } from "vitest";

const fence = vi.hoisted(() => ({ available: true }));
vi.mock("./revisit-intake-recovery-fence", () => ({
  get LIFF_REVISIT_INTAKE_RECOVERY_FENCE_AVAILABLE() {
    return fence.available;
  },
}));
import {
  RECOVERY_MAX_PAGES,
  compareRfc3339,
  discoverRevisitIntakeRecovery,
  summarizeRevisitIntakeRecovery,
  type RecoveryPageFetcher,
} from "./revisit-intake-recovery-discovery";
import { RevisitIntakeRecoveryQueryError } from "../api/revisit-intake-recovery";
import {
  KEY_A,
  KEY_B,
  KEY_C,
  closedItem,
  selfPage,
  succeededItem,
  targetPage,
  unavailableItem,
  unresolvedItem,
  type FixturePage,
} from "../testing/revisit-intake-recovery-fixtures";
import type { RevisitIntakeRecoveryPage } from "../api/revisit-intake-recovery-schemas";

const TARGET = {
  scope: "target" as const,
  targetRef: { appointment_id: "12001", pet_id: "801", target_generation: "3" },
};

function fetcherOf(pages: FixturePage[]): RecoveryPageFetcher & ReturnType<typeof vi.fn> {
  let i = 0;
  return vi.fn((_request: unknown) => {
    const page = pages[i];
    i += 1;
    if (page === undefined) return Promise.reject(new Error("no more pages"));
    return Promise.resolve(page as unknown as RevisitIntakeRecoveryPage);
  }) as unknown as RecoveryPageFetcher & ReturnType<typeof vi.fn>;
}

async function expectViolation(promise: Promise<unknown>) {
  await expect(promise).rejects.toBeInstanceOf(RevisitIntakeRecoveryQueryError);
  await expect(promise).rejects.toMatchObject({ kind: "contract_violation" });
}

describe("discoverRevisitIntakeRecovery（paging）", () => {
  it("next_cursor を page ごとに追い、cursor をそのまま次 request に載せる", async () => {
    const fetchPage = fetcherOf([
      targetPage([succeededItem(KEY_C, "2026-10-20T05:00:00Z")], { next_cursor: "c1", complete: false }),
      targetPage([unresolvedItem(KEY_B, "2026-10-20T04:00:00Z")], { next_cursor: "c2", complete: false }),
      targetPage([closedItem(KEY_A, "2026-10-20T03:00:00Z")]),
    ]);
    const d = await discoverRevisitIntakeRecovery(fetchPage, TARGET);
    expect(d.items.map((i) => i.operation_key)).toEqual([KEY_C, KEY_B, KEY_A]);
    expect(fetchPage.mock.calls.map((c) => (c[0] as { cursor: string | null }).cursor)).toEqual([
      null,
      "c1",
      "c2",
    ]);
    expect(d.scanComplete).toBe(true);
    expect(d.truncated).toBe(false);
    expect(d.recoveryGeneration).toBe("1");
  });

  it("同じ cursor の再利用（loop）は契約違反", async () => {
    await expectViolation(
      discoverRevisitIntakeRecovery(
        fetcherOf([
          targetPage([], { next_cursor: "c1", complete: false }),
          targetPage([], { next_cursor: "c1", complete: false }),
        ]),
        TARGET,
      ),
    );
  });

  it("page 数の上限で打ち切り、truncated として保留する（完結扱いにしない）", async () => {
    const pages = Array.from({ length: RECOVERY_MAX_PAGES + 5 }, (_, i) =>
      targetPage([], { next_cursor: `c${i}`, complete: false }),
    );
    const fetchPage = fetcherOf(pages);
    const d = await discoverRevisitIntakeRecovery(fetchPage, TARGET);
    expect(fetchPage).toHaveBeenCalledTimes(RECOVERY_MAX_PAGES);
    expect(d.truncated).toBe(true);
    expect(d.scanComplete).toBe(false);
    expect(summarizeRevisitIntakeRecovery(d)).toMatchObject({ newSubmissionBlocked: true });
  });

  it("page 間の世代・active 変動、重複 key、順序違反は契約違反", async () => {
    await expectViolation(
      discoverRevisitIntakeRecovery(
        fetcherOf([
          targetPage([], { next_cursor: "c1", complete: false }),
          targetPage([], { recovery_generation: "2" }),
        ]),
        TARGET,
      ),
    );
    await expectViolation(
      discoverRevisitIntakeRecovery(
        fetcherOf([
          targetPage([succeededItem(KEY_A)], { next_cursor: "c1", complete: false }),
          targetPage([succeededItem(KEY_A, "2026-10-19T00:00:00Z")]),
        ]),
        TARGET,
      ),
    );
    await expectViolation(
      discoverRevisitIntakeRecovery(
        fetcherOf([
          targetPage([
            succeededItem(KEY_A, "2026-10-20T01:00:00Z"),
            succeededItem(KEY_B, "2026-10-20T02:00:00Z"),
          ]),
        ]),
        TARGET,
      ),
    );
  });

  it("page 取得の失敗はそのまま伝え、途中までの結果を完結として返さない", async () => {
    const fetchPage = vi.fn()
      .mockResolvedValueOnce(targetPage([unresolvedItem(KEY_A)], { next_cursor: "c1", complete: false }))
      .mockRejectedValueOnce(new RevisitIntakeRecoveryQueryError("result_unavailable", "x"));
    await expect(
      discoverRevisitIntakeRecovery(fetchPage as unknown as RecoveryPageFetcher, TARGET),
    ).rejects.toMatchObject({ kind: "result_unavailable" });
  });

  it("発見全体の deadline を超えたら完結扱いにせず失敗する", async () => {
    const fetchPage = fetcherOf([
      targetPage([], { next_cursor: "c1", complete: false }),
      targetPage([]),
    ]);
    await expect(
      discoverRevisitIntakeRecovery(fetchPage, TARGET, { deadlineMs: -1 }),
    ).rejects.toMatchObject({ kind: "network" });
    expect(fetchPage).not.toHaveBeenCalled();
  });

  it("同一 claimed_at は operation_key DESC を強制する", async () => {
    const t = "2026-10-20T01:00:00Z";
    await expectViolation(
      discoverRevisitIntakeRecovery(
        fetcherOf([targetPage([succeededItem(KEY_A, t), succeededItem(KEY_B, t)])]),
        TARGET,
      ),
    );
    const d = await discoverRevisitIntakeRecovery(
      fetcherOf([targetPage([succeededItem(KEY_B, t), succeededItem(KEY_A, t)])]),
      TARGET,
    );
    expect(d.items.map((i) => i.operation_key)).toEqual([KEY_B, KEY_A]);
  });

  it("ms を超える精度（マイクロ秒）と表記違いの同一 instant を全精度で比較する", async () => {
    expect(compareRfc3339("2026-10-20T01:00:00.000100Z", "2026-10-20T01:00:00.000200Z")).toBe(-1);
    expect(compareRfc3339("2026-10-20T01:00:00.5Z", "2026-10-20T01:00:00.500000Z")).toBe(0);
    expect(compareRfc3339("2026-10-20T10:00:00+09:00", "2026-10-20T01:00:00Z")).toBe(0);
    expect(compareRfc3339("2026-10-20T01:00:60Z", "2026-10-20T01:00:00Z")).toBeNull(); // 閏秒は fail-closed
    // 同一 ms 内でも新しい方が後ろに来たら契約違反
    await expectViolation(
      discoverRevisitIntakeRecovery(
        fetcherOf([
          targetPage([
            succeededItem(KEY_B, "2026-10-20T01:00:00.000100Z"),
            succeededItem(KEY_C, "2026-10-20T01:00:00.000200Z"),
          ]),
        ]),
        TARGET,
      ),
    );
    // 同一 instant の別表記は key DESC を要求する
    await expectViolation(
      discoverRevisitIntakeRecovery(
        fetcherOf([
          targetPage([
            succeededItem(KEY_A, "2026-10-20T10:00:00+09:00"),
            succeededItem(KEY_B, "2026-10-20T01:00:00Z"),
          ]),
        ]),
        TARGET,
      ),
    );
    const ok = await discoverRevisitIntakeRecovery(
      fetcherOf([
        targetPage([
          succeededItem(KEY_B, "2026-10-20T10:00:00+09:00"),
          succeededItem(KEY_A, "2026-10-20T01:00:00Z"),
        ]),
      ]),
      TARGET,
    );
    expect(ok.items).toHaveLength(2);
  });

  it("page 要求の pending 中にも deadline を適用する（signal を無視する fetcher でも止める）", async () => {
    const hang: RecoveryPageFetcher = () => new Promise(() => undefined);
    await expect(
      discoverRevisitIntakeRecovery(hang, TARGET, { deadlineMs: 30 }),
    ).rejects.toMatchObject({ kind: "network" });
  });

  it("deadline 後に完了した page は採用しない", async () => {
    const late: RecoveryPageFetcher = (_r, signal) =>
      new Promise((resolve) => {
        signal?.addEventListener("abort", () => resolve(targetPage() as never));
      });
    await expect(
      discoverRevisitIntakeRecovery(late, TARGET, { deadlineMs: 30 }),
    ).rejects.toMatchObject({ kind: "network" });
  });

  it("self scope は self request を送る", async () => {
    const fetchPage = fetcherOf([selfPage([])]);
    await discoverRevisitIntakeRecovery(fetchPage, { scope: "self" });
    expect(fetchPage.mock.calls[0][0]).toEqual({ scope: "self", cursor: null });
  });
});

describe("summarizeRevisitIntakeRecovery（保留理由。許可は返さない）", () => {
  async function summarize(page: FixturePage) {
    return summarizeRevisitIntakeRecovery(await discoverRevisitIntakeRecovery(fetcherOf([page]), TARGET));
  }

  it("unresolved / unavailable / active key / complete=false は保留", async () => {
    expect((await summarize(targetPage([unresolvedItem(KEY_A)]))).reasons).toEqual(["unresolved_operation"]);
    expect((await summarize(targetPage([unavailableItem(KEY_A)]))).reasons).toEqual(["result_unavailable"]);
    expect((await summarize(targetPage([succeededItem(KEY_A)], { active_operation_key: KEY_A }))).reasons).toEqual([
      "active_operation",
    ]);
    expect((await summarize(targetPage([], { complete: false }))).reasons).toEqual(["scan_incomplete"]);
  });

  it("succeeded / closed_without_effect だけなら保留理由なし（これは許可ではなく理由が無いだけ）", async () => {
    const s = await summarize(targetPage([succeededItem(KEY_B, "2026-10-20T05:00:00Z"), closedItem(KEY_A)]));
    expect(s).toEqual({ newSubmissionBlocked: false, hasOperationHold: false, reasons: [] });
    // 許可を表す field は存在しない
    expect(Object.keys(s).sort()).toEqual(["hasOperationHold", "newSubmissionBlocked", "reasons"]);
  });

  it("Slice A（fence 不可）: 0 件・complete・成功のみでも常に保留（fence_unavailable）", async () => {
    fence.available = false;
    try {
      expect(await summarize(targetPage([]))).toEqual({
        newSubmissionBlocked: true,
        hasOperationHold: false,
        reasons: ["fence_unavailable"],
      });
      const s = await summarize(targetPage([unresolvedItem(KEY_A)]));
      expect(s.newSubmissionBlocked).toBe(true);
      expect(s.hasOperationHold).toBe(true);
    } finally {
      fence.available = true;
    }
  });
});

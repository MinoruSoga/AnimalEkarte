// EMR-474 operation-recovery Slice A: 再読込後の未確定操作の発見・保留・再確認（RevisitIntakePage 結線）。
// 実 route は未実装のため msw で契約どおりの応答を返す。fixture は合成値。
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

const fence = vi.hoisted(() => ({ available: false }));
vi.mock("../lib/revisit-intake-recovery-fence", () => ({
  get LIFF_REVISIT_INTAKE_RECOVERY_FENCE_AVAILABLE() {
    return fence.available;
  },
}));

vi.mock("@line/liff", () => ({
  default: { isInClient: () => true, sendMessages: vi.fn(() => Promise.resolve()) },
}));

import { RevisitIntakePage } from "./RevisitIntakePage";
import {
  findPendingEntry,
  hasDraftEntry,
  type RevisitIntakeDraftEntry,
  type RevisitIntakeDraftStore,
} from "../hooks/use-revisit-intake-drafts";
import {
  KEY_A,
  KEY_B,
  closedItem,
  noStoreJson,
  recoveryError,
  recoveryHandler,
  redactedSuccessResult,
  selfPage,
  succeededItem,
  targetPage,
  unavailableItem,
  unresolvedItem,
  unresolvedResult,
} from "../testing/revisit-intake-recovery-fixtures";

const CLINIC = "1";
const APPT = "12001";
const READ_PATH = "/api/liff/:clinicId/my-reservations/:appointmentId/revisit-intake";
const SUBMIT_PATH = "/api/liff/:clinicId/my-reservations/:appointmentId/revisit-intake/submissions";
const LOOKUP_PATH = "/api/liff/:clinicId/revisit-intake/operations/:operationKey";
const SUBMIT_LABEL = "回答を送信する";

function readBody() {
  return {
    availability: "available",
    form_definition_version: 1,
    current_target: {
      ref: { target_id: null, appointment_id: APPT, pet_id: "801", target_generation: "3" },
      head_revision: null,
      answer_state: "unanswered",
      latest_answer: null,
      content_available: false,
      editable: true,
      edit_block_reason: null,
    },
    history: [],
    next_cursor: null,
  };
}

function createStore(): RevisitIntakeDraftStore {
  const map = new Map<string, RevisitIntakeDraftEntry>();
  return {
    get: (k) => map.get(k),
    set: (k, v) => {
      map.set(k, v);
    },
    clear: (k) => {
      map.delete(k);
    },
    hasDraft: (c, a) => hasDraftEntry(map, c, a),
    findPending: (c, a) => findPendingEntry(map, c, a),
  };
}

function mockRead(status = 200) {
  server.use(
    http.get(READ_PATH, () =>
      status === 200
        ? HttpResponse.json(readBody())
        : HttpResponse.json({ error: "denied", code: "FORBIDDEN" }, { status }),
    ),
  );
}

/** 想定外の送信を検知する（発見・保留の間に submit が出てはならない）。 */
function trapSubmits(): { count: () => number } {
  let n = 0;
  server.use(
    http.post(SUBMIT_PATH, () => {
      n += 1;
      return HttpResponse.json({}, { status: 500 });
    }),
  );
  return { count: () => n };
}

function renderPage() {
  return render(
    <RevisitIntakePage
      clinicId={CLINIC}
      idToken="synthetic-id-token"
      appointmentId={APPT}
      draftStore={createStore()}
      onBack={vi.fn()}
    />,
  );
}

describe("RevisitIntakePage 操作回復 Slice A", () => {
  it("target scope で発見し、未確定があれば editor を出さず保留する", async () => {
    mockRead();
    const traps = trapSubmits();
    const rec = recoveryHandler(() => noStoreJson(targetPage([unresolvedItem(KEY_A)])));
    server.use(rec.handler);
    renderPage();

    expect(await screen.findByText(/新しい回答の送信を一時停止/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
    expect(rec.bodies[0]).toEqual({
      scope: "target",
      target_ref: { appointment_id: APPT, pet_id: "801", target_generation: "3" },
      cursor: null,
    });
    expect(traps.count()).toBe(0);
  });

  it("結果照会は既存 route・同じ key で行い、照会の結果だけでは保留を解かない（発見をやり直す）", async () => {
    mockRead();
    const lookedUp: string[] = [];
    server.use(
      http.get(LOOKUP_PATH, ({ params }) => {
        lookedUp.push(String(params.operationKey));
        return HttpResponse.json(unresolvedResult(KEY_A, "claimed_uncommitted"));
      }),
    );
    const rec = recoveryHandler(() => noStoreJson(targetPage([unresolvedItem(KEY_A)])));
    server.use(rec.handler);
    renderPage();

    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "送信結果を確認" }));
    expect(await screen.findByText(/保存はまだ確認できていません/)).toBeInTheDocument();
    expect(lookedUp).toEqual([KEY_A]);
    await waitFor(() => expect(rec.bodies.length).toBe(2)); // 再発見
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
  });

  it("再照会が 429 + Retry-After: その item の button を期限まで止め、保留は解かない", async () => {
    mockRead();
    server.use(
      http.get(LOOKUP_PATH, () =>
        HttpResponse.json({ error: "slow", code: "RATE_LIMITED" }, { status: 429, headers: { "Retry-After": "30" } }),
      ),
    );
    const rec = recoveryHandler(() => noStoreJson(targetPage([unresolvedItem(KEY_A)])));
    server.use(rec.handler);
    renderPage();
    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "送信結果を確認" }));
    expect(await screen.findByText("30秒後に再確認できます")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "送信結果を確認" })).toBeDisabled();
    expect(rec.bodies.length).toBe(1); // 再発見もしない
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
  });

  it("結果を読めない item（unavailable）も保留する", async () => {
    mockRead();
    server.use(recoveryHandler(() => noStoreJson(targetPage([unavailableItem(KEY_A)]))).handler);
    renderPage();
    expect(await screen.findByText("送信結果を読み取れていません。")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
  });

  it("Slice A: 0 件・complete でも editor を出さず、fence 待ちの停止告知を出す", async () => {
    mockRead();
    const traps = trapSubmits();
    server.use(recoveryHandler(() => noStoreJson(targetPage([]))).handler);
    renderPage();
    expect(await screen.findByText(/操作の回復確認機能が利用できるようになるまで/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "もう一度確認する" })).not.toBeInTheDocument();
    expect(traps.count()).toBe(0);
  });

  it("Slice B 相当（fence 利用可）でのみ、保留理由が無い発見の後に editor が出る", async () => {
    fence.available = true;
    try {
      mockRead();
      server.use(recoveryHandler(() => noStoreJson(targetPage([]))).handler);
      renderPage();
      expect(await screen.findByRole("button", { name: SUBMIT_LABEL })).toBeInTheDocument();
    } finally {
      fence.available = false;
    }
  });

  it("succeeded は既存 result の範囲（完了と日時）だけ表示し、redacted は本文を足さない", async () => {
    mockRead();
    server.use(
      recoveryHandler(() =>
        noStoreJson(
          targetPage([
            { ...succeededItem(KEY_B), result: redactedSuccessResult(KEY_B) },
            closedItem(KEY_A),
          ]),
        ),
      ).handler,
    );
    renderPage();
    expect(await screen.findByText("回答の送信は完了しています。")).toBeInTheDocument();
    expect(screen.getByText("回答内容はこの画面では表示できません。")).toBeInTheDocument();
    expect(screen.getByText(/反映されずに終了した送信が1件/)).toBeInTheDocument();
    // 成功・終了の表示は残るが、Slice A では editor は出さない
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
  });

  it("active_operation_key が残る間は（fence は後続 Slice のため）新規送信を止める", async () => {
    mockRead();
    server.use(
      recoveryHandler(() =>
        noStoreJson(targetPage([succeededItem(KEY_A)], { active_operation_key: KEY_A, recovery_generation: "2" })),
      ).handler,
    );
    renderPage();
    expect(await screen.findByText(/直前の操作の確認が済むまで/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
  });

  it("発見に失敗したら fail-closed で止め、再確認で成功して初めて editor が出る", async () => {
    mockRead();
    const rec = recoveryHandler((_b, i) =>
      i === 0 ? recoveryError(503, "RESULT_UNAVAILABLE") : noStoreJson(targetPage([])),
    );
    server.use(rec.handler);
    renderPage();
    expect(await screen.findByText(/確認できるまで、新しい回答は送信できません/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
    await userEvent.setup().click(screen.getByRole("button", { name: "もう一度確認する" }));
    // 再確認が成功しても Slice A では editor を出さない（fence 待ちの告知へ変わる）
    expect(await screen.findByText(/操作の回復確認機能が利用できるようになるまで/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
  });

  it("通信断・timeout（応答なし）は network 文言と再確認 button を出し、editor は出さない", async () => {
    mockRead();
    server.use(http.post("/api/liff/:clinicId/revisit-intake/operation-recovery/query", () => HttpResponse.error()));
    renderPage();
    expect(await screen.findByText(/通信できませんでした/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "もう一度確認する" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
  });

  it("契約違反の応答（未知 field）も保留側へ倒す", async () => {
    mockRead();
    server.use(
      recoveryHandler(() => noStoreJson({ ...targetPage([]), extra: true })).handler,
    );
    renderPage();
    expect(await screen.findByText(/確認できるまで、新しい回答は送信できません/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
  });

  it("target が 404 SCOPE_NOT_FOUND なら self scope へ切り替える", async () => {
    mockRead();
    const rec = recoveryHandler((body) =>
      body.scope === "target"
        ? recoveryError(404, "SCOPE_NOT_FOUND")
        : noStoreJson(selfPage([unresolvedItem(KEY_A)])),
    );
    server.use(rec.handler);
    renderPage();
    expect(await screen.findByText(/新しい回答の送信を一時停止/)).toBeInTheDocument();
    expect(rec.bodies.map((b) => b.scope)).toEqual(["target", "self"]);
    expect(rec.bodies[1]).toEqual({ scope: "self", cursor: null });
  });

  it("予約の読込が拒否（403）されたら self scope で発見し、未確定を表示する", async () => {
    mockRead(403);
    const rec = recoveryHandler(() => noStoreJson(selfPage([unresolvedItem(KEY_A)])));
    server.use(rec.handler);
    renderPage();
    expect(await screen.findByText(/新しい回答の送信を一時停止/)).toBeInTheDocument();
    expect(rec.bodies[0]).toEqual({ scope: "self", cursor: null });
  });

  it("読込拒否 + self の成功のみ: 別予約の成功を出さず従来どおり ErrorPage のまま", async () => {
    mockRead(403);
    const rec = recoveryHandler(() => noStoreJson(selfPage([succeededItem(KEY_A), closedItem(KEY_B, "2026-10-19T00:00:00Z")])));
    server.use(rec.handler);
    renderPage();
    expect(await screen.findByText(/この予約の問診を表示できません/)).toBeInTheDocument();
    await waitFor(() => expect(rec.bodies.length).toBe(1));
    expect(screen.queryByText("回答の送信は完了しています。")).not.toBeInTheDocument();
    expect(screen.queryByText(/反映されずに終了/)).not.toBeInTheDocument();
  });

  it("同 target の memory pending + 発見の保留: 結果照会だけ出し、成功で store が空になり Back は確認しない", async () => {
    mockRead();
    server.use(
      recoveryHandler(() => noStoreJson(targetPage([unavailableItem(KEY_B)]))).handler,
      http.get(LOOKUP_PATH, () => HttpResponse.json(redactedSuccessResult(KEY_A))),
    );
    const store = createStore();
    const targetKey = `${CLINIC}|${APPT}|801|3`;
    store.set(targetKey, {
      values: {
        visit_purpose: "合成",
        onset: null,
        energy: null,
        appetite: null,
        current_symptoms: null,
        improvement: null,
        examination_preference: null,
        main_concern: null,
      },
      pending: {
        key: KEY_A,
        fingerprint: "fp",
        body: {} as never,
        status: "unknown",
      },
    });
    const onBack = vi.fn();
    render(
      <RevisitIntakePage
        clinicId={CLINIC}
        idToken="synthetic-id-token"
        appointmentId={APPT}
        draftStore={store}
        onBack={onBack}
      />,
    );
    const user = userEvent.setup();
    expect(screen.queryByRole("button", { name: SUBMIT_LABEL })).not.toBeInTheDocument();
    const buttons = await screen.findAllByRole("button", { name: "送信結果を確認" });
    expect(buttons.length).toBe(2); // 発見 item 用 + 同 target pending 用
    await user.click(buttons[buttons.length - 1]);
    await waitFor(() => expect(store.findPending(CLINIC, APPT)).toBeUndefined());
    await user.click(screen.getByRole("button", { name: /戻る/ }));
    expect(onBack).toHaveBeenCalled();
  });

  it("page を最後までたどる（最初の page が空でも 2 page 目の未確定を見落とさない）", async () => {
    mockRead();
    const rec = recoveryHandler((body) =>
      body.cursor === null
        ? noStoreJson(targetPage([succeededItem(KEY_B, "2026-10-20T05:00:00Z")], { next_cursor: "pg2", complete: false }))
        : noStoreJson(targetPage([unresolvedItem(KEY_A, "2026-10-20T03:00:00Z")])),
    );
    server.use(rec.handler);
    renderPage();
    expect(await screen.findByText(/新しい回答の送信を一時停止/)).toBeInTheDocument();
    expect(rec.bodies.map((b) => b.cursor)).toEqual([null, "pg2"]);
  });

  it("body・operation key・世代を storage へ保存しない", async () => {
    mockRead();
    server.use(recoveryHandler(() => noStoreJson(targetPage([unresolvedItem(KEY_A)]))).handler);
    renderPage();
    await screen.findByText(/新しい回答の送信を一時停止/);
    expect(window.localStorage.length).toBe(0);
    expect(window.sessionStorage.length).toBe(0);
    expect(window.location.search).toBe("");
    expect(window.location.hash).toBe("");
  });
});

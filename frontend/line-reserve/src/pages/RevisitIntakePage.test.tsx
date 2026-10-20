// EMR-474: 再診事前問診画面のページ単位テスト。3 route とも runtime 未実装のため msw で契約どおりの応答を返す。
// fixture はすべて合成値（実在の医院・飼主・ペットではない）。
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { http, HttpResponse, delay } from "msw";
import { server } from "@/testing/mocks/node";

const { keyMock, sendMessagesMock } = vi.hoisted(() => ({
  keyMock: vi.fn<() => string>(),
  sendMessagesMock: vi.fn(() => Promise.resolve()),
}));

vi.mock("../lib/revisit-intake-operation-key", () => ({
  createRevisitIntakeOperationKey: keyMock,
}));

// 送信後に LINE トークへの通知副作用を追加しないことを観測する。
vi.mock("@line/liff", () => ({
  default: { isInClient: () => true, sendMessages: sendMessagesMock },
}));

import { RevisitIntakePage } from "./RevisitIntakePage";
import type { RevisitIntakeDraftStore } from "../hooks/use-revisit-intake-drafts";
import type {
  RevisitIntakeAnswer,
  RevisitIntakeFullAnswers,
  RevisitIntakeHistoryEntry,
  RevisitIntakeLIFFRead,
  RevisitIntakeOwnerCurrentTarget,
} from "../types/models";

const CLINIC = "1";
const APPT_A = "12001";
const APPT_B = "12002";
const PET_A = "801";
const PET_B = "802";
const KEYS = [
  "00000000-0000-4000-8000-000000000001",
  "00000000-0000-4000-8000-000000000002",
  "00000000-0000-4000-8000-000000000003",
];

const READ_PATH = "/api/liff/:clinicId/my-reservations/:appointmentId/revisit-intake";
const SUBMIT_PATH = "/api/liff/:clinicId/my-reservations/:appointmentId/revisit-intake/submissions";
const LOOKUP_PATH = "/api/liff/:clinicId/revisit-intake/operations/:operationKey";

const EMPTY: RevisitIntakeFullAnswers = {
  visit_purpose: "",
  onset: null,
  energy: null,
  appetite: null,
  current_symptoms: null,
  improvement: null,
  examination_preference: null,
  main_concern: null,
};

function answers(partial: Partial<RevisitIntakeFullAnswers>): RevisitIntakeFullAnswers {
  return { ...EMPTY, visit_purpose: "合成の来院目的", ...partial };
}

function answer(
  revision: string,
  body: RevisitIntakeFullAnswers,
  receivedAt = "2026-10-20T01:00:00Z",
): RevisitIntakeAnswer {
  return {
    answer_id: `9${revision}0`,
    revision,
    form_definition_version: 1,
    received_at: receivedAt,
    provenance: "owner_submission",
    answers: body,
  };
}

function unansweredTarget(
  appointmentId = APPT_A,
  petId = PET_A,
  generation = "3",
): RevisitIntakeOwnerCurrentTarget {
  return {
    ref: {
      target_id: null,
      appointment_id: appointmentId,
      pet_id: petId,
      target_generation: generation,
    },
    head_revision: null,
    answer_state: "unanswered",
    latest_answer: null,
    content_available: false,
    editable: true,
    edit_block_reason: null,
  };
}

function answeredTarget(
  latest: RevisitIntakeAnswer,
  overrides: Partial<RevisitIntakeOwnerCurrentTarget> = {},
): RevisitIntakeOwnerCurrentTarget {
  return {
    ref: { target_id: "5401", appointment_id: APPT_A, pet_id: PET_A, target_generation: "3" },
    head_revision: latest.revision,
    answer_state: "answered",
    latest_answer: latest,
    content_available: true,
    editable: true,
    edit_block_reason: null,
    ...overrides,
  };
}

function read(
  current: RevisitIntakeOwnerCurrentTarget | null,
  extra: Partial<RevisitIntakeLIFFRead> = {},
): RevisitIntakeLIFFRead {
  return {
    availability: "available",
    form_definition_version: 1,
    current_target: current,
    history: [],
    next_cursor: null,
    ...extra,
  };
}

function successResponse(key: string, revision: string, appointmentId = APPT_A, petId = PET_A) {
  return {
    operation_result: {
      operation_id: key,
      intent: "revisit-intake.submit",
      state: "succeeded",
      committed_at: "2026-10-20T04:00:00Z",
      result_refs: {
        appointment_id: appointmentId,
        pet_id: petId,
        target_id: "5401",
        answer_id: "9101",
      },
      committed_versions: { target_generation: "3", answer_revision: revision },
    },
    current_state: null,
    content_available: true,
  };
}

function unresolvedResponse(key: string, state: "claimed_uncommitted" | "result_unavailable") {
  return {
    operation_result: { operation_id: key, intent: "revisit-intake.submit", state },
    current_state: null,
    content_available: false,
  };
}

function lookupNotFound(key: string) {
  return {
    error: "not found",
    code: "NOT_FOUND",
    operation_result: { operation_id: key, intent: "revisit-intake.submit", state: "not_found" },
    current_state: null,
    content_available: false,
  };
}

interface CapturedSubmit {
  appointmentId: string;
  key: string | null;
  body: Record<string, unknown>;
}

/** GET は呼ばれるたびに reads を順に返す（最後の要素を繰り返す）。 */
function mockReadSequence(
  reads: RevisitIntakeLIFFRead[] | Record<string, RevisitIntakeLIFFRead[]>,
): { calls: URL[] } {
  const calls: URL[] = [];
  const counters: Record<string, number> = {};
  server.use(
    http.get(READ_PATH, ({ request, params }) => {
      const url = new URL(request.url);
      calls.push(url);
      const appointmentId = String(params.appointmentId);
      const list = Array.isArray(reads) ? reads : (reads[appointmentId] ?? []);
      const n = counters[appointmentId] ?? 0;
      counters[appointmentId] = n + 1;
      return HttpResponse.json(list[Math.min(n, list.length - 1)]);
    }),
  );
  return { calls };
}

function captureSubmits(
  respond: (captured: CapturedSubmit, index: number) => Response | Promise<Response>,
): CapturedSubmit[] {
  const captured: CapturedSubmit[] = [];
  server.use(
    http.post(SUBMIT_PATH, async ({ request, params }) => {
      const entry: CapturedSubmit = {
        appointmentId: String(params.appointmentId),
        key: request.headers.get("Idempotency-Key"),
        body: (await request.json()) as Record<string, unknown>,
      };
      captured.push(entry);
      return respond(entry, captured.length - 1);
    }),
  );
  return captured;
}

function captureLookups(respond: (key: string, index: number) => Response): string[] {
  const keys: string[] = [];
  server.use(
    http.get(LOOKUP_PATH, ({ params }) => {
      const key = String(params.operationKey);
      keys.push(key);
      return respond(key, keys.length - 1);
    }),
  );
  return keys;
}

function createStore(): RevisitIntakeDraftStore {
  const map = new Map<string, ReturnType<RevisitIntakeDraftStore["get"]>>();
  return {
    get: (k) => map.get(k),
    set: (k, v) => {
      map.set(k, v);
    },
    clear: (k) => {
      map.delete(k);
    },
  };
}

function renderPage(
  options: { appointmentId?: string; store?: RevisitIntakeDraftStore; onBack?: () => void } = {},
) {
  const onBack = options.onBack ?? vi.fn();
  const store = options.store ?? createStore();
  const utils = render(
    <RevisitIntakePage
      clinicId={CLINIC}
      idToken="synthetic-id-token"
      appointmentId={options.appointmentId ?? APPT_A}
      draftStore={store}
      onBack={onBack}
    />,
  );
  return { ...utils, onBack, store };
}

function purposeInput(): HTMLTextAreaElement {
  return screen.getByLabelText(/^来院目的/) as HTMLTextAreaElement;
}

beforeEach(() => {
  let n = 0;
  keyMock.mockReset();
  keyMock.mockImplementation(
    () => KEYS[n++] ?? `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`,
  );
  sendMessagesMock.mockClear();
});

afterEach(() => {
  try {
    window.localStorage.clear();
    window.sessionStorage.clear();
  } catch {
    // storage が使えない環境では何もしない
  }
});

describe("RevisitIntakePage 表示状態", () => {
  it("未回答: 空 form・任意の注記・送信ボタンが活性で表示される", async () => {
    mockReadSequence([read(unansweredTarget())]);
    renderPage();

    expect(await screen.findByText("未回答")).toBeInTheDocument();
    expect(
      screen.getByText(/回答は任意です。回答しなくても予約・受付には影響しません/),
    ).toBeInTheDocument();
    expect(purposeInput().value).toBe("");
    expect(screen.getByRole("button", { name: "回答を送信する" })).toBeEnabled();
  });

  it("回答済み・editable: 最新回答を prefill し、回答日時と最終訂正日時を表示する", async () => {
    const latest = answer(
      "2",
      answers({ visit_purpose: "合成の訂正後目的", energy: "usual" }),
      "2026-10-20T03:30:00Z",
    );
    const first = answer("1", answers({ visit_purpose: "合成の初回目的" }), "2026-10-19T01:00:00Z");
    const history: RevisitIntakeHistoryEntry[] = [
      {
        target: { ...answeredTarget(latest).ref, target_id: "5401" },
        answer: latest,
        target_relation: "current",
      },
      {
        target: { ...answeredTarget(latest).ref, target_id: "5401" },
        answer: first,
        target_relation: "current",
      },
    ];
    mockReadSequence([read(answeredTarget(latest), { history })]);
    renderPage();

    expect(await screen.findByText("回答済み")).toBeInTheDocument();
    expect(purposeInput().value).toBe("合成の訂正後目的");
    const energy = screen.getByRole("group", { name: /^元気/ });
    expect(within(energy).getByLabelText("いつも通り")).toBeChecked();
    expect(screen.getByText("回答日時 2026年10月19日 10:00")).toBeInTheDocument();
    expect(screen.getByText("最終訂正日時 2026年10月20日 12:30")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "訂正して再送信" })).toBeEnabled();
  });

  it("回答済みだが本文不可視: 回答済みと表示し本文も prefill もせず、期待版は head_revision で送信する", async () => {
    const user = userEvent.setup();
    const hidden = {
      ...answeredTarget(answer("3", answers({ visit_purpose: "不可視の本文" }))),
      latest_answer: null,
      content_available: false,
    };
    mockReadSequence([read(hidden)]);
    const submits = captureSubmits((c) => HttpResponse.json(successResponse(c.key ?? "", "4")));
    renderPage();

    expect(await screen.findByText(/回答内容はこの画面では表示できません/)).toBeInTheDocument();
    expect(screen.queryByText("不可視の本文")).not.toBeInTheDocument();
    expect(purposeInput().value).toBe("");

    await user.type(purposeInput(), "合成の新しい目的");
    await user.click(screen.getByRole("button", { name: "訂正して再送信" }));

    await waitFor(() => expect(submits).toHaveLength(1));
    expect(submits[0].body.expected_revision).toBe("3");
  });

  it.each([
    ["reception_started", "受付後の変更はスタッフへお伝えください。"],
    ["cancelled", "この予約はキャンセルされているため、問診の回答・訂正はできません。"],
    ["deleted", "この予約は取り消されているため、問診の回答・訂正はできません。"],
    ["authorization_denied", /権限を確認できませんでした/],
    ["unsupported_lifecycle", /現在の予約状態では問診の回答・訂正ができません/],
  ] as const)("editable=false（%s）: 理由別の文言で閲覧のみにする", async (reason, message) => {
    const latest = answer("1", answers({ visit_purpose: "合成の閲覧のみ目的" }));
    mockReadSequence([
      read(answeredTarget(latest, { editable: false, edit_block_reason: reason })),
    ]);
    renderPage();

    expect(await screen.findByText(message)).toBeInTheDocument();
    expect(screen.getByText("合成の閲覧のみ目的")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /送信/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("target_unresolved: 対象を確認できない旨を表示し form を出さない", async () => {
    mockReadSequence([read(null, { availability: "target_unresolved" })]);
    renderPage();

    expect(await screen.findByText(/対象ペットを確認できないため/)).toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("403 / 404 は ErrorPage を表示し回答の有無を推測しない", async () => {
    server.use(
      http.get(READ_PATH, () => HttpResponse.json({ error: "forbidden" }, { status: 403 })),
    );
    const { onBack } = renderPage();
    const user = userEvent.setup();

    expect(await screen.findByText(/この予約の問診を表示できません/)).toBeInTheDocument();
    expect(screen.queryByText("未回答")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "戻る" }));
    expect(onBack).toHaveBeenCalled();
  });

  it("401 は再ログイン案内を表示し再試行ボタンを出さない", async () => {
    server.use(
      http.get(READ_PATH, () => HttpResponse.json({ error: "unauthorized" }, { status: 401 })),
    );
    renderPage();

    expect(await screen.findByText(/ログイン情報の有効期限が切れました/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "再試行" })).not.toBeInTheDocument();
  });
});

describe("RevisitIntakePage 履歴（AC-reservations-revisit-new-target / revisions）", () => {
  it("current と previous を分けて read-only 表示し、previous を入力へコピーしない", async () => {
    const prevAnswer = answer("1", answers({ visit_purpose: "以前の対象の合成目的" }));
    const history: RevisitIntakeHistoryEntry[] = [
      {
        target: {
          target_id: "5300",
          appointment_id: APPT_A,
          pet_id: "799",
          target_generation: "2",
        },
        answer: prevAnswer,
        target_relation: "previous",
      },
    ];
    mockReadSequence([read(unansweredTarget(), { history })]);
    renderPage();

    expect(await screen.findByText("以前の対象への回答（参照のみ）")).toBeInTheDocument();
    expect(screen.getByText("以前の対象の合成目的")).toBeInTheDocument();
    expect(purposeInput().value).toBe("");
    expect(screen.getByRole("button", { name: "回答を送信する" })).toBeInTheDocument();
  });

  it("next_cursor があれば「さらに表示」で cursor 付きの次ページを読み込む", async () => {
    const user = userEvent.setup();
    const ref = {
      target_id: "5401",
      appointment_id: APPT_A,
      pet_id: PET_A,
      target_generation: "3",
    };
    const page1 = read(answeredTarget(answer("2", answers({ visit_purpose: "合成2版" }))), {
      history: [
        {
          target: ref,
          answer: answer("2", answers({ visit_purpose: "合成2版" })),
          target_relation: "current",
        },
      ],
      next_cursor: "synthetic-cursor-1",
    });
    const page2 = read(answeredTarget(answer("2", answers({ visit_purpose: "合成2版" }))), {
      history: [
        {
          target: ref,
          answer: answer("1", answers({ visit_purpose: "合成1版" })),
          target_relation: "current",
        },
      ],
      next_cursor: null,
    });
    const { calls } = mockReadSequence([page1, page2]);
    renderPage();

    await user.click(await screen.findByRole("button", { name: "さらに表示" }));

    expect(await screen.findByText("合成1版")).toBeInTheDocument();
    expect(calls[1].searchParams.get("cursor")).toBe("synthetic-cursor-1");
    expect(calls[0].searchParams.toString()).toBe("");
    expect(screen.queryByRole("button", { name: "さらに表示" })).not.toBeInTheDocument();
  });
});

describe("RevisitIntakePage 送信（AC-reservations-revisit-optional / revisions）", () => {
  it("来院目的が空なら欄直下にエラーを出してフォーカスし、送信しない", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits(() => HttpResponse.json({}));
    renderPage();

    await user.click(await screen.findByRole("button", { name: "回答を送信する" }));

    expect(await screen.findByText(/来院目的を入力してください/)).toBeInTheDocument();
    expect(purposeInput()).toHaveFocus();
    expect(purposeInput()).toHaveAttribute("aria-invalid", "true");
    expect(submits).toHaveLength(0);
  });

  it("来院目的が Unicode 空白だけ（全角空白・改行）なら送信しない", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits(() => HttpResponse.json({}));
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "　{enter} ");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    expect(await screen.findByText(/来院目的を入力してください/)).toBeInTheDocument();
    expect(submits).toHaveLength(0);
  });

  it("任意欄を補完せずに送信し、成功後は回答済みとして再表示する。通知（sendMessages）は呼ばない", async () => {
    const user = userEvent.setup();
    const after = read(
      answeredTarget(answer("1", answers({ visit_purpose: "合成の咳の経過", onset: "" }))),
    );
    mockReadSequence([read(unansweredTarget()), after]);
    const submits = captureSubmits((c) => HttpResponse.json(successResponse(c.key ?? "", "1")));
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "合成の咳の経過");
    // いつから: 入力して消した空文字は null と区別して送る
    await user.type(screen.getByLabelText(/^いつから/), "x");
    await user.clear(screen.getByLabelText(/^いつから/));
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    expect(await screen.findByText(/回答を送信しました/)).toBeInTheDocument();
    expect(submits).toHaveLength(1);
    expect(submits[0].key).toBe(KEYS[0]);
    expect(submits[0].body).toEqual({
      pet_id: PET_A,
      target_generation: "3",
      expected_revision: null,
      form_definition_version: 1,
      answers: {
        visit_purpose: "合成の咳の経過",
        onset: "",
        energy: null,
        appetite: null,
        current_symptoms: null,
        improvement: null,
        examination_preference: null,
        main_concern: null,
      },
    });
    expect(await screen.findByText("回答済み")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "訂正して再送信" })).toBeInTheDocument();
    expect(sendMessagesMock).not.toHaveBeenCalled();
  });

  it("成功後の訂正は新しい key と現在の head_revision で送信する", async () => {
    const user = userEvent.setup();
    const v1 = read(answeredTarget(answer("1", answers({ visit_purpose: "合成目的1" }))));
    const v2 = read(answeredTarget(answer("2", answers({ visit_purpose: "合成目的2" }))));
    mockReadSequence([v1, v2]);
    const submits = captureSubmits((c, i) =>
      HttpResponse.json(successResponse(c.key ?? "", String(i + 2))),
    );
    renderPage();

    await user.clear(await screen.findByLabelText(/^来院目的/));
    await user.type(purposeInput(), "合成目的2");
    await user.click(screen.getByRole("button", { name: "訂正して再送信" }));
    await waitFor(() => expect(submits).toHaveLength(1));
    // 成功後の再取得（head=2）が反映されるまで待つ
    expect(await screen.findByText(/^最終訂正日時/)).toBeInTheDocument();
    expect(purposeInput().value).toBe("合成目的2");

    await user.clear(purposeInput());
    await user.type(purposeInput(), "合成目的3");
    await user.click(screen.getByRole("button", { name: "訂正して再送信" }));
    await waitFor(() => expect(submits).toHaveLength(2));

    expect(submits[0].body.expected_revision).toBe("1");
    expect(submits[1].body.expected_revision).toBe("2");
    expect(submits[0].key).toBe(KEYS[0]);
    expect(submits[1].key).toBe(KEYS[1]);
  });

  it("単一選択は任意で、選択を解除すると null で送信する", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits((c) => HttpResponse.json(successResponse(c.key ?? "", "1")));
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "合成目的");
    const appetite = screen.getByRole("group", { name: /^食欲/ });
    await user.click(within(appetite).getByLabelText("少し低下"));
    await user.click(
      within(screen.getByRole("group", { name: /^検査希望/ })).getByLabelText(
        "獣医師と相談して決めたい",
      ),
    );
    await user.click(
      within(screen.getByRole("group", { name: /^前回より改善/ })).getByLabelText(
        "新しい症状が出た",
      ),
    );
    await user.click(screen.getByRole("button", { name: "検査希望の有無の選択を解除" }));
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    await waitFor(() => expect(submits).toHaveLength(1));
    const sent = submits[0].body.answers as RevisitIntakeFullAnswers;
    expect(sent.appetite).toBe("slightly_reduced");
    expect(sent.energy).toBeNull();
    expect(sent.improvement).toBe("new_symptoms");
    expect(sent.examination_preference).toBeNull();
  });

  it("送信中は aria-busy を立て、連打しても POST は1回だけ", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits(async (c) => {
      await delay(200);
      return HttpResponse.json(successResponse(c.key ?? "", "1"));
    });
    const { container } = renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "合成目的");
    const button = screen.getByRole("button", { name: "回答を送信する" });
    await user.dblClick(button);
    await user.click(button);

    await waitFor(() =>
      expect(container.querySelector("form")).toHaveAttribute("aria-busy", "true"),
    );
    expect(await screen.findByText(/回答を送信しました/)).toBeInTheDocument();
    expect(submits).toHaveLength(1);
  });

  it("400 INVALID_INPUT は該当欄の直下へエラーを出し先頭エラー欄へフォーカスする", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    captureSubmits(() =>
      HttpResponse.json(
        {
          error: "invalid input",
          code: "INVALID_INPUT",
          details: { fields: [{ path: "answers.visit_purpose", reason: "blank" }] },
        },
        { status: 400 },
      ),
    );
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "​");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    expect(await screen.findByText(/来院目的を入力してください/)).toBeInTheDocument();
    expect(purposeInput()).toHaveFocus();
  });
});

describe("RevisitIntakePage 409", () => {
  it("REVISION_CONFLICT: 入力を保持し再取得を案内、再取得後は新しい key と新しい head で再送信する", async () => {
    const user = userEvent.setup();
    const before = read(answeredTarget(answer("1", answers({ visit_purpose: "合成目的1" }))));
    const latest = read(
      answeredTarget(answer("2", answers({ visit_purpose: "他端末の合成目的" }))),
    );
    const { calls } = mockReadSequence([before, latest]);
    const submits = captureSubmits((c, i) =>
      i === 0
        ? HttpResponse.json(
            { error: "conflict", code: "REVISION_CONFLICT", details: { head_revision: "2" } },
            { status: 409 },
          )
        : HttpResponse.json(successResponse(c.key ?? "", "3")),
    );
    renderPage();

    await user.clear(await screen.findByLabelText(/^来院目的/));
    await user.type(purposeInput(), "合成の私の訂正");
    await user.click(screen.getByRole("button", { name: "訂正して再送信" }));

    expect(await screen.findByText(/ほかの操作で回答が更新されています/)).toBeInTheDocument();
    expect(purposeInput().value).toBe("合成の私の訂正");
    expect(screen.getByRole("button", { name: "訂正して再送信" })).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "最新の状態を読み込む" }));
    await waitFor(() => expect(calls).toHaveLength(2));
    await waitFor(() =>
      expect(screen.getByRole("button", { name: "訂正して再送信" })).toBeEnabled(),
    );
    expect(purposeInput().value).toBe("合成の私の訂正");

    await user.click(screen.getByRole("button", { name: "訂正して再送信" }));
    await waitFor(() => expect(submits).toHaveLength(2));
    expect(submits[1].body.expected_revision).toBe("2");
    expect(submits[1].key).not.toBe(submits[0].key);
  });

  it("FORM_VERSION_MISMATCH: 再取得するまで送信できない", async () => {
    const user = userEvent.setup();
    const { calls } = mockReadSequence([read(unansweredTarget())]);
    captureSubmits(() =>
      HttpResponse.json(
        {
          error: "conflict",
          code: "FORM_VERSION_MISMATCH",
          details: { supported_form_definition_version: 1 },
        },
        { status: 409 },
      ),
    );
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "合成目的");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    expect(await screen.findByText(/問診の項目が更新されています/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "回答を送信する" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "最新の状態を読み込む" }));
    await waitFor(() => expect(calls).toHaveLength(2));
    await waitFor(() =>
      expect(screen.getByRole("button", { name: "回答を送信する" })).toBeEnabled(),
    );
  });

  it("TARGET_CHANGED: 再取得し、新しい対象には旧対象の入力を引き継がない", async () => {
    const user = userEvent.setup();
    const { calls } = mockReadSequence([
      read(unansweredTarget(APPT_A, PET_A, "3")),
      read(unansweredTarget(APPT_A, "803", "4")),
    ]);
    captureSubmits(() =>
      HttpResponse.json(
        { error: "conflict", code: "TARGET_CHANGED", details: null },
        { status: 409 },
      ),
    );
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "旧対象への合成入力");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    expect(await screen.findByText(/予約の対象が変更されたため/)).toBeInTheDocument();
    await waitFor(() => expect(calls).toHaveLength(2));
    await waitFor(() => expect(purposeInput().value).toBe(""));
  });

  it("INTAKE_CLOSED: 閲覧へ切り替える", async () => {
    const user = userEvent.setup();
    const closed = read(
      answeredTarget(answer("1", answers({ visit_purpose: "合成目的1" })), {
        editable: false,
        edit_block_reason: "reception_started",
      }),
    );
    mockReadSequence([
      read(answeredTarget(answer("1", answers({ visit_purpose: "合成目的1" })))),
      closed,
    ]);
    captureSubmits(() =>
      HttpResponse.json(
        { error: "conflict", code: "INTAKE_CLOSED", details: { reason: "reception_started" } },
        { status: 409 },
      ),
    );
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "追記");
    await user.click(screen.getByRole("button", { name: "訂正して再送信" }));

    expect(
      await screen.findByText(/受付が始まったため、問診は送信されませんでした/),
    ).toBeInTheDocument();
    expect(await screen.findByText("受付後の変更はスタッフへお伝えください。")).toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("KEY_CONTENT_MISMATCH: 次の送信は新しい key で行う", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits((c, i) =>
      i === 0
        ? HttpResponse.json(
            { error: "conflict", code: "KEY_CONTENT_MISMATCH", details: null },
            { status: 409 },
          )
        : HttpResponse.json(successResponse(c.key ?? "", "1")),
    );
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "合成目的");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));
    expect(await screen.findByText(/送信内容を確認できませんでした/)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "回答を送信する" }));
    await waitFor(() => expect(submits).toHaveLength(2));
    expect(submits[1].key).not.toBe(submits[0].key);
  });
});

describe("RevisitIntakePage 応答喪失と結果照会", () => {
  async function submitAndLoseResponse(user: ReturnType<typeof userEvent.setup>) {
    await user.type(await screen.findByLabelText(/^来院目的/), "合成目的");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));
  }

  it("通信断 → 同じ key で照会し succeeded なら成功表示にする（新 key で再送しない）", async () => {
    const user = userEvent.setup();
    mockReadSequence([
      read(unansweredTarget()),
      read(answeredTarget(answer("1", answers({ visit_purpose: "合成目的" })))),
    ]);
    const submits = captureSubmits(() => HttpResponse.error());
    const lookups = captureLookups((key) => HttpResponse.json(successResponse(key, "1")));
    renderPage();

    await submitAndLoseResponse(user);

    expect(await screen.findByText(/回答を送信しました/)).toBeInTheDocument();
    expect(submits).toHaveLength(1);
    expect(lookups).toEqual([KEYS[0]]);
    expect(keyMock).toHaveBeenCalledTimes(1);
  });

  it("5xx → 照会 claimed_uncommitted → 同じ key・同じ内容で再送できる", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits((c, i) =>
      i === 0
        ? HttpResponse.json({ error: "internal", code: "INTERNAL", details: null }, { status: 500 })
        : HttpResponse.json(successResponse(c.key ?? "", "1")),
    );
    captureLookups((key) => HttpResponse.json(unresolvedResponse(key, "claimed_uncommitted")));
    renderPage();

    await submitAndLoseResponse(user);

    expect(await screen.findByText(/保存の完了を確認できていません/)).toBeInTheDocument();
    expect(purposeInput()).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "同じ内容で再送信" }));

    expect(await screen.findByText(/回答を送信しました/)).toBeInTheDocument();
    expect(submits).toHaveLength(2);
    expect(submits[1].key).toBe(submits[0].key);
    expect(submits[1].body).toEqual(submits[0].body);
    expect(keyMock).toHaveBeenCalledTimes(1);
  });

  it("照会 404 not_found → 失敗確定とせず同じ key で再送できる", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits(() => HttpResponse.error());
    captureLookups((key) => HttpResponse.json(lookupNotFound(key), { status: 404 }));
    renderPage();

    await submitAndLoseResponse(user);

    expect(await screen.findByText(/送信の記録が見つかりませんでした/)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "同じ内容で再送信" }));
    await waitFor(() => expect(submits).toHaveLength(2));
    expect(submits[1].key).toBe(KEYS[0]);
    expect(keyMock).toHaveBeenCalledTimes(1);
  });

  it("照会 result_unavailable → 「結果を確認できません」で新規送信・再送をしない", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits(() => HttpResponse.error());
    const lookups = captureLookups((key) =>
      HttpResponse.json(unresolvedResponse(key, "result_unavailable")),
    );
    renderPage();

    await submitAndLoseResponse(user);

    expect(await screen.findByText(/結果を確認できません/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "同じ内容で再送信" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /回答を送信する/ })).not.toBeInTheDocument();
    expect(purposeInput()).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "送信結果を確認" }));
    await waitFor(() => expect(lookups).toHaveLength(2));
    expect(lookups[1]).toBe(KEYS[0]);
    expect(submits).toHaveLength(1);
    expect(keyMock).toHaveBeenCalledTimes(1);
  });

  it("POST 自体が claimed_uncommitted を返したときも同じ key で再送する", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits((c, i) =>
      i === 0
        ? HttpResponse.json(unresolvedResponse(c.key ?? "", "claimed_uncommitted"))
        : HttpResponse.json(successResponse(c.key ?? "", "1")),
    );
    renderPage();

    await submitAndLoseResponse(user);
    await user.click(await screen.findByRole("button", { name: "同じ内容で再送信" }));

    expect(await screen.findByText(/回答を送信しました/)).toBeInTheDocument();
    expect(submits.map((s) => s.key)).toEqual([KEYS[0], KEYS[0]]);
  });
});

describe("RevisitIntakePage 再試行の抑止（レビュー対応）", () => {
  it("TARGET_CHANGED 後に再取得が失敗しても、旧 target のまま入力・送信できない", async () => {
    const user = userEvent.setup();
    let reads = 0;
    server.use(
      http.get(READ_PATH, () => {
        reads += 1;
        if (reads === 1) return HttpResponse.json(read(unansweredTarget()));
        return HttpResponse.json({ error: "unavailable" }, { status: 503 });
      }),
    );
    const submits = captureSubmits(() =>
      HttpResponse.json(
        { error: "conflict", code: "TARGET_CHANGED", details: null },
        { status: 409 },
      ),
    );
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "旧対象への合成入力");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    expect(
      await screen.findByText(
        "予約の対象が変更されました。最新の内容を読み込むまで送信できません。",
      ),
    ).toBeInTheDocument();
    await waitFor(() => expect(reads).toBe(2));
    expect(purposeInput()).toBeDisabled();
    expect(screen.getByRole("button", { name: "回答を送信する" })).toBeDisabled();
    expect(submits).toHaveLength(1);
  });

  it("成功後の再取得が失敗している間は、旧 head のまま訂正送信できない", async () => {
    const user = userEvent.setup();
    let reads = 0;
    server.use(
      http.get(READ_PATH, () => {
        reads += 1;
        if (reads === 1) return HttpResponse.json(read(unansweredTarget()));
        return HttpResponse.json({ error: "unavailable" }, { status: 503 });
      }),
    );
    const submits = captureSubmits((c) => HttpResponse.json(successResponse(c.key ?? "", "1")));
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "合成目的");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    expect(await screen.findByText(/回答を送信しました/)).toBeInTheDocument();
    expect(
      await screen.findByText(
        "最新の状態を読み込めませんでした。読み込みに成功するまで送信できません。",
      ),
    ).toBeInTheDocument();
    await user.type(purposeInput(), "の訂正");
    expect(screen.getByRole("button", { name: "回答を送信する" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "最新の状態を読み込む" })).toBeInTheDocument();
    expect(submits).toHaveLength(1);
  });

  it("429 は Retry-After の期限まで「同じ内容で再送信」を無効にし、期限後は同じ key で再送できる", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits((c, i) =>
      i === 0
        ? HttpResponse.json(
            { error: "too many requests", code: "TOO_MANY_REQUESTS", details: null },
            { status: 429, headers: { "Retry-After": "1" } },
          )
        : HttpResponse.json(successResponse(c.key ?? "", "1")),
    );
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "合成目的");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    const resend = await screen.findByRole("button", { name: "同じ内容で再送信" });
    expect(resend).toBeDisabled();
    await waitFor(
      () => expect(screen.getByRole("button", { name: "同じ内容で再送信" })).toBeEnabled(),
      {
        timeout: 3000,
      },
    );
    await user.click(screen.getByRole("button", { name: "同じ内容で再送信" }));
    await waitFor(() => expect(submits).toHaveLength(2));
    expect(submits[1].key).toBe(submits[0].key);
  });

  it("新規送信の 401 は再認証案内を出し、送信ボタンを無効にして新しい key を作らない", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const submits = captureSubmits(() =>
      HttpResponse.json(
        { error: "unauthorized", code: "UNAUTHORIZED", details: null },
        { status: 401 },
      ),
    );
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "合成目的");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));

    expect(await screen.findByText(/ログイン情報の有効期限が切れました/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "回答を送信する" })).toBeDisabled();
    expect(submits).toHaveLength(1);
    expect(keyMock).toHaveBeenCalledTimes(1);
  });
});

describe("RevisitIntakePage draft（target 別 memory のみ）", () => {
  it("draft は target 別に保持し、別予約（別ペット）へ混ざらない（AC-reservations-revisit-multi-pet）", async () => {
    const user = userEvent.setup();
    mockReadSequence({
      [APPT_A]: [read(unansweredTarget(APPT_A, PET_A))],
      [APPT_B]: [read(unansweredTarget(APPT_B, PET_B))],
    });
    const store = createStore();

    const first = renderPage({ appointmentId: APPT_A, store });
    await user.type(await screen.findByLabelText(/^来院目的/), "ペットA用の合成入力");
    first.unmount();

    const second = renderPage({ appointmentId: APPT_B, store });
    expect((await screen.findByLabelText(/^来院目的/)) as HTMLTextAreaElement).toHaveValue("");
    await user.type(purposeInput(), "ペットB用の合成入力");
    second.unmount();

    renderPage({ appointmentId: APPT_A, store });
    expect(await screen.findByLabelText(/^来院目的/)).toHaveValue("ペットA用の合成入力");
  });

  it("2頭の別予約へ異なる回答を送信すると、それぞれの予約・ペットへ別々に送る", async () => {
    const user = userEvent.setup();
    mockReadSequence({
      [APPT_A]: [read(unansweredTarget(APPT_A, PET_A))],
      [APPT_B]: [read(unansweredTarget(APPT_B, PET_B))],
    });
    const submits = captureSubmits((c) =>
      HttpResponse.json(successResponse(c.key ?? "", "1", c.appointmentId, String(c.body.pet_id))),
    );

    const first = renderPage({ appointmentId: APPT_A });
    await user.type(await screen.findByLabelText(/^来院目的/), "合成A");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));
    await screen.findByText(/回答を送信しました/);
    first.unmount();

    renderPage({ appointmentId: APPT_B });
    await user.type(await screen.findByLabelText(/^来院目的/), "合成B");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));
    await screen.findByText(/回答を送信しました/);

    expect(submits.map((s) => [s.appointmentId, s.body.pet_id])).toEqual([
      [APPT_A, PET_A],
      [APPT_B, PET_B],
    ]);
    expect((submits[0].body.answers as RevisitIntakeFullAnswers).visit_purpose).toBe("合成A");
    expect((submits[1].body.answers as RevisitIntakeFullAnswers).visit_purpose).toBe("合成B");
    expect(submits[0].key).not.toBe(submits[1].key);
  });

  it("入力本文・予約 ID を URL / localStorage / sessionStorage へ保存しない", async () => {
    const user = userEvent.setup();
    const setItem = vi.spyOn(Storage.prototype, "setItem");
    const hrefBefore = window.location.href;
    mockReadSequence([read(unansweredTarget())]);
    captureSubmits((c) => HttpResponse.json(successResponse(c.key ?? "", "1")));
    renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "保存されてはいけない合成本文");
    await user.click(screen.getByRole("button", { name: "回答を送信する" }));
    await screen.findByText(/回答を送信しました/);

    expect(setItem).not.toHaveBeenCalled();
    expect(window.location.href).toBe(hrefBefore);
    expect(window.location.href).not.toContain(APPT_A);
    expect(window.localStorage.length).toBe(0);
    expect(window.sessionStorage.length).toBe(0);
    setItem.mockRestore();
  });

  it("dirty のときだけ戻る操作で確認し、未入力ならそのまま戻る", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const { onBack } = renderPage();

    await screen.findByLabelText(/^来院目的/);
    await user.type(purposeInput(), "合成入力");
    await user.click(screen.getByRole("button", { name: "戻る" }));

    expect(
      screen.getByText("回答はまだ送信されていません。このまま戻りますか？"),
    ).toBeInTheDocument();
    expect(onBack).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "入力を続ける" }));
    expect(purposeInput().value).toBe("合成入力");

    await user.clear(purposeInput());
    await act(async () => {
      await user.click(screen.getByRole("button", { name: "戻る" }));
    });
    expect(onBack).toHaveBeenCalledTimes(1);
  });

  it("dirty で「送信せずに戻る」を選ぶと戻る", async () => {
    const user = userEvent.setup();
    mockReadSequence([read(unansweredTarget())]);
    const { onBack } = renderPage();

    await user.type(await screen.findByLabelText(/^来院目的/), "合成入力");
    await user.click(screen.getByRole("button", { name: "戻る" }));
    await user.click(screen.getByRole("button", { name: "送信せずに戻る" }));

    expect(onBack).toHaveBeenCalledTimes(1);
  });
});

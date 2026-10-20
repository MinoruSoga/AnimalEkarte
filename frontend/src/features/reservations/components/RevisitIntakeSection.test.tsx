import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { RevisitIntakeSection } from "./RevisitIntakeSection";

const PATH = "/api/v1/reservations/:id/revisit-intake";

// fixture は合成値のみ。
function ref(petId = "10", generation = "2") {
  return { target_id: "5", appointment_id: "88", pet_id: petId, target_generation: generation };
}

function answer(
  id: string,
  revision: string,
  overrides: Partial<Record<string, string | null>> = {},
) {
  return {
    answer_id: id,
    revision,
    form_definition_version: 1,
    received_at: "2026-10-17T00:30:00Z",
    provenance: "owner_submission",
    answers: {
      visit_purpose: `目的${id}`,
      onset: "3日前から",
      energy: "usual",
      appetite: "markedly_reduced",
      current_symptoms: null,
      improvement: "new_symptoms",
      examination_preference: "decide_with_vet",
      main_concern: null,
      ...overrides,
    },
  };
}

function body(partial: Record<string, unknown> = {}) {
  return {
    availability: "available",
    form_definition_version: 1,
    current_target: {
      ref: ref(),
      head_revision: "2",
      answer_state: "answered",
      latest_answer: answer("a2", "2"),
      content_available: true,
    },
    history: [],
    next_cursor: null,
    ...partial,
  };
}

function renderSection(props: Partial<{ appointmentId: string; clinicId: string }> = {}) {
  return render(<RevisitIntakeSection appointmentId="88" {...props} />, {
    wrapper: createTestWrapper(),
  });
}

function respond(json: unknown, status = 200) {
  server.use(http.get(PATH, () => HttpResponse.json(json, { status })));
}

describe("RevisitIntakeSection", () => {
  it("読込中はセクション内スケルトンを aria-busy で出す", async () => {
    server.use(
      http.get(PATH, async () => {
        await new Promise((r) => setTimeout(r, 50));
        return HttpResponse.json(body());
      }),
    );
    renderSection();
    expect(screen.getByRole("status")).toHaveTextContent("問診を読み込み中");
    expect(screen.getByRole("region", { name: "飼主事前問診（LINE）" })).toHaveAttribute(
      "aria-busy",
      "true",
    );
    expect(await screen.findByText("回答済み")).toBeInTheDocument();
  });

  it("target_unresolved は「対象を特定できません」を出す", async () => {
    respond(body({ availability: "target_unresolved", current_target: null }));
    renderSection();
    expect(await screen.findByText("対象を特定できません。")).toBeInTheDocument();
  });

  it("未回答は文字+アイコンのバッジと空状態 3 点（無い／条件／次の操作）を出す", async () => {
    respond(
      body({
        current_target: {
          ref: { ...ref(), target_id: null },
          head_revision: null,
          answer_state: "unanswered",
          latest_answer: null,
          content_available: false,
        },
      }),
    );
    renderSection();
    expect(await screen.findByText("未回答")).toBeInTheDocument();
    expect(screen.getByText("この予約の事前問診は届いていません。")).toBeInTheDocument();
    expect(screen.getByText(/任意提出/)).toBeInTheDocument();
    expect(screen.getByText(/受付時に聞き取り/)).toBeInTheDocument();
  });

  it("回答済みは 7 項目をラベル変換して read-only 表示し、版・受信時刻・出所を出す", async () => {
    respond(body());
    renderSection();
    expect(await screen.findByText("回答済み")).toBeInTheDocument();
    expect(screen.getByText("目的a2")).toBeInTheDocument();
    expect(screen.getByText("3日前から")).toBeInTheDocument();
    expect(screen.getByText("元気: いつも通り ／ 食欲: かなり低下")).toBeInTheDocument();
    expect(screen.getByText("新しい症状が出た")).toBeInTheDocument();
    expect(screen.getByText("獣医師と相談して決めたい")).toBeInTheDocument();
    expect(screen.getByText(/回答版 2/)).toBeInTheDocument();
    expect(screen.getByText(/2026\/10\/17 09:30/)).toBeInTheDocument();
    expect(screen.getByText(/飼主（LINE）/)).toBeInTheDocument();
    // 任意の空値は補完せず「—」
    expect(screen.getAllByText("—").length).toBeGreaterThanOrEqual(2);
    // 編集・補足入力 UI を置かない
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("未知の code は別の意味へ誤読せず「不明な回答」とする", async () => {
    respond(
      body({
        current_target: {
          ref: ref(),
          head_revision: "2",
          answer_state: "answered",
          latest_answer: answer("a2", "2", { improvement: "future_code" }),
          content_available: true,
        },
      }),
    );
    renderSection();
    expect(await screen.findByText("不明な回答")).toBeInTheDocument();
  });

  it("回答済みで content_available=false は本文を出さず注記する", async () => {
    respond(
      body({
        current_target: {
          ref: ref(),
          head_revision: "2",
          answer_state: "answered",
          latest_answer: null,
          content_available: false,
        },
      }),
    );
    renderSection();
    expect(await screen.findByText("回答済み（本文は閲覧できません）")).toBeInTheDocument();
    expect(screen.queryByText("来院目的")).not.toBeInTheDocument();
  });

  it("current の旧 revision は訂正履歴、previous は別グループで分離し最新と重複しない", async () => {
    respond(
      body({
        history: [
          { target: ref(), answer: answer("a2", "2"), target_relation: "current" },
          { target: ref(), answer: answer("a1", "1"), target_relation: "current" },
          {
            target: ref("77", "1"),
            answer: answer("p1", "1"),
            target_relation: "previous",
          },
        ],
      }),
    );
    renderSection();
    expect(await screen.findByText("目的a2")).toBeInTheDocument();
    // 最新(a2)は本体のみ。履歴側に a2 は重複しない
    expect(screen.getAllByText("目的a2")).toHaveLength(1);

    const correction = screen.getByText("訂正履歴（1件）").closest("details")!;
    const previous = screen.getByText("以前の対象ペットの回答（1件）").closest("details")!;
    expect(correction).not.toBe(previous);
    expect(within(correction).getByText("目的a1")).toBeInTheDocument();
    expect(within(correction).queryByText("目的p1")).not.toBeInTheDocument();
    expect(within(previous).getByText("目的p1")).toBeInTheDocument();
    expect(within(previous).getByText(/ペットID 77 ・ 世代 1/)).toBeInTheDocument();
    expect(within(previous).queryByText("目的a1")).not.toBeInTheDocument();
  });

  it("next_cursor があれば「さらに表示」で cursor 付きの次ページを追記する", async () => {
    const seen: (string | null)[] = [];
    server.use(
      http.get(PATH, ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        seen.push(cursor);
        return HttpResponse.json(
          cursor
            ? body({
                history: [{ target: ref(), answer: answer("a0", "1"), target_relation: "current" }],
                next_cursor: null,
              })
            : body({
                history: [{ target: ref(), answer: answer("a1", "1"), target_relation: "current" }],
                next_cursor: "CURSOR_2",
              }),
        );
      }),
    );
    renderSection();
    const more = await screen.findByRole("button", { name: "さらに表示" });
    await userEvent.click(more);
    await waitFor(() => expect(screen.getByText("訂正履歴（2件）")).toBeInTheDocument());
    expect(seen).toEqual([null, "CURSOR_2"]);
    expect(screen.queryByRole("button", { name: "さらに表示" })).not.toBeInTheDocument();
  });

  it("403 は 1 行のみで本文・件数を出さない", async () => {
    respond({ error: "forbidden", code: "FORBIDDEN" }, 403);
    renderSection();
    expect(await screen.findByText("閲覧権限がありません。")).toBeInTheDocument();
    expect(screen.queryByText("回答済み")).not.toBeInTheDocument();
    expect(screen.queryByText("未回答")).not.toBeInTheDocument();
    expect(screen.queryByText(/件/)).not.toBeInTheDocument();
  });

  it("404 はセクションを出さず 1 行のみ", async () => {
    respond({ error: "not found", code: "NOT_FOUND" }, 404);
    renderSection();
    expect(await screen.findByText("問診情報はありません。")).toBeInTheDocument();
    expect(screen.queryByRole("region")).not.toBeInTheDocument();
  });

  it("取得失敗はセクション内に再試行を出し、押下で再取得する", async () => {
    let calls = 0;
    server.use(
      http.get(PATH, () => {
        calls += 1;
        // retry:1 の自動再試行分(2 回)は失敗、手動再試行で成功
        return calls <= 2
          ? HttpResponse.json({ error: "x", code: "INTERNAL" }, { status: 500 })
          : HttpResponse.json(body());
      }),
    );
    renderSection();
    const retry = await screen.findByRole("button", { name: "再試行" });
    await userEvent.click(retry);
    expect(await screen.findByText("目的a2")).toBeInTheDocument();
  });

  it("clinicId を X-Clinic-ID ヘッダで送る", async () => {
    let header: string | null = null;
    server.use(
      http.get(PATH, ({ request }) => {
        header = request.headers.get("X-Clinic-ID");
        return HttpResponse.json(body());
      }),
    );
    renderSection({ clinicId: "2" });
    await screen.findByText("回答済み");
    expect(header).toBe("2");
  });

  it("契約外の nested 応答は fail-closed で取得失敗表示にする", async () => {
    respond(
      body({
        history: [
          {
            target: { pet_id: "10" },
            answer: answer("a1", "1"),
            target_relation: "current",
          },
        ],
      }),
    );
    renderSection();
    expect(await screen.findByText("問診を取得できませんでした。")).toBeInTheDocument();
    expect(screen.queryByText("目的a2")).not.toBeInTheDocument();
  });

  it("answered なのに content_available と latest_answer が不整合なら fail-closed", async () => {
    respond(
      body({
        current_target: {
          ref: ref(),
          head_revision: "2",
          answer_state: "answered",
          latest_answer: null,
          content_available: true,
        },
      }),
    );
    renderSection();
    expect(await screen.findByText("問診を取得できませんでした。")).toBeInTheDocument();
  });

  it("2 ページ目が 403 になったら取得済みの本文を表示しない", async () => {
    server.use(
      http.get(PATH, ({ request }) => {
        if (new URL(request.url).searchParams.get("cursor")) {
          return HttpResponse.json({ error: "forbidden", code: "FORBIDDEN" }, { status: 403 });
        }
        return HttpResponse.json(body({ next_cursor: "CURSOR_2" }));
      }),
    );
    renderSection();
    await userEvent.click(await screen.findByRole("button", { name: "さらに表示" }));
    expect(await screen.findByText("閲覧権限がありません。")).toBeInTheDocument();
    expect(screen.queryByText("目的a2")).not.toBeInTheDocument();
  });

  it("ページ間で current_target.ref が変わったら後続ページを merge せず再読み込みを促す", async () => {
    server.use(
      http.get(PATH, ({ request }) => {
        if (new URL(request.url).searchParams.get("cursor")) {
          return HttpResponse.json(
            body({
              current_target: {
                ref: ref("99", "3"),
                head_revision: "1",
                answer_state: "answered",
                latest_answer: answer("z1", "1"),
                content_available: true,
              },
              history: [
                { target: ref("99", "3"), answer: answer("z0", "1"), target_relation: "current" },
              ],
            }),
          );
        }
        return HttpResponse.json(body({ next_cursor: "CURSOR_2" }));
      }),
    );
    renderSection();
    await userEvent.click(await screen.findByRole("button", { name: "さらに表示" }));
    expect(
      await screen.findByText("対象または回答状態が変わりました。再読み込みしてください。"),
    ).toBeInTheDocument();
    expect(screen.queryByText("目的a2")).not.toBeInTheDocument();
    expect(screen.queryByText("目的z0")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "再読み込み" })).toBeInTheDocument();
  });

  it("ref が同じでも後続ページで回答状態が変わったら merge せず再読み込みを促す", async () => {
    server.use(
      http.get(PATH, ({ request }) => {
        if (new URL(request.url).searchParams.get("cursor")) {
          return HttpResponse.json(
            body({
              current_target: {
                ref: ref(),
                head_revision: "3",
                answer_state: "answered",
                latest_answer: answer("a3", "3"),
                content_available: true,
              },
              history: [
                { target: ref(), answer: answer("a3", "3"), target_relation: "current" },
              ],
            }),
          );
        }
        return HttpResponse.json(body({ next_cursor: "CURSOR_2" }));
      }),
    );
    renderSection();
    await userEvent.click(await screen.findByRole("button", { name: "さらに表示" }));
    expect(
      await screen.findByText("対象または回答状態が変わりました。再読み込みしてください。"),
    ).toBeInTheDocument();
    expect(screen.queryByText("目的a2")).not.toBeInTheDocument();
    expect(screen.queryByText("目的a3")).not.toBeInTheDocument();
  });
});

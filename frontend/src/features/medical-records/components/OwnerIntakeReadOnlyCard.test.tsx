import type { ComponentProps } from "react";
import { afterEach, describe, expect, it } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { OwnerIntakeReadOnlyCard } from "./OwnerIntakeReadOnlyCard";

// EMR-477: route は runtime 未実装（EMR-476 / EMR-472 待ち）。契約 backend/docs/api.yaml の
// RevisitIntakeRecordRead 形を、合成値のテスト MSW handler で代替する。

const ROUTE = "/api/v1/medical-records/:recordId/revisit-intake";

const REF = {
  target_id: "501",
  appointment_id: "301",
  pet_id: "11",
  target_generation: "1",
};

function answer(revision: string, overrides: Record<string, unknown> = {}) {
  return {
    answer_id: `9${revision}`,
    revision,
    form_definition_version: 1,
    received_at: "2026-10-18T01:30:00Z", // JST 10:30
    provenance: "owner_submission",
    answers: {
      visit_purpose: `合成の来院目的 v${revision}`,
      onset: "3日前から",
      energy: "slightly_reduced",
      appetite: "markedly_reduced",
      current_symptoms: "合成の症状\n2行目",
      improvement: "new_symptoms",
      examination_preference: "decide_with_vet",
      main_concern: null,
      ...overrides,
    },
  };
}

function boundBody(overrides: Record<string, unknown> = {}) {
  return {
    binding: "bound",
    target: {
      ref: REF,
      head_revision: "2",
      answer_state: "answered",
      latest_answer: answer("2"),
      content_available: true,
    },
    history: [
      { target: REF, answer: answer("2"), target_relation: "current" },
      { target: REF, answer: answer("1", { visit_purpose: "合成の旧版目的" }), target_relation: "current" },
    ],
    next_cursor: null,
    ...overrides,
  };
}

function useIntake(body: unknown, status = 200) {
  server.use(http.get(ROUTE, () => HttpResponse.json(body as object, { status })));
}

function renderCard(props: Partial<ComponentProps<typeof OwnerIntakeReadOnlyCard>> = {}) {
  return render(
    <OwnerIntakeReadOnlyCard
      recordId="10"
      recordClinicId="7"
      petId="11"
      isNewRecord={false}
      {...props}
    />,
    { wrapper: createTestWrapper() },
  );
}

afterEach(() => {
  server.resetHandlers();
});

describe("OwnerIntakeReadOnlyCard — bound・回答あり", () => {
  it("7 項目・回答版・受信時刻・出所を読み取り専用で表示する", async () => {
    useIntake(boundBody({ history: [] }));
    renderCard();

    const card = await screen.findByTestId("owner-intake-card");
    expect(within(card).getByRole("heading", { name: "飼主事前問診（LINE・読み取り専用）" })).toBeInTheDocument();
    expect(within(card).getByText("合成の来院目的 v2")).toBeInTheDocument();
    expect(within(card).getByText("3日前から")).toBeInTheDocument();
    expect(within(card).getByText("元気: 少し低下 / 食欲: かなり低下")).toBeInTheDocument();
    expect(within(card).getByText(/合成の症状/)).toBeInTheDocument();
    expect(within(card).getByText("新しい症状が出た")).toBeInTheDocument();
    expect(within(card).getByText("獣医師と相談して決めたい")).toBeInTheDocument();
    // 任意値 null は未回答
    expect(within(card).getByText("未回答")).toBeInTheDocument();
    expect(within(card).getByText(/出所: 飼主（LINE）/)).toHaveTextContent("第2版");
    expect(within(card).getByText(/出所: 飼主（LINE）/)).toHaveTextContent("2026/10/18 10:30");
  });

  it("書き込み操作（入力欄・保存・複写・クリア）を一切持たない", async () => {
    useIntake(boundBody());
    renderCard();

    const card = await screen.findByTestId("owner-intake-card");
    expect(within(card).queryAllByRole("textbox")).toHaveLength(0);
    expect(within(card).queryAllByRole("combobox")).toHaveLength(0);
    expect(card.querySelectorAll("input, textarea, select")).toHaveLength(0);
    expect(within(card).queryByRole("button", { name: /保存|コピー|複写|クリア|元に戻す|編集/ })).toBeNull();
  });

  it("未知 code は最初の選択肢へ補完せず契約失敗として扱う", async () => {
    useIntake(
      boundBody({
        target: {
          ref: REF,
          head_revision: "1",
          answer_state: "answered",
          latest_answer: answer("1", { energy: "future_code" }),
          content_available: true,
        },
        history: [],
      }),
    );
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText(/元気: いつも通り/)).toBeNull();
    expect(screen.queryByText("未回答")).toBeNull();
  });

  it("X-Clinic-ID に record 自身の clinicId を送る", async () => {
    let clinicHeader: string | null = null;
    server.use(
      http.get(ROUTE, ({ request }) => {
        clinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json(boundBody());
      }),
    );
    renderCard({ recordClinicId: "42" });

    await screen.findByTestId("owner-intake-card");
    expect(clinicHeader).toBe("42");
  });
});

describe("OwnerIntakeReadOnlyCard — bound・その他の状態", () => {
  it("未回答は「未回答」バッジと空状態 3 点を出す", async () => {
    useIntake({
      binding: "bound",
      target: {
        ref: REF,
        head_revision: null,
        answer_state: "unanswered",
        latest_answer: null,
        content_available: false,
      },
      history: [],
      next_cursor: null,
    });
    renderCard();

    expect(await screen.findByText("未回答")).toBeInTheDocument();
    expect(screen.getByText("この予約の飼主事前問診の回答はありません。")).toBeInTheDocument();
    expect(screen.getByText(/受付後は飼主からの新しい送信はできません/)).toBeInTheDocument();
    expect(screen.getByText(/スタッフ用の問診欄に記録してください/)).toBeInTheDocument();
  });

  it("content_available=false は本文を出さず「回答済み（本文は閲覧できません）」", async () => {
    useIntake({
      binding: "bound",
      target: {
        ref: REF,
        head_revision: "3",
        answer_state: "answered",
        latest_answer: null,
        content_available: false,
      },
      history: [],
      next_cursor: null,
    });
    renderCard();

    expect(await screen.findByText("回答済み（本文は閲覧できません）")).toBeInTheDocument();
    expect(screen.queryByText("未回答")).toBeNull();
    expect(screen.queryByText("来院目的")).toBeNull();
  });

  it("unbound は未回答と文言を分けて「対応する予約回答を特定できません」", async () => {
    useIntake({ binding: "unbound", target: null, history: [], next_cursor: null });
    renderCard();

    expect(await screen.findByText("対応する予約回答を特定できません")).toBeInTheDocument();
    expect(screen.queryByText("未回答")).toBeNull();
    expect(screen.queryByText("この予約の飼主事前問診の回答はありません。")).toBeNull();
  });

  it("mismatch は「出所不明」で本文を出さない（応答に本文が混入しても採用しない）", async () => {
    useIntake({
      binding: "mismatch",
      target: null,
      history: [{ target: REF, answer: answer("1", { visit_purpose: "混入した本文" }), target_relation: "current" }],
      next_cursor: null,
    });
    renderCard();

    expect(await screen.findByText("出所不明")).toBeInTheDocument();
    expect(screen.queryByText("混入した本文")).toBeNull();
    expect(screen.queryByText("来院目的")).toBeNull();
  });
});

describe("OwnerIntakeReadOnlyCard — 取得結果の分岐", () => {
  it("新規カルテは非表示で取得もしない", async () => {
    let called = 0;
    server.use(
      http.get(ROUTE, () => {
        called += 1;
        return HttpResponse.json(boundBody());
      }),
    );
    const { container } = renderCard({ isNewRecord: true, recordId: undefined });

    expect(container).toBeEmptyDOMElement();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(called).toBe(0);
  });

  it("403 はカードを出さず 1 行注記のみ", async () => {
    useIntake({ error: "forbidden", code: "FORBIDDEN", details: null }, 403);
    renderCard();

    expect(await screen.findByTestId("owner-intake-forbidden-note")).toBeInTheDocument();
    expect(screen.queryByTestId("owner-intake-card")).toBeNull();
  });

  it("404 は非表示", async () => {
    useIntake({ error: "not found", code: "NOT_FOUND", details: null }, 404);
    const { container } = renderCard();

    await waitFor(() => expect(container).toBeEmptyDOMElement());
  });

  it("取得失敗はカード内の再試行で回復できる", async () => {
    let calls = 0;
    server.use(
      http.get(ROUTE, () => {
        calls += 1;
        if (calls <= 1) {
          return HttpResponse.json({ error: "x", code: "INTERNAL", details: null }, { status: 500 });
        }
        return HttpResponse.json(boundBody());
      }),
    );
    renderCard();

    const retry = await screen.findByRole("button", { name: "再試行" });
    expect(screen.getByRole("alert")).toHaveTextContent("飼主事前問診を取得できませんでした。");
    await userEvent.click(retry);

    expect(await screen.findByText("合成の来院目的 v2")).toBeInTheDocument();
  });

  it("契約外の形は本文を推測表示せず取得失敗として扱う", async () => {
    useIntake({ binding: "bound", target: null, history: [] });
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
  });
});

describe("OwnerIntakeReadOnlyCard — 履歴（固定 target 内の旧 revision）", () => {
  it("旧版を折りたたみで出し、現在版を重複表示しない", async () => {
    useIntake(boundBody());
    renderCard();

    await screen.findByText("合成の来院目的 v2");
    const entries = screen.getAllByTestId("owner-intake-history-entry");
    expect(entries).toHaveLength(1);
    expect(within(entries[0]).getByText("合成の旧版目的")).toBeInTheDocument();
    expect(screen.getByText(/過去の回答版（1件）/)).toBeInTheDocument();
  });

  it("別 target の履歴行は表示しない", async () => {
    useIntake(
      boundBody({
        history: [
          {
            target: { ...REF, target_id: "999", target_generation: "9" },
            answer: answer("1", { visit_purpose: "別targetの本文" }),
            target_relation: "previous",
          },
        ],
      }),
    );
    renderCard();

    await screen.findByText("合成の来院目的 v2");
    expect(screen.queryByText("別targetの本文")).toBeNull();
  });

  it("next_cursor があれば「さらに表示」で次ページを cursor 付きで取得する", async () => {
    const cursors: Array<string | null> = [];
    server.use(
      http.get(ROUTE, ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        cursors.push(cursor);
        if (cursor === "cur-2") {
          return HttpResponse.json(
            boundBody({
              history: [
                { target: REF, answer: { ...answer("1", { visit_purpose: "合成の次ページ目的" }), answer_id: "95" }, target_relation: "current" },
              ],
              next_cursor: null,
            }),
          );
        }
        return HttpResponse.json(boundBody({ next_cursor: "cur-2" }));
      }),
    );
    renderCard();

    await screen.findByText("合成の来院目的 v2");
    const more = screen.getByRole("button", { name: "さらに表示" });
    await userEvent.click(more);

    expect(await screen.findByText("合成の次ページ目的")).toBeInTheDocument();
    expect(cursors).toEqual([null, "cur-2"]);
    expect(screen.queryByRole("button", { name: "さらに表示" })).toBeNull();
  });
});

describe("OwnerIntakeReadOnlyCard — 契約の厳格検証と部分失敗", () => {
  it("未回答なのに latest_answer がある不整合は契約失敗（redacted に正規化しない）", async () => {
    useIntake(
      boundBody({
        target: {
          ref: REF,
          head_revision: null,
          answer_state: "unanswered",
          latest_answer: answer("1"),
          content_available: false,
        },
        history: [],
      }),
    );
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText("回答済み（本文は閲覧できません）")).toBeNull();
  });

  it("answered・content_available=true で latest_answer が無い場合は契約失敗", async () => {
    useIntake(
      boundBody({
        target: {
          ref: REF,
          head_revision: "2",
          answer_state: "answered",
          latest_answer: null,
          content_available: true,
        },
        history: [],
      }),
    );
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText("回答済み（本文は閲覧できません）")).toBeNull();
  });

  it("target_relation が契約外の履歴行は契約失敗", async () => {
    useIntake(
      boundBody({ history: [{ target: REF, answer: answer("1"), target_relation: "other" }] }),
    );
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("target_id が欠落した履歴の target は契約失敗", async () => {
    const { target_id: _omit, ...noId } = REF;
    void _omit;
    useIntake(
      boundBody({ history: [{ target: noId, answer: answer("1"), target_relation: "current" }] }),
    );
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("本文不可でも行認可済みの履歴と「さらに表示」は描画する", async () => {
    useIntake({
      binding: "bound",
      target: {
        ref: REF,
        head_revision: "3",
        answer_state: "answered",
        latest_answer: null,
        content_available: false,
      },
      history: [
        { target: REF, answer: answer("1", { visit_purpose: "合成の閲覧可能な旧版" }), target_relation: "current" },
      ],
      next_cursor: "cur-2",
    });
    renderCard();

    expect(await screen.findByText("回答済み（本文は閲覧できません）")).toBeInTheDocument();
    expect(screen.getByText("合成の閲覧可能な旧版")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "さらに表示" })).toBeInTheDocument();
  });

  it("次ページ取得に失敗しても取得済みの本文を保持し inline で再試行できる", async () => {
    let calls = 0;
    server.use(
      http.get(ROUTE, ({ request }) => {
        calls += 1;
        const cursor = new URL(request.url).searchParams.get("cursor");
        if (cursor === "cur-2" && calls <= 2) {
          return HttpResponse.json({ error: "x", code: "INTERNAL", details: null }, { status: 500 });
        }
        if (cursor === "cur-2") {
          return HttpResponse.json(
            boundBody({
              history: [
                { target: REF, answer: { ...answer("1", { visit_purpose: "合成の次ページ目的" }), answer_id: "95" }, target_relation: "current" },
              ],
              next_cursor: null,
            }),
          );
        }
        return HttpResponse.json(boundBody({ next_cursor: "cur-2" }));
      }),
    );
    renderCard();

    await screen.findByText("合成の来院目的 v2");
    await userEvent.click(screen.getByRole("button", { name: "さらに表示" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("最新の状態を取得できませんでした。");
    expect(screen.getByText("合成の来院目的 v2")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "再試行" }));
    expect(await screen.findByText("合成の次ページ目的")).toBeInTheDocument();
  });

  it("任意項目の key 欠落・型違いは「未回答」にせず契約失敗", async () => {
    const missing = answer("1");
    delete (missing.answers as Record<string, unknown>).onset;
    useIntake(
      boundBody({
        target: {
          ref: REF,
          head_revision: "1",
          answer_state: "answered",
          latest_answer: missing,
          content_available: true,
        },
        history: [],
      }),
    );
    const { unmount } = renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    unmount();

    useIntake(
      boundBody({
        target: {
          ref: REF,
          head_revision: "1",
          answer_state: "answered",
          latest_answer: answer("1", { main_concern: 123 }),
          content_available: true,
        },
        history: [],
      }),
    );
    renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("予約側 target が変わり relation が previous でも、固定 target 内の旧版は表示する", async () => {
    useIntake(
      boundBody({
        history: [
          {
            target: REF,
            answer: answer("1", { visit_purpose: "合成の固定target旧版" }),
            target_relation: "previous",
          },
        ],
      }),
    );
    renderCard();

    await screen.findByText("合成の来院目的 v2");
    expect(screen.getByText("合成の固定target旧版")).toBeInTheDocument();
  });

  it("応答の pet が呼び出し患者と違えば（target / 履歴いずれも）契約失敗", async () => {
    useIntake(boundBody({ history: [] }));
    const first = renderCard({ petId: "12" });
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText("合成の来院目的 v2")).toBeNull();
    first.unmount();

    useIntake(
      boundBody({
        history: [
          {
            target: { ...REF, pet_id: "12" },
            answer: answer("1"),
            target_relation: "current",
          },
        ],
      }),
    );
    renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText("合成の来院目的 v2")).toBeNull();
  });

  it("visit_purpose が空文字・Unicode 空白のみなら契約失敗", async () => {
    const withAnswer = (latest: unknown) =>
      boundBody({
        target: {
          ref: REF,
          head_revision: "1",
          answer_state: "answered",
          latest_answer: latest,
          content_available: true,
        },
        history: [],
      });
    useIntake(withAnswer(answer("1", { visit_purpose: "\u3000 \n" })));
    const first = renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    first.unmount();

    useIntake(withAnswer(answer("1", { visit_purpose: "" })));
    renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("本文ありの latest_answer.revision が head_revision と違えば契約失敗", async () => {
    useIntake(
      boundBody({
        target: {
          ref: REF,
          head_revision: "3",
          answer_state: "answered",
          latest_answer: answer("2"),
          content_available: true,
        },
        history: [],
      }),
    );
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText("合成の来院目的 v2")).toBeNull();
  });

  it("next_cursor の欠落・非 string は契約失敗", async () => {
    const body = boundBody({ history: [] }) as Record<string, unknown>;
    delete body.next_cursor;
    useIntake(body);
    const first = renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    first.unmount();

    useIntake(boundBody({ history: [], next_cursor: 5 }));
    renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("form_definition_version が 1 以外（v2 等）は契約失敗", async () => {
    useIntake(
      boundBody({
        target: {
          ref: REF,
          head_revision: "2",
          answer_state: "answered",
          latest_answer: { ...answer("2"), form_definition_version: 2 },
          content_available: true,
        },
        history: [],
      }),
    );
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText("合成の来院目的 v2")).toBeNull();
  });

  it("2 ページ目が mismatch を返したら 1 ページ目の本文を出し続けない", async () => {
    server.use(
      http.get(ROUTE, ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        if (cursor === "cur-2") {
          return HttpResponse.json({ binding: "mismatch", target: null, history: [], next_cursor: null });
        }
        return HttpResponse.json(boundBody({ next_cursor: "cur-2" }));
      }),
    );
    renderCard();

    await screen.findByText("合成の来院目的 v2");
    await userEvent.click(screen.getByRole("button", { name: "さらに表示" }));

    expect(await screen.findByText("出所不明")).toBeInTheDocument();
    expect(screen.queryByText("合成の来院目的 v2")).toBeNull();
  });

  it("ページ間で target.ref が変わったら本文を出さず再読込へ誘導する", async () => {
    server.use(
      http.get(ROUTE, ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        if (cursor === "cur-2") {
          const other = { ...REF, target_generation: "2" };
          return HttpResponse.json({
            ...boundBody({ history: [] }),
            target: { ...boundBody().target, ref: other },
            next_cursor: null,
          });
        }
        return HttpResponse.json(boundBody({ next_cursor: "cur-2" }));
      }),
    );
    renderCard();

    await screen.findByText("合成の来院目的 v2");
    await userEvent.click(screen.getByRole("button", { name: "さらに表示" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("対象または回答状態が変わりました");
    expect(screen.queryByText("合成の来院目的 v2")).toBeNull();
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("received_at が offset 無し・不正な日時なら契約失敗", async () => {
    const bad = (receivedAt: string) =>
      boundBody({
        target: {
          ref: REF,
          head_revision: "1",
          answer_state: "answered",
          latest_answer: { ...answer("1"), received_at: receivedAt },
          content_available: true,
        },
        history: [],
      });
    useIntake(bad("2026-10-18T10:30:00"));
    const first = renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    first.unmount();

    useIntake(bad("2026-13-45T10:30:00+09:00"));
    const second = renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    second.unmount();

    // 暦上存在しない日・RFC3339 範囲外の時は Date.parse が正規化して通してしまうため明示拒否。
    useIntake(bad("2026-02-30T10:30:00+09:00"));
    const third = renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    third.unmount();

    useIntake(bad("2026-01-01T24:00:00Z"));
    renderCard();
    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("閏日の offset 付き received_at（UTC では前日）は受理し JST で表示する", async () => {
    useIntake(
      boundBody({
        target: {
          ref: REF,
          head_revision: "1",
          answer_state: "answered",
          latest_answer: { ...answer("1"), received_at: "2024-02-29T00:00:00+09:00" },
          content_available: true,
        },
        history: [],
      }),
    );
    renderCard();

    expect(await screen.findByText(/受信: 2024\/02\/29 00:00/)).toBeInTheDocument();
  });

  it("読込中は aria-busy の skeleton（role=status）を出す", async () => {
    server.use(http.get(ROUTE, () => new Promise(() => {})));
    renderCard();

    const card = await screen.findByTestId("owner-intake-card");
    expect(card).toHaveAttribute("aria-busy", "true");
    expect(within(card).getByRole("status")).toHaveTextContent("飼主事前問診を読み込み中");
    expect(card.querySelectorAll(".animate-pulse").length).toBeGreaterThan(0);
  });

  it.each([
    ["空文字", ""],
    ["0", "0"],
    ["先頭ゼロ", "01"],
    ["20 桁", "12345678901234567890"],
    ["BIGINT 超 19 桁", "9223372036854775808"],
  ])("ID・revision が正規形十進文字列でない（%s）なら契約失敗", async (_label, bad) => {
    useIntake(
      boundBody({
        target: {
          ref: { ...REF, target_id: bad },
          head_revision: "2",
          answer_state: "answered",
          latest_answer: answer("2"),
          content_available: true,
        },
        history: [],
      }),
    );
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText("合成の来院目的 v2")).toBeNull();
  });

  it("answer_id・revision・head_revision も正規形十進文字列を要求する", async () => {
    for (const latest of [
      { ...answer("2"), answer_id: "ans-2" },
      { ...answer("2"), revision: "02" },
    ]) {
      useIntake(
        boundBody({
          target: {
            ref: REF,
            head_revision: latest.revision,
            answer_state: "answered",
            latest_answer: latest,
            content_available: true,
          },
          history: [],
        }),
      );
      const view = renderCard();
      expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
      view.unmount();
    }
  });

  it("BIGINT 上限ちょうどの ID は受理する", async () => {
    useIntake(
      boundBody({
        target: {
          ref: { ...REF, target_id: "9223372036854775807" },
          head_revision: "2",
          answer_state: "answered",
          latest_answer: answer("2"),
          content_available: true,
        },
        history: [],
      }),
    );
    renderCard();

    expect(await screen.findByText("合成の来院目的 v2")).toBeInTheDocument();
  });

  it("同じ ref でも後続ページで回答状態が変わったら本文を出さず再読込へ誘導する", async () => {
    server.use(
      http.get(ROUTE, ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        if (cursor === "cur-2") {
          return HttpResponse.json({
            binding: "bound",
            target: {
              ref: REF,
              head_revision: "2",
              answer_state: "answered",
              latest_answer: null,
              content_available: false,
            },
            history: [],
            next_cursor: null,
          });
        }
        return HttpResponse.json(boundBody({ next_cursor: "cur-2" }));
      }),
    );
    renderCard();

    await screen.findByText("合成の来院目的 v2");
    await userEvent.click(screen.getByRole("button", { name: "さらに表示" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("対象または回答状態が変わりました");
    expect(screen.queryByText("合成の来院目的 v2")).toBeNull();
  });

  it("同 ref の履歴 revision が head_revision を超える応答は契約失敗（桁数比較）", async () => {
    // "10" > "9" を辞書順で誤判定しない（head=9, history=10）。
    useIntake(
      boundBody({
        target: {
          ref: REF,
          head_revision: "9",
          answer_state: "answered",
          latest_answer: answer("9"),
          content_available: true,
        },
        history: [{ target: REF, answer: answer("10"), target_relation: "current" }],
      }),
    );
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText("合成の来院目的 v9")).toBeNull();
  });

  it("未回答（head=null）なのに同 ref の履歴行がある応答は契約失敗", async () => {
    useIntake({
      binding: "bound",
      target: {
        ref: REF,
        head_revision: null,
        answer_state: "unanswered",
        latest_answer: null,
        content_available: false,
      },
      history: [{ target: REF, answer: answer("1"), target_relation: "current" }],
      next_cursor: null,
    });
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.queryByText("この予約の飼主事前問診の回答はありません。")).toBeNull();
  });

  it("next_cursor が空文字なら契約失敗", async () => {
    useIntake(boundBody({ history: [], next_cursor: "" }));
    renderCard();

    expect(await screen.findByRole("button", { name: "再試行" })).toBeInTheDocument();
  });
});

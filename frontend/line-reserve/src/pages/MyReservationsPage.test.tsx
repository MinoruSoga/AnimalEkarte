import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { MyReservationsPage } from "./MyReservationsPage";
import type { Reservation } from "../types/models";

const BASE_PROPS = {
  clinicId: "1",
  idToken: "test-id-token",
};

const confirmedReservation: Reservation = {
  id: 1,
  course_name: "一般診察",
  staff_name: "山田先生",
  date: "2026-08-01",
  start_time: "10:00",
  end_time: "10:30",
  status: "confirmed",
  notes: "R-20260801-0001",
  created_at: "2026-07-01T00:00:00Z",
};

describe("MyReservationsPage（R-F25: ネイティブダイアログ→インラインUI置換）", () => {
  it("予約一覧を表示する", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    expect(await screen.findByText("一般診察")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "キャンセルする" })).toBeInTheDocument();
  });

  it("「キャンセルする」クリックでネイティブダイアログではなくインライン確認UIを表示する", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "キャンセルする" }));

    expect(screen.getByText("本当にキャンセルしますか？")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "はい、キャンセルする" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "いいえ" })).toBeInTheDocument();
  });

  it("確認UIで「いいえ」を選ぶとキャンセル操作を中断し、元のボタンに戻る", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "キャンセルする" }));
    await user.click(screen.getByRole("button", { name: "いいえ" }));

    expect(screen.queryByText("本当にキャンセルしますか？")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "キャンセルする" })).toBeInTheDocument();
  });

  it("確認UIで「はい」を選び成功すると予約がキャンセル済に更新される", async () => {
    const user = userEvent.setup();
    let deleteCalled = false;
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
      http.delete("/api/liff/:clinicId/my-reservations/:id", () => {
        deleteCalled = true;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "キャンセルする" }));
    await user.click(screen.getByRole("button", { name: "はい、キャンセルする" }));

    await waitFor(() => {
      expect(deleteCalled).toBe(true);
    });
    expect(await screen.findByText("キャンセル済")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "キャンセルする" })).not.toBeInTheDocument();
  });

  it("確認UIで「はい」を選び失敗するとネイティブalertではなくインラインバナーを表示し、再試行可能にする", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
      http.delete("/api/liff/:clinicId/my-reservations/:id", () =>
        HttpResponse.json(null, { status: 500 }),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "キャンセルする" }));
    await user.click(screen.getByRole("button", { name: "はい、キャンセルする" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("キャンセルに失敗しました");
    expect(screen.getByRole("button", { name: "キャンセルする" })).not.toBeDisabled();
  });

  it("一覧取得API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する（R-F22/R-F23）", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json(null, { status: 500 }),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    expect(await screen.findByRole("alert")).toHaveTextContent("サーバーエラーが発生しました");
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("一覧取得API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない（R-F22/R-F23）", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json(null, { status: 401 }),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "LINEアプリを再起動して開き直してください",
    );
    expect(screen.queryByRole("button", { name: "再試行" })).not.toBeInTheDocument();
  });

  it("一覧の再試行ボタンをクリックすると予約一覧を再取得する（R-F22/R-F23）", async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json([confirmedReservation]);
      }),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "再試行" }));

    expect(await screen.findByText("一般診察")).toBeInTheDocument();
    expect(callCount).toBe(2);
  });
});

describe("MyReservationsPage 再診問診の導線（EMR-474）", () => {
  const cancelledReservation: Reservation = {
    ...confirmedReservation,
    id: 2,
    status: "cancelled",
    notes: "R-20260802-0002",
  };

  function revisitRead(answered: boolean, historyCount: number) {
    const ref = { target_id: "5401", appointment_id: "2", pet_id: "801", target_generation: "1" };
    const ans = {
      answer_id: "9101",
      revision: "1",
      form_definition_version: 1,
      received_at: "2026-07-20T01:00:00Z",
      provenance: "owner_submission",
      answers: {
        visit_purpose: "合成の来院目的",
        onset: null,
        energy: null,
        appetite: null,
        current_symptoms: null,
        improvement: null,
        examination_preference: null,
        main_concern: null,
      },
    };
    return {
      availability: "available",
      form_definition_version: 1,
      current_target: answered
        ? {
            ref,
            head_revision: "1",
            answer_state: "answered",
            latest_answer: ans,
            content_available: true,
            editable: false,
            edit_block_reason: "cancelled",
          }
        : {
            ref: { ...ref, target_id: null },
            head_revision: null,
            answer_state: "unanswered",
            latest_answer: null,
            content_available: false,
            editable: false,
            edit_block_reason: "cancelled",
          },
      history: Array.from({ length: historyCount }, () => ({
        target: ref,
        answer: ans,
        target_relation: "current",
      })),
      next_cursor: null,
    };
  }

  it("confirmed カードに「問診を回答する／訂正する」を出し、押すとその予約で問診を開く", async () => {
    const user = userEvent.setup();
    const onOpenRevisitIntake = vi.fn();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
    );

    render(
      <MyReservationsPage
        {...BASE_PROPS}
        onBack={vi.fn()}
        onOpenRevisitIntake={onOpenRevisitIntake}
      />,
    );

    await user.click(await screen.findByRole("button", { name: "問診を回答する／訂正する" }));
    expect(onOpenRevisitIntake).toHaveBeenCalledWith(
      expect.objectContaining({ id: 1, status: "confirmed" }),
    );
  });

  it("取消済みカードは回答履歴があるときだけ「回答履歴を見る」を出す", async () => {
    const user = userEvent.setup();
    const onOpenRevisitIntake = vi.fn();
    const requested: URL[] = [];
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([cancelledReservation]),
      ),
      http.get("/api/liff/:clinicId/my-reservations/:id/revisit-intake", ({ request }) => {
        requested.push(new URL(request.url));
        return HttpResponse.json(revisitRead(true, 1));
      }),
    );

    render(
      <MyReservationsPage
        {...BASE_PROPS}
        onBack={vi.fn()}
        onOpenRevisitIntake={onOpenRevisitIntake}
      />,
    );

    await user.click(await screen.findByRole("button", { name: "回答履歴を見る" }));
    expect(onOpenRevisitIntake).toHaveBeenCalledWith(expect.objectContaining({ id: 2 }));
    expect(requested[0].searchParams.get("limit")).toBe("1");
    expect(
      screen.queryByRole("button", { name: "問診を回答する／訂正する" }),
    ).not.toBeInTheDocument();
  });

  it("取消済みで回答履歴が無い場合と、取得に失敗した場合は導線を出さない", async () => {
    const cancelledWithoutHistory: Reservation = { ...cancelledReservation, id: 3 };
    let resolved = 0;
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([cancelledReservation, cancelledWithoutHistory]),
      ),
      http.get("/api/liff/:clinicId/my-reservations/:id/revisit-intake", ({ params }) => {
        resolved += 1;
        if (params.id === "2") return HttpResponse.json({ error: "not found" }, { status: 404 });
        return HttpResponse.json(revisitRead(false, 0));
      }),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} onOpenRevisitIntake={vi.fn()} />);

    await screen.findAllByText("キャンセル済");
    await waitFor(() => expect(resolved).toBe(2));
    expect(screen.queryByRole("button", { name: "回答履歴を見る" })).not.toBeInTheDocument();
  });

  it("onOpenRevisitIntake 未指定なら問診の導線を出さず、問診 API も呼ばない", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation, cancelledReservation]),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    expect(await screen.findByText("キャンセル済")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /問診/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "回答履歴を見る" })).not.toBeInTheDocument();
  });
});

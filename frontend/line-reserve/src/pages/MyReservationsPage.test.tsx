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

// 2^53: 正確に表現できるが Number.isSafeInteger では false（unsafe 数値の代表）
const UNSAFE_RESERVATION_ID_NUMBER = Number.MAX_SAFE_INTEGER + 1;

const confirmedReservation: Reservation = {
  id: "1",
  pet_id: null,
  pet_name: null,
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
    id: "2",
    status: "cancelled",
    notes: "R-20260802-0002",
  };

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
      expect.objectContaining({ id: "1", status: "confirmed" }),
    );
  });

  it("pending カードにも問診の導線を出し、押すとその予約で問診を開く（編集可否は問診 GET が決める）", async () => {
    const user = userEvent.setup();
    const onOpenRevisitIntake = vi.fn();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([{ ...confirmedReservation, id: "3", status: "pending" }]),
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
      expect.objectContaining({ id: "3", status: "pending" }),
    );
  });

  it.each(["checked_in", "in_consultation", "accounting", "completed"] as const)(
    "%s のカードには閲覧の入口「回答を見る」を出す（書き込みはサーバーの editable=false が止める）",
    async (status) => {
      const user = userEvent.setup();
      const onOpenRevisitIntake = vi.fn();
      server.use(
        http.get("/api/liff/:clinicId/my-reservations", () =>
          HttpResponse.json([{ ...confirmedReservation, id: "4", status }]),
        ),
      );

      render(
        <MyReservationsPage
          {...BASE_PROPS}
          onBack={vi.fn()}
          onOpenRevisitIntake={onOpenRevisitIntake}
        />,
      );

      await user.click(await screen.findByRole("button", { name: "回答を見る" }));
      expect(onOpenRevisitIntake).toHaveBeenCalledWith(
        expect.objectContaining({ id: "4", status }),
      );
    },
  );

  it("no_show でも導線を出し、FE では閲覧専用に固定しない（サーバー判定に従う）", async () => {
    const user = userEvent.setup();
    const onOpenRevisitIntake = vi.fn();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([{ ...confirmedReservation, id: "5", status: "no_show" }]),
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
    expect(onOpenRevisitIntake).toHaveBeenCalledWith(expect.objectContaining({ id: "5" }));
  });

  it("int64 範囲内の大きな十進文字列 ID は丸めず問診導線を出す（9007199254740992/993 は衝突しない）", async () => {
    const user = userEvent.setup();
    const onOpenRevisitIntake = vi.fn();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([
          { ...confirmedReservation, id: "9007199254740992" },
          { ...confirmedReservation, id: "9007199254740993" },
        ]),
      ),
    );

    render(
      <MyReservationsPage
        {...BASE_PROPS}
        onBack={vi.fn()}
        onOpenRevisitIntake={onOpenRevisitIntake}
      />,
    );

    const buttons = await screen.findAllByRole("button", { name: "問診を回答する／訂正する" });
    expect(buttons).toHaveLength(2);
    await user.click(buttons[0]);
    await user.click(buttons[1]);
    expect(onOpenRevisitIntake).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({ id: "9007199254740992" }),
    );
    expect(onOpenRevisitIntake).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({ id: "9007199254740993" }),
    );
  });

  it("一覧応答の id が十進文字列でない（数値・非正規形）は読込エラーとして再試行導線を出す", async () => {
    let revisitCalls = 0;
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([{ ...confirmedReservation, id: UNSAFE_RESERVATION_ID_NUMBER }]),
      ),
      http.get("/api/liff/:clinicId/my-reservations/:id/revisit-intake", () => {
        revisitCalls += 1;
        return HttpResponse.json({});
      }),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(revisitCalls).toBe(0);
  });

  it("取消済みカードには問診の導線を出さず、一覧表示だけで問診 API を呼ばない", async () => {
    let revisitCalls = 0;
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([cancelledReservation]),
      ),
      http.get("/api/liff/:clinicId/my-reservations/:id/revisit-intake", () => {
        revisitCalls += 1;
        return HttpResponse.json({});
      }),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} onOpenRevisitIntake={vi.fn()} />);

    expect(await screen.findByText("キャンセル済")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /問診|回答履歴/ })).not.toBeInTheDocument();
    expect(revisitCalls).toBe(0);
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

describe("MyReservationsPage 取消後の送信結果照会", () => {
  const cancelledReservation: Reservation = {
    ...confirmedReservation,
    id: "2",
    status: "cancelled",
    notes: "R-20260802-0002",
  };
  const pendingStore = (has: boolean) => ({
    get: vi.fn(),
    set: vi.fn(),
    clear: vi.fn(),
    hasDraft: vi.fn(() => false),
    findPending: vi.fn((clinicId: string, appointmentId: string) =>
      has && clinicId === "1" && appointmentId === cancelledReservation.id
        ? ({ targetKey: "k", pending: {} } as never)
        : undefined,
    ),
  });

  it("取消済みでも未確定の送信がある予約には結果照会の入口だけを出し、通常の問診導線は出さない", async () => {
    const user = userEvent.setup();
    const onOpen = vi.fn();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([cancelledReservation]),
      ),
    );

    render(
      <MyReservationsPage
        {...BASE_PROPS}
        onBack={vi.fn()}
        onOpenRevisitIntake={onOpen}
        draftStore={pendingStore(true)}
      />,
    );

    await user.click(await screen.findByRole("button", { name: "送信結果を確認する" }));
    expect(onOpen).toHaveBeenCalledWith(cancelledReservation);
    expect(screen.queryByRole("button", { name: /問診を回答/ })).not.toBeInTheDocument();
  });

  it("取消済みで未確定の送信が無ければ入口を出さない", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([cancelledReservation]),
      ),
    );

    render(
      <MyReservationsPage
        {...BASE_PROPS}
        onBack={vi.fn()}
        onOpenRevisitIntake={vi.fn()}
        draftStore={pendingStore(false)}
      />,
    );

    expect(await screen.findByText("キャンセル済")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "送信結果を確認する" })).not.toBeInTheDocument();
  });

  it("EMR-519: 隣接する大きな ID の取消はその ID 文字列のまま DELETE し、他カードに影響しない", async () => {
    const user = userEvent.setup();
    const deletedIds: string[] = [];
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([
          { ...confirmedReservation, id: "9007199254740992" },
          { ...confirmedReservation, id: "9007199254740993" },
        ]),
      ),
      http.delete("/api/liff/:clinicId/my-reservations/:id", ({ params }) => {
        deletedIds.push(String(params.id));
        return new HttpResponse(null, { status: 204 });
      }),
    );

    render(
      <MyReservationsPage
        {...BASE_PROPS}
        onBack={vi.fn()}
        onOpenRevisitIntake={vi.fn()}
        draftStore={pendingStore(false)}
      />,
    );

    const cancelButtons = await screen.findAllByRole("button", { name: "キャンセルする" });
    await user.click(cancelButtons[0]);
    await user.click(screen.getByRole("button", { name: "はい、キャンセルする" }));

    const cancelledBadges = await screen.findAllByText("キャンセル済");
    expect(cancelledBadges).toHaveLength(1);
    expect(deletedIds).toEqual(["9007199254740992"]);
    expect(screen.getByRole("button", { name: "キャンセルする" })).toBeInTheDocument();
  });

  it.each(["1", "9007199254740991", "9007199254740992", "9007199254740993", "9223372036854775807"])(
    "EMR-519: 境界 ID %s の取消は同一文字列の DELETE を送る",
    async (id) => {
      const user = userEvent.setup();
      const deletedIds: string[] = [];
      server.use(
        http.get("/api/liff/:clinicId/my-reservations", () =>
          HttpResponse.json([{ ...confirmedReservation, id }]),
        ),
        http.delete("/api/liff/:clinicId/my-reservations/:id", ({ params }) => {
          deletedIds.push(String(params.id));
          return new HttpResponse(null, { status: 204 });
        }),
      );

      render(
        <MyReservationsPage
          {...BASE_PROPS}
          onBack={vi.fn()}
          onOpenRevisitIntake={vi.fn()}
          draftStore={pendingStore(false)}
        />,
      );

      await user.click(await screen.findByRole("button", { name: "キャンセルする" }));
      await user.click(screen.getByRole("button", { name: "はい、キャンセルする" }));

      expect(await screen.findByText("キャンセル済")).toBeInTheDocument();
      expect(deletedIds).toEqual([id]);
    },
  );

  it("EMR-519: 取消済み予約の送信結果確認は同一 ID 文字列で revisit 導線を開く", async () => {
    const user = userEvent.setup();
    const cancelled = { ...confirmedReservation, id: "9007199254740992", status: "cancelled" };
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () => HttpResponse.json([cancelled])),
    );
    const onOpen = vi.fn();

    render(
      <MyReservationsPage
        {...BASE_PROPS}
        onBack={vi.fn()}
        onOpenRevisitIntake={onOpen}
        draftStore={pendingStore(true)}
      />,
    );

    await user.click(await screen.findByRole("button", { name: "送信結果を確認する" }));
    expect(onOpen).toHaveBeenCalledWith(expect.objectContaining({ id: "9007199254740992" }));
  });
});

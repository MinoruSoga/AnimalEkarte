import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import type { Reservation } from "@/types";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  );
}

/** 空レスポンスを返すハンドラ群。ReservationFormModal 内部のクエリを全て黙らせる */
const silentApiHandlers = [
  http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
  http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
  http.get("/api/v1/masters/reservation-types", () => HttpResponse.json([])),
  http.get("/api/v1/masters/staffs", () => HttpResponse.json([])),
  http.get("/api/v1/shifts/on-duty-staffs", () => HttpResponse.json([])),
  http.get("/api/v1/masters/animal-species", () => HttpResponse.json([])),
];

const noop = () => {};

afterEach(() => {
  server.resetHandlers();
  localStorage.removeItem("auth_current_clinic:v1");
});

// ─────────────────────────────────────────────────────────────
// Issue #52: クリックした日時がフォームの初期値に反映されること
// ─────────────────────────────────────────────────────────────

describe("ReservationFormModal — 初期値セット (Issue #52)", () => {
  // Calendar(react-day-picker) + Radix Popover/Select の初回 render が Docker jsdom で 5s+ かかるため 15s に設定
  it("isOpen=true + initialData.start 設定済みのとき「日付を選択」プレースホルダーが消える", () => {
    server.use(...silentApiHandlers);

    // 週次ビューのスロットクリックで生成される stub と同等の initialData
    const clickedStart = new Date(2026, 4, 21, 14, 30, 0); // 2026-05-21 14:30
    const initialData: Partial<Reservation> = {
      start: clickedStart,
      end: new Date(2026, 4, 21, 15, 30, 0),
      visitType: "first",
      doctor: "",
      isDesignated: false,
      status: "confirmed",
    };

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={initialData}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    // initialData.start が formData に反映されていれば「日付を選択」は表示されない
    expect(screen.queryByText("日付を選択")).not.toBeInTheDocument();
  }, 15000);

  it("isOpen=true + initialData.start=14:30 のとき開始時刻セレクトに 14:30 が表示される", () => {
    server.use(...silentApiHandlers);

    const clickedStart = new Date(2026, 4, 21, 14, 30, 0);
    const initialData: Partial<Reservation> = {
      start: clickedStart,
      end: new Date(2026, 4, 21, 15, 30, 0),
      visitType: "first",
      doctor: "",
      isDesignated: false,
      status: "confirmed",
    };

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={initialData}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    // 開始時刻セレクトのトリガーに "14:30" が描画されているはず
    // (formData.start が未設定なら "10:00" がデフォルト表示される)
    const triggers = screen.getAllByRole("combobox");
    const timeLabels = triggers.map((t) => t.textContent ?? "");
    expect(timeLabels.some((label) => label.includes("14:30"))).toBe(true);
  }, 15000);

  it("isOpen=true + initialData=null のとき今日 10:00 がデフォルトでセットされ「日付を選択」は表示されない", () => {
    server.use(...silentApiHandlers);

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    // 新規作成時は今日 10:00 がデフォルト → 「日付を選択」は出ない
    expect(screen.queryByText("日付を選択")).not.toBeInTheDocument();
    // 開始時刻デフォルトは "10:00"
    const triggers = screen.getAllByRole("combobox");
    const timeLabels = triggers.map((t) => t.textContent ?? "");
    expect(timeLabels.some((label) => label.includes("10:00"))).toBe(true);
  }, 15000);
});

// ─────────────────────────────────────────────────────────────
// Issue #51: 新規飼主インラインフォーム (既存飼主/新規飼主切替)
// ─────────────────────────────────────────────────────────────

describe("ReservationFormModal — 新規飼主モード (Issue #51)", () => {
  it("デフォルト（既存飼主）モードで患者検索UIが表示される", () => {
    server.use(...silentApiHandlers);

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    // 既存飼主モードでは患者検索ラベルが表示される
    expect(screen.getByText("患者検索")).toBeInTheDocument();
    // 新規飼主フォームは表示されない
    expect(screen.queryByTestId("new-owner-name")).not.toBeInTheDocument();
  });

  it("「新規飼主」ボタンをクリックすると4項目の入力フォームが表示される", () => {
    server.use(...silentApiHandlers);

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    const newOwnerBtn = screen.getByTestId("mode-new");
    fireEvent.click(newOwnerBtn);

    // 新規飼主フォームの4項目が表示される
    expect(screen.getByTestId("new-owner-name")).toBeInTheDocument();
    expect(screen.getByTestId("new-owner-phone")).toBeInTheDocument();
    expect(screen.getByTestId("new-owner-pet-name")).toBeInTheDocument();
    expect(screen.getByTestId("new-owner-chief-complaint")).toBeInTheDocument();
    // 患者検索は非表示になる
    expect(screen.queryByText("患者検索")).not.toBeInTheDocument();
  });

  it("電話番号が 0 始まりでない場合 BE 整合フォーマットエラーが表示される", async () => {
    server.use(...silentApiHandlers);
    const user = userEvent.setup({ delay: null });

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    await user.click(screen.getByTestId("mode-new"));
    fireEvent.change(screen.getByTestId("new-owner-name"), { target: { value: "山田太郎" } });
    // 0 始まりでない番号 — BE regex に通らない形式
    fireEvent.change(screen.getByTestId("new-owner-phone"), { target: { value: "1234-5678" } });

    fireEvent.click(screen.getByRole("button", { name: "予約を確定" }));

    await waitFor(() => {
      expect(
        screen.getByText("電話番号の形式が正しくありません（例：090-1234-5678 または 09012345678）")
      ).toBeInTheDocument();
    });
  });

  // Radix Select x2 + MSW + waitFor x3 → Docker jsdom で累計 5 秒超えが稀に発生するため 15s に設定
  it("全フィールド入力後に onSave が newOwnerData を含む引数で呼ばれる", async () => {
    server.use(
      http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
      http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
      http.get("/api/v1/masters/staffs", () => HttpResponse.json([])),
      http.get("/api/v1/shifts/on-duty-staffs", () => HttpResponse.json([])),
      // 動物種: 犬 1件
      http.get("/api/v1/masters/animal-species", () =>
        HttpResponse.json([{ id: 1, name: "犬", is_active: true }])
      ),
      // 予約区分: 一般診療 1件（group なし → "その他" グループ）
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          { id: 5, name: "一般診療", color: "#000000", is_active: true, group_id: null, group: null },
        ])
      )
    );

    const onSave = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={onSave}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    // 新規飼主モードに切り替え（act() ラップ必要なため user.click を維持）
    await user.click(screen.getByTestId("mode-new"));

    // テキストフィールドを入力（fireEvent.change で1イベント完結、タイムアウト防止）
    fireEvent.change(screen.getByTestId("new-owner-name"), { target: { value: "山田太郎" } });
    fireEvent.change(screen.getByTestId("new-owner-phone"), { target: { value: "090-1234-5678" } });
    fireEvent.change(screen.getByTestId("new-owner-pet-name"), { target: { value: "ポチ" } });
    fireEvent.change(screen.getByTestId("new-owner-chief-complaint"), { target: { value: "食欲不振" } });

    // 動物種 Select: Radix は pointerdown でドロップダウンを開くため user.click を維持
    // 選択肢は waitFor で DOM 確認済みのため fireEvent.click で十分
    await user.click(screen.getByTestId("new-owner-species"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "犬" })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("option", { name: "犬" }));

    // 予約区分 Select: 同上
    await user.click(screen.getByTestId("res-type-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "一般診療" })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("option", { name: "一般診療" }));

    // 保存を実行
    fireEvent.click(screen.getByRole("button", { name: "予約を確定" }));

    await waitFor(() => {
      expect(onSave).toHaveBeenCalledOnce();
    });

    const [, , newOwnerArg] = onSave.mock.calls[0];
    expect(newOwnerArg).toMatchObject({
      ownerName: "山田太郎",
      phone: "090-1234-5678",
      petName: "ポチ",
      chiefComplaint: "食欲不振",
      animalSpeciesId: 1,
    });
  }, 15000);
});

describe("ReservationFormModal — 担当者候補", () => {
  it("選択した予約区分を対応不可にしているスタッフを担当者候補から除外する", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
      http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
      http.get("/api/v1/masters/animal-species", () => HttpResponse.json([])),
      http.get("/api/v1/masters/staffs", () =>
        HttpResponse.json([
          {
            id: 10,
            name: "非対応スタッフ",
            is_active: true,
            clinic_assignments: [{ clinic_id: 1, is_main: true }],
          },
          {
            id: 11,
            name: "対応スタッフ",
            is_active: true,
            clinic_assignments: [{ clinic_id: 1, is_main: true }],
          },
        ])
      ),
      http.get("/api/v1/shifts/on-duty-staffs", () =>
        HttpResponse.json([
          { id: 10, name: "非対応スタッフ" },
          { id: 11, name: "対応スタッフ" },
        ])
      ),
      http.get("/api/v1/clinics/1/reservation-staffs", () =>
        HttpResponse.json([
          {
            id: 10,
            name: "非対応スタッフ",
            is_active: true,
            excluded_courses: [{ id: 5, name: "トリミング" }],
          },
          {
            id: 11,
            name: "対応スタッフ",
            is_active: true,
            excluded_courses: [],
          },
        ])
      ),
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          {
            id: 5,
            name: "トリミング",
            color: "#111111",
            is_active: true,
            duration_minutes: 60,
            sort_order: 1,
            is_internal: false,
            category: "trimming",
            group_id: null,
            group: null,
          },
        ])
      )
    );

    const user = userEvent.setup({ delay: null });

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    await user.click(screen.getByTestId("res-type-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "トリミング" })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("option", { name: "トリミング" }));

    await user.click(screen.getByTestId("res-staff-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "対応スタッフ" })).toBeInTheDocument();
    });
    expect(screen.queryByRole("option", { name: "非対応スタッフ" })).not.toBeInTheDocument();
  }, 15000);
});

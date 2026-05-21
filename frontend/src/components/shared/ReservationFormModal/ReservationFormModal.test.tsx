import { describe, it, expect, afterEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
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
});

// ─────────────────────────────────────────────────────────────
// Issue #52: クリックした日時がフォームの初期値に反映されること
// ─────────────────────────────────────────────────────────────

describe("ReservationFormModal — 初期値セット (Issue #52)", () => {
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
  });

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
  });

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
  });
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
});

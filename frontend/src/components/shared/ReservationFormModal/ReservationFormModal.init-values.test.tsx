import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import type { Reservation } from "@/types";
import { createWrapper, silentApiHandlers, noop } from "./ReservationFormModal.test-helpers";

// SearchableSelect は Radix Popover を Radix Dialog の内側で開く。jsdom + カバレッジ計装下では
// Dialog の FocusScope が focus を掴み直し続け、開いた Popover が focus-outside 判定で即座に
// 閉じるため、CI でのみ aria-expanded が false のまま option が現れない。2026-08-23 に CI 上で
// 実測して確認した（click は届いており、body/trigger の pointer-events も disabled も正常。
// fireEvent.click でも開かない）。本テストの対象は Popover の開閉実装ではないので、開閉の
// 意味論だけを保った素の実装へ差し替え、Dialog×Popover の相互作用を構造的に取り除く。
vi.mock("@/components/ui/searchable-select", async (importOriginal) => {
  const actual =
    await importOriginal<typeof import("@/components/ui/searchable-select")>();
  const { useState } = await import("react");
  type Props = Parameters<typeof actual.SearchableSelect>[0];
  function SearchableSelectStub(props: Props) {
    const [open, setOpen] = useState(false);
    const flat = props.groups
      ? props.groups.flatMap((g) => g.options)
      : (props.options ?? []);
    const selected = flat.find((o) => o.value === props.value);
    return (
      <div>
        <button
          type="button"
          role="combobox"
          id={props.id}
          aria-label={props.ariaLabel}
          aria-expanded={open}
          aria-invalid={props.ariaInvalid}
          aria-describedby={props.ariaDescribedBy}
          disabled={props.disabled}
          data-testid={props.triggerTestId}
          className={props.className}
          onClick={() => setOpen((v) => !v)}
        >
          {selected ? selected.label : props.placeholder}
        </button>
        {open
          ? flat.map((o) => (
              <button
                key={o.value}
                type="button"
                role="option"
                aria-selected={o.value === props.value}
                disabled={o.disabled}
                onClick={() => {
                  props.onValueChange(o.value);
                  setOpen(false);
                }}
              >
                {o.label}
              </button>
            ))
          : null}
      </div>
    );
  }
  return { ...actual, SearchableSelect: SearchableSelectStub };
});

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

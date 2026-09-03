import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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
  const actual = await importOriginal<typeof import("@/components/ui/searchable-select")>();
  const { useState } = await import("react");
  type Props = Parameters<typeof actual.SearchableSelect>[0];
  function SearchableSelectStub(props: Props) {
    const [open, setOpen] = useState(false);
    const flat = props.groups ? props.groups.flatMap((g) => g.options) : (props.options ?? []);
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
// FE-RC-003: 過去日付判定は JST の暦日で行う（ブラウザの実タイムゾーンに依存しない）
// ─────────────────────────────────────────────────────────────

describe("ReservationFormModal — 過去日付検証 (FE-RC-003, JST基準)", () => {
  const ORIGINAL_TZ = process.env.TZ;

  afterEach(() => {
    process.env.TZ = ORIGINAL_TZ;
    vi.useRealTimers();
  });

  it("実TZがJSTより1日遅れていても、JST暦日で今日より前の予約は過去日付エラーになる", async () => {
    server.use(...silentApiHandlers);

    // 実UTC now: 2026-05-21T16:00:00Z → JST暦日は 2026-05-22、America/Los_Angeles(PDT)暦日は 2026-05-21
    // Date のみ fake 化し、setTimeout/setInterval は実時間のまま（waitFor/findBy* を阻害しない）
    process.env.TZ = "America/Los_Angeles";
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(Date.UTC(2026, 4, 21, 16, 0, 0)));

    // DatePicker のローカル正午 parse 契約に合わせ、ローカル wall-clock で "2026-05-21" を意図した Date を構築。
    // 旧実装（ブラウザローカル日で truncate 比較）だと LA では「今日と同じ日」= 過去ではないと誤判定する。
    const initialData: Partial<Reservation> = {
      start: new Date(2026, 4, 21, 10, 0, 0),
      end: new Date(2026, 4, 21, 11, 0, 0),
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
      { wrapper: createWrapper() },
    );

    const user = userEvent.setup({ delay: null });
    await user.click(screen.getByRole("button", { name: "予約を確定" }));

    expect(await screen.findByText("本日以降の日付を選択してください")).toBeInTheDocument();
  });

  it("実TZがJSTより1日進んでいても、JST暦日で今日以降の予約は過去日付エラーにならない", async () => {
    server.use(...silentApiHandlers);

    // 実UTC now: 2026-05-20T16:00:00Z → JST暦日は 2026-05-21。
    // Pacific/Kiritimati (UTC+14) のローカル暦日は 2026-05-21 06:00 相当で同じ 05-21 だが、
    // 旧実装は「実行環境のローカル日」に依存するため、TZ次第で判定がぶれること自体が問題だった。
    // JST基準に固定した新実装は、意図した JST 暦日（05-21 = 今日）に対して常に非過去と判定する。
    process.env.TZ = "Pacific/Kiritimati";
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(Date.UTC(2026, 4, 20, 16, 0, 0)));

    const initialData: Partial<Reservation> = {
      start: new Date(2026, 4, 21, 10, 0, 0),
      end: new Date(2026, 4, 21, 11, 0, 0),
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
      { wrapper: createWrapper() },
    );

    const user = userEvent.setup({ delay: null });
    // 既存飼主モードのまま確定 → 患者未選択/区分未選択エラーは出るが、過去日付エラーは出ない
    await user.click(screen.getByRole("button", { name: "予約を確定" }));

    expect(await screen.findByText("患者を選択してください")).toBeInTheDocument();
    expect(screen.queryByText("本日以降の日付を選択してください")).not.toBeInTheDocument();
  });
});

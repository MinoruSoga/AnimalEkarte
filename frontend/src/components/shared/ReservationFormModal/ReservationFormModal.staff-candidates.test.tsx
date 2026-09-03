import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import { createWrapper, noop } from "./ReservationFormModal.test-helpers";

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

describe("ReservationFormModal — 担当者候補", () => {
  it("対応可能コースを持つスタッフだけを担当者候補に残す（肯定形 capability）", async () => {
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
            capable_courses: [],
          },
          {
            id: 11,
            name: "対応スタッフ",
            is_active: true,
            capable_courses: [{ id: 5, name: "トリミング" }],
          },
        ])
      ),
      http.get("/api/v1/masters/reservation-types/5/unavailable-times", () =>
        HttpResponse.json({ data: [] })
      ),
      http.get("/api/v1/reservations/available-times", () =>
        HttpResponse.json([
          { start_time: "0945", end_time: "1045" },
          { start_time: "1230", end_time: "1330" },
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
    // 予約区分サブダイアログでカード選択(id 5 = トリミング)
    fireEvent.click(await screen.findByTestId("res-type-card-5"));

    await user.click(screen.getByTestId("res-staff-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "対応スタッフ" })).toBeInTheDocument();
    });
    expect(screen.queryByRole("option", { name: "非対応スタッフ" })).not.toBeInTheDocument();
  }, 15000);
});

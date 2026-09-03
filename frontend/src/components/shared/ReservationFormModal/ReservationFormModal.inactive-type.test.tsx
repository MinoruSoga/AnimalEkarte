import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import type { Reservation } from "@/types";
import { createWrapper, noop } from "./ReservationFormModal.test-helpers";

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

describe("ReservationFormModal — BUG-015 inactive reservation type edit", () => {
  const inactiveEditHandlers = [
    http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
    http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
    http.get("/api/v1/masters/animal-species", () => HttpResponse.json([])),
    http.get("/api/v1/masters/staffs", () => HttpResponse.json([])),
    http.get("/api/v1/shifts/on-duty-staffs", () => HttpResponse.json([])),
    http.get("/api/v1/clinics/1/reservation-staffs", () => HttpResponse.json([])),
    http.get("/api/v1/masters/reservation-types/2/unavailable-times", () =>
      HttpResponse.json({ data: [] }),
    ),
    http.get("/api/v1/masters/reservation-types", () =>
      HttpResponse.json([
        {
          id: 1,
          name: "一般診療",
          color: "#111111",
          is_active: true,
          duration_minutes: 30,
          sort_order: 1,
          is_internal: false,
          category: "general",
          group_id: null,
          group: null,
        },
        {
          id: 2,
          name: "旧コース",
          color: "#222222",
          is_active: false,
          duration_minutes: 60,
          sort_order: 2,
          is_internal: false,
          category: "general",
          group_id: null,
          group: null,
        },
        {
          id: 3,
          name: "別の無効コース",
          color: "#333333",
          is_active: false,
          duration_minutes: 45,
          sort_order: 3,
          is_internal: false,
          category: "general",
          group_id: null,
          group: null,
        },
      ]),
    ),
  ];

  it("edit mode keeps inactive type label with （無効） and retains start time when slots are empty", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      ...inactiveEditHandlers,
      http.get("/api/v1/reservations/available-times", () => HttpResponse.json([])),
    );

    const initialData: Partial<Reservation> = {
      id: "100",
      start: new Date(2026, 5, 1, 14, 30, 0),
      end: new Date(2026, 5, 1, 15, 30, 0),
      visitType: "revisit",
      type: "2",
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
        canCreate={false}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    await waitFor(() => {
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("旧コース");
    });
    expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("（無効）");
    expect(screen.getByTestId("res-start-time-trigger")).toHaveTextContent("14:30");
  }, 15000);

  it("picker does not offer other inactive types while keeping the current inactive selection visible", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      ...inactiveEditHandlers,
      http.get("/api/v1/reservations/available-times", () => HttpResponse.json([])),
    );
    const user = userEvent.setup({ delay: null });

    const initialData: Partial<Reservation> = {
      id: "100",
      start: new Date(2026, 5, 1, 14, 30, 0),
      end: new Date(2026, 5, 1, 15, 30, 0),
      visitType: "revisit",
      type: "2",
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
        canCreate={false}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    await waitFor(() => {
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("旧コース");
    });

    await user.click(screen.getByTestId("res-type-trigger"));
    expect(await screen.findByTestId("res-type-card-1")).toBeInTheDocument();
    expect(screen.queryByTestId("res-type-card-3")).not.toBeInTheDocument();
  }, 15000);

  it("available-times failure does not clear type or start time", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      ...inactiveEditHandlers,
      http.get("/api/v1/reservations/available-times", () =>
        HttpResponse.json({ error: "reservation type is inactive" }, { status: 400 }),
      ),
    );

    const initialData: Partial<Reservation> = {
      id: "100",
      start: new Date(2026, 5, 1, 14, 30, 0),
      end: new Date(2026, 5, 1, 15, 30, 0),
      visitType: "revisit",
      type: "2",
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
        canCreate={false}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    await waitFor(() => {
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("旧コース");
    });
    expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("（無効）");
    expect(screen.getByTestId("res-start-time-trigger")).toHaveTextContent("14:30");
  }, 15000);
});

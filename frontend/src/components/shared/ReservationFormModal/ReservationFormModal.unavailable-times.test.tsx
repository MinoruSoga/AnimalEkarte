import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
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

describe("ReservationFormModal — 予約不可時間", () => {
  it("選択した予約区分の予約不可時間を開始時刻候補から除外する", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
      http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
      http.get("/api/v1/masters/animal-species", () => HttpResponse.json([])),
      http.get("/api/v1/masters/staffs", () => HttpResponse.json([])),
      http.get("/api/v1/shifts/on-duty-staffs", () => HttpResponse.json([])),
      http.get("/api/v1/clinics/1/reservation-staffs", () => HttpResponse.json([])),
      http.get("/api/v1/masters/reservation-types/5/unavailable-times", () =>
        HttpResponse.json({
          data: [
            {
              id: 1,
              clinic_id: 1,
              reservation_type_id: 5,
              unavailable_type: "specific",
              specific_date: "2026-06-01",
              start_time: "10:00",
              end_time: "11:00",
              created_at: "2026-05-29T00:00:00Z",
              updated_at: "2026-05-29T00:00:00Z",
            },
          ],
        })
      ),
      http.get("/api/v1/reservations/available-times", () =>
        HttpResponse.json([
          { start_time: "0900", end_time: "0930" },
          { start_time: "0945", end_time: "1045" },
          { start_time: "1100", end_time: "1200" },
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
    const initialData: Partial<Reservation> = {
      start: new Date(2026, 5, 1, 9, 0, 0),
      end: new Date(2026, 5, 1, 9, 30, 0),
      visitType: "revisit",
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

    await user.click(screen.getByTestId("res-type-trigger"));
    // 予約区分サブダイアログでカード選択(id 5 = トリミング)
    fireEvent.click(await screen.findByTestId("res-type-card-5"));

    await waitFor(() => {
      expect(screen.getByTestId("res-start-time-trigger")).toHaveTextContent("09:00");
    });
    await user.click(screen.getByTestId("res-start-time-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "09:45" })).toBeInTheDocument();
    });
    expect(screen.queryByRole("option", { name: "10:00" })).not.toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "10:45" })).not.toBeInTheDocument();
    expect(screen.getByRole("option", { name: "11:00" })).toBeInTheDocument();
  }, 15000);

  it("空き枠の開始時刻を選ぶと終了時刻も同じ枠に合わせる", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
      http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
      http.get("/api/v1/masters/animal-species", () => HttpResponse.json([])),
      http.get("/api/v1/masters/staffs", () => HttpResponse.json([])),
      http.get("/api/v1/shifts/on-duty-staffs", () => HttpResponse.json([])),
      http.get("/api/v1/clinics/1/reservation-staffs", () => HttpResponse.json([])),
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
    const initialData: Partial<Reservation> = {
      start: new Date(2026, 5, 1, 9, 0, 0),
      end: new Date(2026, 5, 1, 9, 30, 0),
      visitType: "revisit",
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

    await user.click(screen.getByTestId("res-type-trigger"));
    // 予約区分サブダイアログでカード選択(id 5 = トリミング)
    fireEvent.click(await screen.findByTestId("res-type-card-5"));

    await user.click(screen.getByTestId("res-start-time-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "12:30" })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("option", { name: "12:30" }));

    expect(screen.getByTestId("res-start-time-trigger")).toHaveTextContent("12:30");
    expect(screen.getByTestId("res-end-time-trigger")).toHaveTextContent("13:30");
  }, 15000);
});

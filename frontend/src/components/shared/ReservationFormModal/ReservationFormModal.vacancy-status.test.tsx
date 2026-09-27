import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import type { Reservation } from "@/types";
import { createWrapper, noop } from "./ReservationFormModal.test-helpers";

// unavailable-times テストと同じ理由で SearchableSelect を stub に差し替える
// （jsdom 下で Dialog 内 Popover が即座に閉じるため）。
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

describe("ReservationFormModal — 空き状況〇△✕ (EMR-170)", () => {
  const baseHandlers = [
    http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
    http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
    http.get("/api/v1/masters/animal-species", () => HttpResponse.json([])),
    http.get("/api/v1/masters/staffs", () => HttpResponse.json([])),
    http.get("/api/v1/shifts/on-duty-staffs", () => HttpResponse.json([])),
    http.get("/api/v1/clinics/1/reservation-staffs", () => HttpResponse.json([])),
    http.get("/api/v1/masters/reservation-types/5/unavailable-times", () =>
      HttpResponse.json({ data: [] }),
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
      ]),
    ),
  ];

  const initialData: Partial<Reservation> = {
    start: new Date(2026, 5, 1, 9, 0, 0),
    end: new Date(2026, 5, 1, 9, 30, 0),
    visitType: "revisit",
    doctor: "",
    isDesignated: false,
    status: "confirmed",
  };

  it("枠ごとの空き状況を 〇/△/✕ のテキスト付きで表示し、満員枠は選択不可にする", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      ...baseHandlers,
      http.get("/api/v1/reservations/available-times", () =>
        HttpResponse.json([
          { start_time: "0900", end_time: "0930", status: "available", remaining: 2 },
          { start_time: "1000", end_time: "1030", status: "low", remaining: 1 },
          { start_time: "1100", end_time: "1200", status: "full", remaining: 0 },
        ]),
      ),
    );

    const user = userEvent.setup({ delay: null });
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

    await user.click(screen.getByTestId("res-type-trigger"));
    fireEvent.click(await screen.findByTestId("res-type-card-5"));

    await user.click(screen.getByTestId("res-start-time-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: /09:00/ })).toBeInTheDocument();
    });

    // 記号＋テキストラベル（色だけに依存しない）
    // 注: Radix は選択中アイテムの内容を trigger にも複製するため option 内に限定して検査する
    const availableOption = screen.getByRole("option", { name: /09:00/ });
    const lowOption = screen.getByRole("option", { name: /10:00/ });
    const fullOption = screen.getByRole("option", { name: /11:00/ });
    expect(within(availableOption).getByTestId("res-start-time-vacancy-09:00")).toHaveTextContent(
      "〇 空きあり",
    );
    expect(within(lowOption).getByTestId("res-start-time-vacancy-10:00")).toHaveTextContent(
      "△ 残り1枠",
    );
    expect(within(fullOption).getByTestId("res-start-time-vacancy-11:00")).toHaveTextContent(
      "✕ 満員",
    );

    // 満員枠は aria-disabled で選択不可
    expect(fullOption).toHaveAttribute("aria-disabled", "true");
    expect(availableOption).not.toHaveAttribute("aria-disabled");
    expect(lowOption).not.toHaveAttribute("aria-disabled");
  }, 15000);

  it("満員枠をクリックしても開始時刻は変わらない", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      ...baseHandlers,
      http.get("/api/v1/reservations/available-times", () =>
        HttpResponse.json([
          { start_time: "0900", end_time: "0930", status: "available", remaining: 2 },
          { start_time: "1100", end_time: "1200", status: "full", remaining: 0 },
        ]),
      ),
    );

    const user = userEvent.setup({ delay: null });
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

    await user.click(screen.getByTestId("res-type-trigger"));
    fireEvent.click(await screen.findByTestId("res-type-card-5"));

    await user.click(screen.getByTestId("res-start-time-trigger"));
    const fullOption = await screen.findByRole("option", { name: /11:00/ });
    fireEvent.click(fullOption);

    // 選択は変わらず 09:00 のまま
    expect(screen.getByTestId("res-start-time-trigger")).toHaveTextContent("09:00");
    expect(screen.getByTestId("res-end-time-trigger")).toHaveTextContent("09:30");
  }, 15000);

  it("status なしの旧応答 shape では従来どおり時刻のみ表示する", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      ...baseHandlers,
      http.get("/api/v1/reservations/available-times", () =>
        HttpResponse.json([
          { start_time: "0945", end_time: "1045" },
          { start_time: "1230", end_time: "1330" },
        ]),
      ),
    );

    const user = userEvent.setup({ delay: null });
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

    await user.click(screen.getByTestId("res-type-trigger"));
    fireEvent.click(await screen.findByTestId("res-type-card-5"));

    await user.click(screen.getByTestId("res-start-time-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "09:45" })).toBeInTheDocument();
    });
    expect(screen.queryByTestId("res-start-time-vacancy-09:45")).not.toBeInTheDocument();
    expect(screen.queryByTestId("res-start-time-vacancy-12:30")).not.toBeInTheDocument();
  }, 15000);
});

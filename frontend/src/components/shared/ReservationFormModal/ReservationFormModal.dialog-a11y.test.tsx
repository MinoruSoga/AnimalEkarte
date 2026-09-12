import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import { createWrapper, noop, silentApiHandlers } from "./ReservationFormModal.test-helpers";

const MISSING_DESCRIPTION_RE =
  /Missing `Description` or `aria-describedby=\{undefined\}` for \{DialogContent\}/;

// SearchableSelect Popover×Dialog focus interaction is out of scope for this a11y contract.
// Keep production Dialog/DialogContent/DialogDescription — do not mock them away.
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

function missingDescriptionWarnings(spy: { mock: { calls: unknown[][] } }): string[] {
  return spy.mock.calls
    .map((args) => String(args[0] ?? ""))
    .filter((msg) => MISSING_DESCRIPTION_RE.test(msg));
}

afterEach(() => {
  server.resetHandlers();
  localStorage.removeItem("auth_current_clinic:v1");
});

describe("ReservationFormModal — Dialog Description a11y (BUG-RES-DIALOG-A11Y-CONSOLE)", () => {
  it("opens with accessible description and emits 0 Missing Description warnings", async () => {
    server.use(...silentApiHandlers);
    const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

    try {
      render(
        <ReservationFormModal
          isOpen={true}
          onClose={noop}
          onSave={noop}
          canCreate={true}
          canEdit={false}
        />,
        { wrapper: createWrapper() },
      );

      const dialog = await screen.findByRole("dialog");
      expect(dialog).toHaveAccessibleName("新規予約作成");
      expect(dialog).toHaveAccessibleDescription(
        "左側のリストからペットを選択し、右側のフォームで予約情報を入力してください",
      );

      await waitFor(() => {
        expect(missingDescriptionWarnings(warnSpy)).toEqual([]);
      });
    } finally {
      warnSpy.mockRestore();
    }
  }, 15000);

  it("nested ReservationTypePickerDialog open/close keeps 0 Missing Description warnings and restores focus", async () => {
    server.use(
      ...silentApiHandlers,
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          {
            id: 1,
            name: "一般診療",
            category: "診察",
            color: "#111111",
            duration_minutes: 20,
            is_active: true,
          },
        ]),
      ),
    );
    const user = userEvent.setup();
    const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

    try {
      render(
        <ReservationFormModal
          isOpen={true}
          onClose={noop}
          onSave={noop}
          canCreate={true}
          canEdit={false}
        />,
        { wrapper: createWrapper() },
      );

      await screen.findByRole("dialog", { name: "新規予約作成" });

      const typeTrigger = await screen.findByTestId("res-type-trigger");
      await user.click(typeTrigger);

      const picker = await screen.findByRole("dialog", { name: "予約区分を選択" });
      expect(picker).toHaveAccessibleDescription(
        "予約区分をカテゴリや検索で絞り込んで選択します。",
      );

      await user.keyboard("{Escape}");

      await waitFor(() => {
        expect(screen.queryByRole("dialog", { name: "予約区分を選択" })).not.toBeInTheDocument();
      });

      expect(screen.getByRole("dialog", { name: "新規予約作成" })).toBeInTheDocument();
      expect(missingDescriptionWarnings(warnSpy)).toEqual([]);
    } finally {
      warnSpy.mockRestore();
    }
  }, 20000);
});

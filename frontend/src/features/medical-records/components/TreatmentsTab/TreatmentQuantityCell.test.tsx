import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createRef } from "react";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { TreatmentQuantityCell } from "./TreatmentQuantityCell";
import type { Treatment, UpdateTreatmentInput } from "../../types";
import type { MedicineDoseContext } from "../../api/medicine-dose-lookup";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

const treatment: Treatment = {
  id: "1",
  medical_record_id: "10",
  item_type: "consultation",
  unit_price: 1000,
  quantity: 1,
  is_selected: true,
  status: "pending",
  content: "診察",
  memo: "",
  is_insurance: false,
  discount_rate: 0,
  discount_amount: 0,
  sort_order: 0,
  created_at: "2026-07-12T00:00:00Z",
  updated_at: "2026-07-12T00:00:00Z",
};

const doseContext: MedicineDoseContext = {
  medicines: undefined,
  petSpecies: null,
  weightKg: null,
};

describe("TreatmentQuantityCell — Enter×2 / Blur / Escape", () => {
  beforeEach(() => {
    server.use(http.get("*/v1/masters/medicines/:id/dose-params", () => HttpResponse.json([])));
  });

  function renderEditing(
    onUpdate: (id: string, input: UpdateTreatmentInput) => void,
    onStopEdit = vi.fn(),
  ) {
    const inputRef = createRef<HTMLInputElement | null>();
    render(
      <table>
        <tbody>
          <tr>
            <TreatmentQuantityCell
              treatment={treatment}
              doseContext={doseContext}
              isEditing={true}
              inputRef={inputRef}
              onStartEdit={vi.fn()}
              onStopEdit={onStopEdit}
              onUpdate={onUpdate}
            />
          </tr>
        </tbody>
      </table>,
      { wrapper: createWrapper() },
    );
    return { onStopEdit };
  }

  it("first Enter does not call onUpdate; second Enter persists quantity once", async () => {
    const onUpdate = vi.fn();
    const user = userEvent.setup();
    renderEditing(onUpdate);

    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    expect(quantityInput).toHaveAccessibleDescription(/Enterを2回押して確定/);
    await user.clear(quantityInput);
    await user.type(quantityInput, "3");
    await user.keyboard("{Enter}");
    expect(onUpdate).not.toHaveBeenCalled();

    await user.keyboard("{Enter}");
    expect(onUpdate).toHaveBeenCalledTimes(1);
    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 3 });
  });

  it("Blur still commits on a single blur without requiring Enter", async () => {
    const onUpdate = vi.fn();
    const user = userEvent.setup();
    renderEditing(onUpdate);

    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "4");
    await user.tab();

    expect(onUpdate).toHaveBeenCalledTimes(1);
    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 4 });
  });

  it("Escape cancels without onUpdate even after first Enter arming", async () => {
    const onUpdate = vi.fn();
    const onStopEdit = vi.fn();
    const user = userEvent.setup();
    renderEditing(onUpdate, onStopEdit);

    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "7");
    await user.keyboard("{Enter}");
    expect(onUpdate).not.toHaveBeenCalled();

    await user.keyboard("{Escape}");
    expect(onUpdate).not.toHaveBeenCalled();
    expect(onStopEdit).toHaveBeenCalled();
  });

  it("does not treat a held Enter repeat as the second confirmation", () => {
    const onUpdate = vi.fn();
    renderEditing(onUpdate);

    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    fireEvent.change(quantityInput, { target: { value: "3" } });
    fireEvent.keyDown(quantityInput, { key: "Enter", repeat: false });
    fireEvent.keyDown(quantityInput, { key: "Enter", repeat: true });
    expect(onUpdate).not.toHaveBeenCalled();

    fireEvent.keyDown(quantityInput, { key: "Enter", repeat: false });
    expect(onUpdate).toHaveBeenCalledTimes(1);
    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 3 });
  });

  it("does not arm a commit from an IME composition Enter", () => {
    const onUpdate = vi.fn();
    renderEditing(onUpdate);

    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    fireEvent.change(quantityInput, { target: { value: "3" } });
    fireEvent.keyDown(quantityInput, { key: "Enter", isComposing: true, keyCode: 229 });
    fireEvent.keyDown(quantityInput, { key: "Enter", isComposing: false });
    expect(onUpdate).not.toHaveBeenCalled();

    fireEvent.keyDown(quantityInput, { key: "Enter", isComposing: false });
    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 3 });
  });

  it("does not commit when an IME composition Enter arrives after arming", () => {
    const onUpdate = vi.fn();
    renderEditing(onUpdate);

    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    fireEvent.change(quantityInput, { target: { value: "3" } });
    fireEvent.keyDown(quantityInput, { key: "Enter", isComposing: false });
    fireEvent.keyDown(quantityInput, { key: "Enter", isComposing: true, keyCode: 229 });
    expect(onUpdate).not.toHaveBeenCalled();

    fireEvent.keyDown(quantityInput, { key: "Enter", isComposing: false });
    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 3 });
  });
});

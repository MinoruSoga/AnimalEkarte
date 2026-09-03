import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { UpdateVitalInput, Vital } from "../../types";
import { VitalsAddRow, VitalsEditRow } from "./VitalsTabRows";
import type { VitalsAddFormState } from "./vitals-tab-table-model";

const baseVital: Vital = {
  id: "55",
  medical_record_id: "77",
  recorded_at: "2026-08-01T10:00:00+09:00",
  temperature: 38.5,
  heart_rate: 100,
  respiration_rate: 20,
  weight: 8.5,
  weight_unit: "Kg",
  note: null,
  created_at: "2026-08-01T10:00:00+09:00",
  updated_at: "2026-08-01T10:00:00+09:00",
};

describe("VitalsEditRow weight unit toggle (BUG-015)", () => {
  it("converts value and unit atomically; save payload preserves physical mass", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();

    render(
      <table>
        <tbody>
          <VitalsEditRow
            vital={baseVital}
            onSave={onSave}
            onCancel={() => undefined}
            isPending={false}
          />
        </tbody>
      </table>
    );

    const unitToggle = screen.getByRole("button", { name: "Kg" });
    await user.click(unitToggle);

    expect(screen.getByRole("button", { name: "g" })).toBeInTheDocument();
    const weightInput = screen.getByLabelText(/体重/);
    expect(weightInput).toHaveValue(8500);

    await user.click(screen.getByTitle("保存"));

    expect(onSave).toHaveBeenCalledTimes(1);
    const [, payload] = onSave.mock.calls[0] as [string, UpdateVitalInput];
    expect(payload.weight).toBe(8500);
    expect(payload.weight_unit).toBe("g");
  });

  it("round-trips Kg → g → Kg without changing physical mass in the save payload", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();

    render(
      <table>
        <tbody>
          <VitalsEditRow
            vital={{ ...baseVital, weight: 5, weight_unit: "Kg" }}
            onSave={onSave}
            onCancel={() => undefined}
            isPending={false}
          />
        </tbody>
      </table>
    );

    await user.click(screen.getByRole("button", { name: "Kg" }));
    expect(screen.getByLabelText(/体重/)).toHaveValue(5000);
    await user.click(screen.getByRole("button", { name: "g" }));
    expect(screen.getByLabelText(/体重/)).toHaveValue(5);

    await user.click(screen.getByTitle("保存"));
    const [, payload] = onSave.mock.calls[0] as [string, UpdateVitalInput];
    expect(payload.weight).toBe(5);
    expect(payload.weight_unit).toBe("Kg");
  });
});

describe("VitalsAddRow weight unit toggle (BUG-015)", () => {
  it("invokes onChange with converted weight and unit for a filled weight", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const addForm: VitalsAddFormState = {
      recorded_at: "2026-08-01T10:00",
      temperature: "",
      heart_rate: "",
      respiration_rate: "",
      weight: "5",
      weight_unit: "Kg",
      note: "",
    };

    render(
      <VitalsAddRow
        addForm={addForm}
        errors={{}}
        isPending={false}
        onChange={onChange}
        formAction={() => undefined}
        onCancel={() => undefined}
      />
    );

    await user.click(screen.getByRole("button", { name: "Kg" }));

    // Atomic multi-field update via single onChange with weight + unit.
    expect(onChange).toHaveBeenCalledWith({
      weight: "5000",
      weight_unit: "g",
    });
  });

  it("toggles unit only when weight is empty", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const addForm: VitalsAddFormState = {
      recorded_at: "",
      temperature: "",
      heart_rate: "",
      respiration_rate: "",
      weight: "",
      weight_unit: "Kg",
      note: "",
    };

    render(
      <VitalsAddRow
        addForm={addForm}
        errors={{}}
        isPending={false}
        onChange={onChange}
        formAction={() => undefined}
        onCancel={() => undefined}
      />
    );

    await user.click(screen.getByRole("button", { name: "Kg" }));
    expect(onChange).toHaveBeenCalledWith({
      weight: "",
      weight_unit: "g",
    });
  });
});

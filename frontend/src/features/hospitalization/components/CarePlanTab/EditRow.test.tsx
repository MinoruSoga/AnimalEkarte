import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EditRow } from "./EditRow";
import type { CarePlanItem } from "../../api/care-plan-items";

// AddForm.test.tsx と同じ理由でテストダブルに置き換える(CarePlanRefSelect.test.tsx で別途検証済み)。
vi.mock("./CarePlanRefSelect", () => ({
    CarePlanRefSelect: ({
        type,
        value,
        onChange,
    }: {
        type: string;
        value: string | null;
        onChange: (v: string | null) => void;
    }) => (
        <input
            aria-label="ref-select-stub"
            data-type={type}
            value={value ?? ""}
            onChange={(e) => onChange(e.target.value || null)}
        />
    ),
}));

afterEach(() => {
    vi.clearAllMocks();
});

const baseItem: CarePlanItem = {
    id: "item-1",
    hospitalization_id: "hosp-1",
    type: "instruction",
    name: "既存指示",
    description: "",
    timing: ["morning"],
    status: "active",
    notes: "",
    medicine_id: null,
    procedure_id: null,
    hospitalization_plan_id: null,
    unit_price: 0,
    category: "",
    sort_order: 0,
    created_at: "2026-07-17T00:00:00Z",
    updated_at: "2026-07-17T00:00:00Z",
};

async function selectType(user: ReturnType<typeof userEvent.setup>, label: string) {
    await user.click(screen.getByRole("combobox"));
    const listbox = await screen.findByRole("listbox");
    await user.click(within(listbox).getByText(label));
}

describe("EditRow — type連動マスタ参照(BUG-403)", () => {
    it("既存 medicine 項目を編集し type=処置・検査 へ変更すると、保存には処置選択が必須になる", async () => {
        const user = userEvent.setup();
        const onSave = vi.fn();
        const item: CarePlanItem = { ...baseItem, type: "medicine", medicine_id: "med-9", name: "既存投薬" };
        render(<EditRow item={item} onSave={onSave} onCancel={vi.fn()} />);

        await selectType(user, "処置・検査");

        expect(screen.getByRole("button", { name: /保存/ })).toBeDisabled();
    });

    it("type=持ち物 で入院プランを選択して保存すると hospitalization_plan_id が payload に含まれる", async () => {
        const user = userEvent.setup();
        const onSave = vi.fn();
        render(<EditRow item={baseItem} onSave={onSave} onCancel={vi.fn()} />);

        await selectType(user, "持ち物");
        await user.type(screen.getByLabelText("ref-select-stub"), "plan-1");
        await user.click(screen.getByRole("button", { name: /保存/ }));

        expect(onSave).toHaveBeenCalledWith(
            expect.objectContaining({
                type: "item",
                hospitalization_plan_id: "plan-1",
                medicine_id: null,
                procedure_id: null,
            })
        );
    });

    it("type=指示・その他(参照不要)のままなら参照選択欄は表示されない", () => {
        render(<EditRow item={baseItem} onSave={vi.fn()} onCancel={vi.fn()} />);
        expect(screen.queryByLabelText("ref-select-stub")).not.toBeInTheDocument();
    });
});

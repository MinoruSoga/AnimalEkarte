import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/testing/TestUtils";
import { CarePlanRefSelect } from "./CarePlanRefSelect";

vi.mock("@/hooks/use-treatment-master", () => ({
    useGetAllMedicinesMaster: () => ({
        data: [{ id: "1", name: "アモキシシリン" }],
        isLoading: false,
    }),
    useGetAllProcedures: () => ({
        data: [{ id: "2", name: "血液検査" }],
        isLoading: false,
    }),
    useGetAllHospitalizationPlansMaster: () => ({
        data: [{ id: "3", name: "スタンダード入院プラン" }],
        isLoading: false,
    }),
}));

afterEach(() => {
    vi.clearAllMocks();
});

describe("CarePlanRefSelect", () => {
    it("type=medicine のとき薬剤選択の combobox を表示する", () => {
        render(
            <CarePlanRefSelect type="medicine" value={null} onChange={vi.fn()} />,
            { wrapper: createTestWrapper() }
        );
        expect(screen.getByRole("combobox")).toHaveTextContent("薬剤を選択");
    });

    it("type=treatment のとき処置・検査選択の combobox を表示する", () => {
        render(
            <CarePlanRefSelect type="treatment" value={null} onChange={vi.fn()} />,
            { wrapper: createTestWrapper() }
        );
        expect(screen.getByRole("combobox")).toHaveTextContent("処置・検査を選択");
    });

    it("type=item のとき入院プラン選択の combobox を表示する", () => {
        render(
            <CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />,
            { wrapper: createTestWrapper() }
        );
        expect(screen.getByRole("combobox")).toHaveTextContent("入院プランを選択");
    });

    it("type=food のとき何も描画しない(食事はマスタ参照不要)", () => {
        const { container } = render(
            <CarePlanRefSelect type="food" value={null} onChange={vi.fn()} />,
            { wrapper: createTestWrapper() }
        );
        expect(container).toBeEmptyDOMElement();
    });

    it("薬剤を選択すると onChange が選択した medicine の id で呼ばれる", async () => {
        const user = userEvent.setup();
        const handleChange = vi.fn();
        render(
            <CarePlanRefSelect type="medicine" value={null} onChange={handleChange} />,
            { wrapper: createTestWrapper() }
        );
        await user.click(screen.getByRole("combobox"));
        await user.click(await screen.findByText("アモキシシリン"));
        expect(handleChange).toHaveBeenCalledWith("1");
    });
});

import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/testing/TestUtils";
import { CarePlanRefSelect } from "./CarePlanRefSelect";

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllMedicinesMaster: () => ({
    data: [{ id: "1", name: "アモキシシリン", price: 100 }],
    isLoading: false,
  }),
  useGetAllProcedures: () => ({
    data: [{ id: "2", name: "血液検査", price: 4000 }],
    isLoading: false,
  }),
  // onUnitPriceChange 経路の回帰 pin: onChange は id のみを運び、price は別経路。
  useGetAllHospitalizationPlansMaster: () => ({
    data: [{ id: "3", name: "スタンダード入院プラン", price: 1200 }],
    isLoading: false,
  }),
}));

afterEach(() => {
  vi.clearAllMocks();
});

describe("CarePlanRefSelect", () => {
  it("type=medicine のとき薬剤選択の combobox を表示する", () => {
    render(<CarePlanRefSelect type="medicine" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper(),
    });
    expect(screen.getByRole("combobox")).toHaveTextContent("薬剤を選択");
  });

  it("type=treatment のとき処置・検査選択の combobox を表示する", () => {
    render(<CarePlanRefSelect type="treatment" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper(),
    });
    expect(screen.getByRole("combobox")).toHaveTextContent("処置・検査を選択");
  });

  it("type=item のとき入院プラン選択の combobox を表示する", () => {
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper(),
    });
    expect(screen.getByRole("combobox")).toHaveTextContent("入院プランを選択");
  });

  it("type=food のとき何も描画しない(食事はマスタ参照不要)", () => {
    const { container } = render(
      <CarePlanRefSelect type="food" value={null} onChange={vi.fn()} />,
      { wrapper: createTestWrapper() },
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("薬剤を選択すると onChange が選択した medicine の id で呼ばれる", async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    render(<CarePlanRefSelect type="medicine" value={null} onChange={handleChange} />, {
      wrapper: createTestWrapper(),
    });
    await user.click(screen.getByRole("combobox"));
    await user.click(await screen.findByText("アモキシシリン"));
    expect(handleChange).toHaveBeenCalledWith("1");
  });

  it("薬剤を選択すると onUnitPriceChange が薬剤マスタ price 100 で呼ばれる", async () => {
    const user = userEvent.setup();
    const handleUnitPrice = vi.fn();
    render(
      <CarePlanRefSelect
        type="medicine"
        value={null}
        onChange={vi.fn()}
        onUnitPriceChange={handleUnitPrice}
      />,
      { wrapper: createTestWrapper() },
    );
    await user.click(screen.getByRole("combobox"));
    await user.click(await screen.findByText("アモキシシリン"));
    expect(handleUnitPrice).toHaveBeenCalledWith(100);
  });

  it("処置・検査を選択すると onUnitPriceChange が処置マスタ price 4000 で呼ばれる", async () => {
    const user = userEvent.setup();
    const handleUnitPrice = vi.fn();
    render(
      <CarePlanRefSelect
        type="treatment"
        value={null}
        onChange={vi.fn()}
        onUnitPriceChange={handleUnitPrice}
      />,
      { wrapper: createTestWrapper() },
    );
    await user.click(screen.getByRole("combobox"));
    await user.click(await screen.findByText("血液検査"));
    expect(handleUnitPrice).toHaveBeenCalledWith(4000);
  });

  it("入院プランを選択すると onUnitPriceChange がプランマスタ price 1200 で呼ばれる", async () => {
    const user = userEvent.setup();
    const handleUnitPrice = vi.fn();
    render(
      <CarePlanRefSelect
        type="item"
        value={null}
        onChange={vi.fn()}
        onUnitPriceChange={handleUnitPrice}
      />,
      { wrapper: createTestWrapper() },
    );
    await user.click(screen.getByRole("combobox"));
    await user.click(await screen.findByText("スタンダード入院プラン"));
    expect(handleUnitPrice).toHaveBeenCalledWith(1200);
  });

  it("入院プラン選択の onChange は id のみで、マスタ price 1200 を第2引数にもオブジェクトにも渡さない", async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    render(<CarePlanRefSelect type="item" value={null} onChange={handleChange} />, {
      wrapper: createTestWrapper(),
    });
    await user.click(screen.getByRole("combobox"));
    await user.click(await screen.findByText("スタンダード入院プラン"));

    expect(handleChange).toHaveBeenCalledTimes(1);
    expect(handleChange).toHaveBeenCalledWith("3");
    expect(handleChange.mock.calls[0]).toHaveLength(1);
    expect(handleChange.mock.calls[0][0]).not.toEqual(expect.objectContaining({ price: 1200 }));
  });
});

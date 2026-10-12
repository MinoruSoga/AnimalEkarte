import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/testing/TestUtils";
import { CarePlanRefSelect } from "./CarePlanRefSelect";

// EMR-355: マスタ取得の loading/error/empty/success を per-test で差し替えられるよう
// hook の戻り値を mutable に保持する。実物の React Query 結果と同じフィールドを持つ
// 最小のクエリ形状を queryState() が作る。
const mocks = vi.hoisted(() => ({
  medicines: {} as Record<string, unknown>,
  procedures: {} as Record<string, unknown>,
  plans: {} as Record<string, unknown>,
  permission: { canView: true, canCreate: true, canEdit: true, canDelete: false },
  permissionResources: [] as string[],
}));

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllMedicinesMaster: () => mocks.medicines,
  useGetAllProcedures: () => mocks.procedures,
  // onUnitPriceChange 経路の回帰 pin: onChange は id のみを運び、price は別経路。
  useGetAllHospitalizationPlansMaster: () => mocks.plans,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: (resource: string) => {
    mocks.permissionResources.push(resource);
    return mocks.permission;
  },
}));

function queryState(partial: Record<string, unknown> = {}) {
  return {
    data: undefined,
    isPending: false,
    isLoading: false,
    isSuccess: true,
    isError: false,
    isFetching: false,
    error: null,
    refetch: vi.fn(),
    ...partial,
  };
}

beforeEach(() => {
  mocks.medicines = queryState({
    data: [{ id: "1", name: "アモキシシリン", price: 100 }],
  });
  mocks.procedures = queryState({
    data: [{ id: "2", name: "血液検査", price: 4000 }],
  });
  mocks.plans = queryState({
    data: [{ id: "3", name: "スタンダード入院プラン", price: 1200 }],
  });
  mocks.permission = { canView: true, canCreate: true, canEdit: true, canDelete: false };
  mocks.permissionResources.length = 0;
});

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

// EMR-355: 候補取得の loading/error/成功0件/成功有りを区別する。
// 「選択肢がないのに必須」という行き止まりを、取得失敗の再試行導線と
// 未登録の説明+設定導線で解消する回帰テスト。
describe("CarePlanRefSelect — マスタ取得状態 (EMR-355)", () => {
  it("選択 type のクエリが読み込み中のとき combobox は disabled で「読み込み中」を表示する", () => {
    mocks.plans = queryState({
      isPending: true,
      isLoading: true,
      isFetching: true,
      isSuccess: false,
    });
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper(),
    });
    const combobox = screen.getByRole("combobox");
    expect(combobox).toBeDisabled();
    expect(combobox).toHaveTextContent("読み込み中");
  });

  it("取得エラー時は汎用メッセージと type=button の再試行を出し、エラーの生文言を表示しない", () => {
    mocks.plans = queryState({
      isSuccess: false,
      isError: true,
      error: new Error("GET /v1/masters/hospitalization-plans 500 db-connection-detail"),
    });
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("入院プランの候補を取得できませんでした");
    expect(alert).not.toHaveTextContent("500");
    expect(alert).not.toHaveTextContent("db-connection-detail");
    const retry = screen.getByRole("button", { name: /再試行/ });
    expect(retry).toHaveAttribute("type", "button");
    expect(retry).toBeEnabled();
  });

  it("再試行ボタンを押すと選択 type のクエリだけ refetch される", async () => {
    const user = userEvent.setup();
    const plansRefetch = vi.fn();
    mocks.plans = queryState({ isSuccess: false, isError: true, refetch: plansRefetch });
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    await user.click(screen.getByRole("button", { name: /再試行/ }));
    expect(plansRefetch).toHaveBeenCalledTimes(1);
    expect(mocks.medicines.refetch).not.toHaveBeenCalled();
    expect(mocks.procedures.refetch).not.toHaveBeenCalled();
  });

  it("エラーからの再取得中(isFetching)は再試行ボタンを disabled にする", () => {
    mocks.plans = queryState({ isSuccess: false, isError: true, isFetching: true });
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    expect(screen.getByRole("button", { name: /再試行/ })).toBeDisabled();
  });

  it("エラー表示は選択 type のクエリだけを追跡し、非選択 type の失敗は表示しない", () => {
    mocks.medicines = queryState({ isSuccess: false, isError: true });
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("クエリがエラー状態なら data=[] のキャッシュがあってもエラー表示し「未登録」とは扱わない", () => {
    mocks.plans = queryState({ data: [], isSuccess: false, isError: true });
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    expect(screen.getByRole("alert")).toHaveTextContent("取得できませんでした");
    expect(screen.queryByText(/登録されている入院プランがありません/)).not.toBeInTheDocument();
  });

  it("成功0件でも再取得中(isFetching)は未登録の説明を確定表示しない", () => {
    mocks.plans = queryState({ data: [], isFetching: true });
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    expect(screen.queryByText(/登録されている入院プランがありません/)).not.toBeInTheDocument();
  });

  it("取得成功・候補0件では未登録の説明と入院プラン設定リンクを出す(canView+canCreate)", () => {
    mocks.plans = queryState({ data: [] });
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("登録されている入院プランがありません");
    expect(status).toHaveTextContent("登録してください");
    expect(within(status).getByRole("link")).toHaveAttribute("href", "/settings/hospitalization");
    expect(mocks.permissionResources).toContain("master-hospitalization");
  });

  it("canView だが canCreate 無しでは設定リンクを出しつつ登録は権限スタッフ依頼と案内する", () => {
    mocks.plans = queryState({ data: [] });
    mocks.permission = { canView: true, canCreate: false, canEdit: false, canDelete: false };
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("登録されている入院プランがありません");
    expect(status).toHaveTextContent("権限のあるスタッフ");
    expect(within(status).getByRole("link")).toHaveAttribute("href", "/settings/hospitalization");
  });

  it("canCreate だけで canView 無しでは設定導線を出さず権限のあるスタッフへの依頼を案内する", () => {
    mocks.plans = queryState({ data: [] });
    mocks.permission = { canView: false, canCreate: true, canEdit: false, canDelete: false };
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("登録されている入院プランがありません");
    expect(status).toHaveTextContent("権限のあるスタッフ");
    expect(status).not.toHaveTextContent("登録してください");
    expect(within(status).queryByRole("link")).not.toBeInTheDocument();
  });

  it("canView 無しでは設定リンクを出さず権限のあるスタッフへの依頼を案内する", () => {
    mocks.plans = queryState({ data: [] });
    mocks.permission = { canView: false, canCreate: false, canEdit: false, canDelete: false };
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("登録されている入院プランがありません");
    expect(status).toHaveTextContent("権限のあるスタッフ");
    expect(within(status).queryByRole("link")).not.toBeInTheDocument();
  });

  it("type=medicine の候補0件では master-medical 権限で薬剤マスタへのリンクを出す", () => {
    mocks.medicines = queryState({ data: [] });
    render(<CarePlanRefSelect type="medicine" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("登録されている薬剤がありません");
    expect(within(status).getByRole("link")).toHaveAttribute("href", "/settings/medicine");
    expect(mocks.permissionResources).toContain("master-medical");
    expect(mocks.permissionResources).not.toContain("master-hospitalization");
  });

  it("type=treatment の候補0件では master-medical 権限で処置マスタへのリンクを出す", () => {
    mocks.procedures = queryState({ data: [] });
    render(<CarePlanRefSelect type="treatment" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("登録されている処置・検査がありません");
    expect(within(status).getByRole("link")).toHaveAttribute(
      "href",
      expect.stringContaining("procedure"),
    );
    expect(mocks.permissionResources).toContain("master-medical");
    expect(mocks.permissionResources).not.toContain("master-hospitalization");
  });

  it("候補がある成功時は error/empty の案内を出さない", () => {
    render(<CarePlanRefSelect type="item" value={null} onChange={vi.fn()} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });

  it("フェッチ状態の遷移だけでは選択値と単価コールバックをリセットしない", () => {
    const handleChange = vi.fn();
    const handlePrice = vi.fn();
    mocks.plans = queryState({ data: [] });
    const props = {
      type: "item" as const,
      value: "3",
      onChange: handleChange,
      onUnitPriceChange: handlePrice,
    };
    const { rerender } = render(<CarePlanRefSelect {...props} />, {
      wrapper: createTestWrapper({ router: true }),
    });
    // error 状態に遷移してもコールバックは発火しない
    mocks.plans = queryState({ isSuccess: false, isError: true });
    rerender(<CarePlanRefSelect {...props} />);
    // 成功・候補有りに復帰しても同様
    mocks.plans = queryState({
      data: [{ id: "3", name: "スタンダード入院プラン", price: 1200 }],
    });
    rerender(<CarePlanRefSelect {...props} />);
    expect(handleChange).not.toHaveBeenCalled();
    expect(handlePrice).not.toHaveBeenCalled();
    expect(screen.getByRole("combobox")).toHaveTextContent("スタンダード入院プラン");
  });
});

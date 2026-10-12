import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/testing/TestUtils";
import { AddForm } from "./AddForm";

// EMR-355: AddForm.test.tsx は CarePlanRefSelect を stub 化しているため、
// マスタ取得状態(loading/error/empty)と実コンポーネントを通した payload を
// 検証できない。ここでは実物の CarePlanRefSelect を使い、マスタ取得 hook と
// 権限 hook だけをテストダブルにする。
const mocks = vi.hoisted(() => ({
  medicines: {} as Record<string, unknown>,
  procedures: {} as Record<string, unknown>,
  plans: {} as Record<string, unknown>,
  permission: { canView: true, canCreate: true, canEdit: true, canDelete: false },
}));

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllMedicinesMaster: () => mocks.medicines,
  useGetAllProcedures: () => mocks.procedures,
  useGetAllHospitalizationPlansMaster: () => mocks.plans,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mocks.permission,
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
});

afterEach(() => {
  vi.clearAllMocks();
});

function renderForm(onSubmit = vi.fn()) {
  render(<AddForm onSubmit={onSubmit} />, {
    wrapper: createTestWrapper({ router: true }),
  });
  return onSubmit;
}

async function selectType(user: ReturnType<typeof userEvent.setup>, label: string) {
  await user.click(screen.getByRole("combobox", { name: "ケアプラン項目種別" }));
  const listbox = await screen.findByRole("listbox");
  await user.click(within(listbox).getByText(label));
}

async function selectPlan(user: ReturnType<typeof userEvent.setup>, label: string) {
  await user.click(screen.getByRole("combobox", { name: "入院プランを選択" }));
  await user.click(await screen.findByText(label));
}

describe("AddForm + 実 CarePlanRefSelect (EMR-355)", () => {
  it("持ち物で登録済み入院プランを選択して追加すると hospitalization_plan_id と unit_price が payload に入る", async () => {
    const user = userEvent.setup();
    const onSubmit = renderForm();

    await selectType(user, "持ち物");
    await user.type(screen.getByPlaceholderText("名称を入力"), "スタンダード入院プラン");
    await selectPlan(user, "スタンダード入院プラン");
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        name: "スタンダード入院プラン",
        hospitalization_plan_id: "3",
        medicine_id: null,
        procedure_id: null,
        unit_price: 1200,
      }),
    );
  });

  it("候補があるのに未選択のまま追加を押すと参照必須エラーで拒否される", async () => {
    const user = userEvent.setup();
    const onSubmit = renderForm();

    await selectType(user, "持ち物");
    await user.type(screen.getByPlaceholderText("名称を入力"), "入院セット");
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(await screen.findByRole("alert")).toHaveTextContent("参照するマスタを選択してください");
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("候補0件のときは必須要求の代わりに未登録の説明が出て、手入力チェックは残る", async () => {
    const user = userEvent.setup();
    mocks.plans = queryState({ data: [] });
    const onSubmit = renderForm();

    await selectType(user, "持ち物");

    expect(await screen.findByRole("status")).toHaveTextContent(
      "登録されている入院プランがありません",
    );
    expect(screen.getByRole("checkbox", { name: /手入力（その他）/ })).toBeInTheDocument();

    // 未選択のまま追加を押しても参照必須エラー自体は維持される（必須の無条件解除ではない）
    await user.type(screen.getByPlaceholderText("名称を入力"), "入院セット");
    await user.click(screen.getByRole("button", { name: /追加/ }));
    expect(await screen.findByRole("alert")).toHaveTextContent("参照するマスタを選択してください");
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("取得エラーのときは汎用エラーと再試行が出て「未登録」扱いにせず、手入力チェックは残る", async () => {
    const user = userEvent.setup();
    mocks.plans = queryState({ isSuccess: false, isError: true });
    renderForm();

    await selectType(user, "持ち物");

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "入院プランの候補を取得できませんでした",
    );
    expect(screen.getByRole("button", { name: /再試行/ })).toBeInTheDocument();
    expect(screen.queryByText(/登録されている入院プランがありません/)).not.toBeInTheDocument();
    expect(screen.getByRole("checkbox", { name: /手入力（その他）/ })).toBeInTheDocument();
  });

  it("手入力で理由が空のまま追加を押すと理由必須で拒否される", async () => {
    const user = userEvent.setup();
    const onSubmit = renderForm();

    await selectType(user, "持ち物");
    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));
    await user.type(screen.getByPlaceholderText("名称を入力"), "持ち込み品");
    await user.type(screen.getByLabelText("手入力の単価"), "800");
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(await screen.findByRole("alert")).toHaveTextContent("その他の理由を入力してください");
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("手入力で単価が不正なら単価エラーで拒否される", async () => {
    const user = userEvent.setup();
    const onSubmit = renderForm();

    await selectType(user, "持ち物");
    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));
    await user.type(screen.getByPlaceholderText("名称を入力"), "持ち込み品");
    await user.type(screen.getByLabelText("その他理由"), "持ち込み品のため");
    fireEvent.change(screen.getByLabelText("手入力の単価"), { target: { value: "-1" } });
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "単価は0以上の整数で入力してください",
    );
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("手入力で名称・理由・単価を満たせば manual=true で追加できる", async () => {
    const user = userEvent.setup();
    const onSubmit = renderForm();

    await selectType(user, "持ち物");
    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));
    await user.type(screen.getByPlaceholderText("名称を入力"), "持ち込み療養食");
    await user.type(screen.getByLabelText("手入力の単価"), "800");
    await user.type(screen.getByLabelText("その他理由"), "持ち込み品のため");
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        name: "持ち込み療養食",
        manual: true,
        other_reason: "持ち込み品のため",
        unit_price: 800,
        hospitalization_plan_id: null,
      }),
    );
  });
});

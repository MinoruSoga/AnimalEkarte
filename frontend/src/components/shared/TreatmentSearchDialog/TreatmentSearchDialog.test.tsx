import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TreatmentSearchDialog } from "./TreatmentSearchDialog";

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllConsultations: vi.fn(),
  useGetAllProcedures: vi.fn(),
  useGetAllVaccinesMaster: vi.fn(),
  useGetAllCheckupTypes: vi.fn(),
  useGetAllMedicinesMaster: vi.fn(),
}));

import {
  useGetAllConsultations,
  useGetAllProcedures,
  useGetAllVaccinesMaster,
  useGetAllCheckupTypes,
  useGetAllMedicinesMaster,
} from "@/hooks/use-treatment-master";

const CONSULTATIONS = [{ id: "1", name: "一般診察", price: 3000, isActive: true }];
const PROCEDURES = [{ id: "2", name: "点滴処置", price: 2000, isActive: true }];
const VACCINES = [{ id: "3", name: "混合ワクチン", price: 5000, isActive: true }];
const CHECKUP_TYPES = [{ id: "4", name: "血液検査", price: 4000, isActive: true }];
// #201: 薬剤カテゴリ。parentId 未設定・price>0 のためカテゴリ見出し行として除外されない。
const MEDICINES = [
  { id: "5", name: "アモキシシリン", price: 100, isActive: true, parentId: "cat-1" },
];

function mockHooksWithData() {
  vi.mocked(useGetAllConsultations).mockReturnValue({
    data: CONSULTATIONS,
  } as ReturnType<typeof useGetAllConsultations>);
  vi.mocked(useGetAllProcedures).mockReturnValue({
    data: PROCEDURES,
  } as ReturnType<typeof useGetAllProcedures>);
  vi.mocked(useGetAllVaccinesMaster).mockReturnValue({
    data: VACCINES,
  } as ReturnType<typeof useGetAllVaccinesMaster>);
  vi.mocked(useGetAllCheckupTypes).mockReturnValue({
    data: CHECKUP_TYPES,
  } as ReturnType<typeof useGetAllCheckupTypes>);
  vi.mocked(useGetAllMedicinesMaster).mockReturnValue({
    data: MEDICINES,
  } as unknown as ReturnType<typeof useGetAllMedicinesMaster>);
}

function renderDialog(overrides?: { onSelect?: (item: unknown) => void }) {
  return render(
    <TreatmentSearchDialog
      open
      onOpenChange={() => {}}
      onSelect={overrides?.onSelect ?? (() => {})}
    />,
  );
}

beforeEach(() => {
  mockHooksWithData();
});

describe("TreatmentSearchDialog", () => {
  it("開くと全カテゴリの治療プランが表示される", () => {
    renderDialog();
    expect(screen.getByText("一般診察")).toBeInTheDocument();
    expect(screen.getByText("点滴処置")).toBeInTheDocument();
    expect(screen.getByText("混合ワクチン")).toBeInTheDocument();
    expect(screen.getByText("血液検査")).toBeInTheDocument();
    expect(screen.getByText("アモキシシリン")).toBeInTheDocument();
  });

  it("#201: 薬剤選択時は onSelect に medicineId が含まれる", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(<TreatmentSearchDialog open onOpenChange={() => {}} onSelect={onSelect} />);

    await user.click(screen.getByText("アモキシシリン"));

    expect(onSelect).toHaveBeenCalledWith(
      expect.objectContaining({ id: "5", medicineId: "5", category: "薬剤" }),
    );
  });

  it("#201: カテゴリ見出し行（parentId なし・price=0）の薬剤は一覧から除外される", () => {
    vi.mocked(useGetAllMedicinesMaster).mockReturnValue({
      data: [{ id: "6", name: "抗生剤カテゴリ", price: 0, isActive: true, parentId: undefined }],
    } as unknown as ReturnType<typeof useGetAllMedicinesMaster>);
    renderDialog();

    expect(screen.queryByText("抗生剤カテゴリ")).not.toBeInTheDocument();
  });

  it("検索語に一致しない場合、空状態メッセージを表示する", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByPlaceholderText("治療プランを検索..."), "存在しないプランZZZ");

    expect(await screen.findByText("該当する治療プランが見つかりません。")).toBeInTheDocument();
    expect(screen.queryByText("一般診察")).not.toBeInTheDocument();
  });

  it("検索語で名前一致のプランのみ表示される", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByPlaceholderText("治療プランを検索..."), "点滴");

    expect(await screen.findByText("点滴処置")).toBeInTheDocument();
    expect(screen.queryByText("一般診察")).not.toBeInTheDocument();
  });

  it("非アクティブな項目は一覧に含まれない", () => {
    vi.mocked(useGetAllConsultations).mockReturnValue({
      data: [{ id: "9", name: "廃止済み診察", price: 1000, isActive: false }],
    } as ReturnType<typeof useGetAllConsultations>);
    renderDialog();

    expect(screen.queryByText("廃止済み診察")).not.toBeInTheDocument();
  });

  it("プランをクリックすると onSelect が呼ばれダイアログを閉じる", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    const onOpenChange = vi.fn();
    render(<TreatmentSearchDialog open onOpenChange={onOpenChange} onSelect={onSelect} />);

    await user.click(screen.getByText("一般診察"));

    expect(onSelect).toHaveBeenCalledWith(expect.objectContaining({ id: "1", name: "一般診察" }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("カテゴリチップで当該カテゴリのみに絞り込まれる", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.click(screen.getByRole("button", { name: "処置" }));

    expect(await screen.findByText("点滴処置")).toBeInTheDocument();
    expect(screen.queryByText("一般診察")).not.toBeInTheDocument();
    expect(screen.queryByText("混合ワクチン")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "解除" })).toBeInTheDocument();
  });

  it("検索後にクリアボタンで全件表示に戻る", async () => {
    const user = userEvent.setup();
    renderDialog();

    const input = screen.getByPlaceholderText("治療プランを検索...");
    await user.type(input, "点滴");
    expect(input).toHaveValue("点滴");
    expect(screen.queryByText("一般診察")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "検索をクリア" }));

    expect(input).toHaveValue("");
    expect(screen.getByText("一般診察")).toBeInTheDocument();
    expect(screen.getByText("点滴処置")).toBeInTheDocument();
  });
});

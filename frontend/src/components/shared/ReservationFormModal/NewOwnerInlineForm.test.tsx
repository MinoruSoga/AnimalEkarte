import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useAnimalSpecies } from "@/hooks/use-animal-species";
import type { NewOwnerFormData } from "@/types/reservation-form";

import { NewOwnerInlineForm } from "./NewOwnerInlineForm";

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: vi.fn(),
}));

const DOG_SPECIES = {
  id: 1,
  name: "犬",
  is_active: true,
  sort_order: 1,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
  label: "犬",
  isInactive: false,
};

const EMPTY_FORM: NewOwnerFormData = {
  ownerName: "",
  phone: "",
  petName: "",
  chiefComplaint: "",
  animalSpeciesId: 0,
};

const mockedUseAnimalSpecies = vi.mocked(useAnimalSpecies);
type AnimalSpeciesState = ReturnType<typeof useAnimalSpecies>;

function mockAnimalSpecies(
  overrides: Partial<AnimalSpeciesState> = {},
) {
  mockedUseAnimalSpecies.mockReturnValue({
    allSpecies: [],
    activeSpecies: [],
    isLoading: false,
    isError: false,
    error: null,
    ...overrides,
  });
}

function renderForm(onChange = vi.fn()) {
  render(
    <NewOwnerInlineForm
      value={EMPTY_FORM}
      onChange={onChange}
      errors={{}}
    />,
  );
  return onChange;
}

describe("NewOwnerInlineForm animal species states", () => {
  beforeEach(() => {
    mockAnimalSpecies();
  });

  it("取得失敗を読み込み中や候補より優先して表示し、生のエラー詳細を隠す", () => {
    mockAnimalSpecies({
      activeSpecies: [DOG_SPECIES],
      allSpecies: [DOG_SPECIES],
      isLoading: true,
      isError: true,
      error: new Error("GET /api/v1/masters/animal-species failed"),
    });

    renderForm();

    expect(screen.getByText("動物種の取得に失敗しました。")).toBeInTheDocument();
    expect(screen.queryByText("動物種を読み込み中です。")).not.toBeInTheDocument();
    expect(screen.queryByText("犬")).not.toBeInTheDocument();
    expect(screen.queryByText(/GET \/api/)).not.toBeInTheDocument();
    expect(screen.getByRole("combobox")).toBeDisabled();
  });

  it("読み込み中を候補より優先して表示し、動物種だけを無効化する", () => {
    mockAnimalSpecies({
      activeSpecies: [DOG_SPECIES],
      allSpecies: [DOG_SPECIES],
      isLoading: true,
    });

    renderForm();

    expect(screen.getByText("動物種を読み込み中です。")).toBeInTheDocument();
    expect(screen.queryByText("動物種の取得に失敗しました。")).not.toBeInTheDocument();
    expect(screen.queryByText("犬")).not.toBeInTheDocument();
    expect(screen.getByRole("combobox")).toBeDisabled();
    expect(screen.getByTestId("new-owner-phone")).toBeEnabled();
  });

  it("取得成功かつ0件では空状態を表示して動物種を無効化する", () => {
    renderForm();

    expect(screen.getByText("動物種マスタが登録されていません。")).toBeInTheDocument();
    expect(screen.queryByText("動物種を読み込み中です。")).not.toBeInTheDocument();
    expect(screen.queryByText("動物種の取得に失敗しました。")).not.toBeInTheDocument();
    expect(screen.getByRole("combobox")).toBeDisabled();
  });

  it("取得成功かつ候補ありでは動物種を選択できる", async () => {
    const user = userEvent.setup();
    mockAnimalSpecies({
      activeSpecies: [DOG_SPECIES],
      allSpecies: [DOG_SPECIES],
    });
    const onChange = renderForm();

    const speciesSelect = screen.getByRole("combobox");
    expect(speciesSelect).toBeEnabled();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();

    await user.click(speciesSelect);
    await user.click(screen.getByRole("option", { name: "犬" }));

    expect(onChange).toHaveBeenCalledWith({
      ...EMPTY_FORM,
      animalSpeciesId: 1,
    });
  });

  it("失敗はalert、読み込み中と空状態はpoliteなstatusとして通知する", () => {
    mockAnimalSpecies({ isError: true, error: new Error("internal detail") });
    const { rerender } = render(
      <NewOwnerInlineForm
        value={EMPTY_FORM}
        onChange={vi.fn()}
        errors={{}}
      />,
    );

    expect(screen.getByRole("alert")).toHaveAttribute("aria-atomic", "true");

    mockAnimalSpecies({ isLoading: true });
    rerender(
      <NewOwnerInlineForm
        value={EMPTY_FORM}
        onChange={vi.fn()}
        errors={{}}
      />,
    );

    expect(screen.getByRole("status")).toHaveAttribute("aria-live", "polite");
    expect(screen.getByRole("status")).toHaveAttribute("aria-atomic", "true");

    mockAnimalSpecies();
    rerender(
      <NewOwnerInlineForm
        value={EMPTY_FORM}
        onChange={vi.fn()}
        errors={{}}
      />,
    );

    expect(screen.getByRole("status")).toHaveAttribute("aria-live", "polite");
    expect(screen.getByRole("status")).toHaveAttribute("aria-atomic", "true");
  });

  it("電話番号エラーを入力へ aria で関連付ける", () => {
    render(
      <NewOwnerInlineForm
        value={EMPTY_FORM}
        onChange={vi.fn()}
        errors={{ phone: "電話番号の形式が正しくありません" }}
      />,
    );

    const phone = screen.getByRole("textbox", { name: "電話番号" });
    expect(phone).toHaveAttribute("aria-invalid", "true");
    expect(phone).toHaveAttribute("aria-describedby", "new-owner-phone-error");
    expect(screen.getByRole("alert")).toHaveAttribute("id", "new-owner-phone-error");
  });

  it("動物種の取得失敗中も動物種以外の入力を利用できる", async () => {
    const user = userEvent.setup();
    mockAnimalSpecies({ isError: true, error: new Error("internal detail") });

    function FormHarness() {
      const [value, setValue] = useState(EMPTY_FORM);
      return (
        <NewOwnerInlineForm
          value={value}
          onChange={setValue}
          errors={{}}
        />
      );
    }

    render(<FormHarness />);

    const ownerNameInput = screen.getByRole("textbox", { name: "飼主名" });
    expect(ownerNameInput).toBeEnabled();
    expect(screen.getByRole("textbox", { name: "電話番号" })).toBeEnabled();
    expect(screen.getByRole("textbox", { name: "ペット名" })).toBeEnabled();
    expect(screen.getByRole("textbox", { name: "主訴" })).toBeEnabled();

    await user.type(ownerNameInput, "山田太郎");

    expect(ownerNameInput).toHaveValue("山田太郎");
    expect(screen.getByRole("combobox")).toBeDisabled();
  });
});

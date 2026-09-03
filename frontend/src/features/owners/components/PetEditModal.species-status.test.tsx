import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { useAnimalSpecies } from "../hooks/use-animal-species";
import { PetEditModal } from "./PetEditModal";

type AnimalSpeciesState = ReturnType<typeof useAnimalSpecies>;

const mocks = vi.hoisted(() => ({
  useAnimalSpecies: vi.fn<() => AnimalSpeciesState>(),
}));

// SearchableSelect は Radix Popover を Radix Dialog の内側で開く。jsdom + カバレッジ計装下では
// Dialog の FocusScope が focus を掴み直し続け、開いた Popover が focus-outside 判定で即座に
// 閉じるため、CI でのみ aria-expanded が false のまま option が現れない。2026-08-23 に CI 上で
// 実測して確認した（click は届いており、body/trigger の pointer-events も disabled も正常。
// fireEvent.click でも開かない）。本テストの対象は Popover の開閉実装ではないので、開閉の
// 意味論だけを保った素の実装へ差し替え、Dialog×Popover の相互作用を構造的に取り除く。
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

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({ data: undefined, isLoading: false, isError: false }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true }),
}));

vi.mock("../hooks/use-animal-species", () => ({
  useAnimalSpecies: mocks.useAnimalSpecies,
}));

vi.mock("../api/get-insurances", () => ({
  useGetInsurances: () => ({ data: [], isLoading: false }),
}));

const animalSpecies = [
  {
    id: 1,
    name: "犬",
    is_active: true,
    sort_order: 1,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    label: "犬",
    isInactive: false,
  },
  {
    id: 2,
    name: "猫",
    is_active: true,
    sort_order: 2,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    label: "猫",
    isInactive: false,
  },
] satisfies AnimalSpeciesState["activeSpecies"];

function createSpeciesState(overrides: Partial<AnimalSpeciesState> = {}): AnimalSpeciesState {
  return {
    allSpecies: animalSpecies,
    activeSpecies: animalSpecies,
    isLoading: false,
    isError: false,
    error: null,
    ...overrides,
  };
}

function renderModal() {
  render(
    <MemoryRouter>
      <PetEditModal open onOpenChange={vi.fn()} ownerName="山田太郎" onSave={vi.fn()} />
    </MemoryRouter>,
  );
}

describe("PetEditModal species status", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useAnimalSpecies.mockReturnValue(createSpeciesState());
  });

  it("取得失敗を最優先の accessible alert で示し、ペット名入力と操作を使える", async () => {
    const rawError = "GET /v1/masters/animal-species: database timeout";
    mocks.useAnimalSpecies.mockReturnValue(
      createSpeciesState({
        allSpecies: [],
        activeSpecies: [],
        isLoading: true,
        isError: true,
        error: new Error(rawError),
      }),
    );
    const user = userEvent.setup();

    renderModal();

    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("動物種の取得に失敗しました。");
    expect(alert).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.queryByText(rawError)).not.toBeInTheDocument();
    const speciesSelect = screen.getByRole("combobox", { name: /動物種/ });
    expect(speciesSelect).toBeDisabled();
    expect(speciesSelect).toHaveTextContent(/^取得に失敗しました$/);

    const petNameInput = screen.getByRole("textbox", { name: /^ペット名 \*$/ });
    await user.type(petNameInput, "ポチ");
    expect(petNameInput).toHaveValue("ポチ");
    expect(screen.getByRole("button", { name: "キャンセル" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "登録" })).toBeEnabled();
  });

  it("読み込み中を空状態より優先して accessible status で示す", () => {
    mocks.useAnimalSpecies.mockReturnValue(
      createSpeciesState({
        allSpecies: [],
        activeSpecies: [],
        isLoading: true,
      }),
    );

    renderModal();

    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("動物種を読み込み中です。");
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText("動物種マスタが登録されていません。")).not.toBeInTheDocument();
    const speciesSelect = screen.getByRole("combobox", { name: /動物種/ });
    expect(speciesSelect).toBeDisabled();
    expect(speciesSelect).toHaveTextContent(/^読み込み中\.\.\.$/);
  });

  it("取得成功かつ0件を distinct accessible status で示す", () => {
    mocks.useAnimalSpecies.mockReturnValue(
      createSpeciesState({
        allSpecies: [],
        activeSpecies: [],
      }),
    );

    renderModal();

    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("動物種マスタが登録されていません。");
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText("動物種を読み込み中です。")).not.toBeInTheDocument();
    const speciesSelect = screen.getByRole("combobox", { name: /動物種/ });
    expect(speciesSelect).toBeDisabled();
    expect(speciesSelect).toHaveTextContent(/^登録されていません$/);
  });

  it("取得成功かつ候補ありでは状態表示を消して動物種を選べる", async () => {
    const user = userEvent.setup({ delay: null });
    renderModal();

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();

    const speciesSelect = screen.getByRole("combobox", { name: /動物種/ });
    expect(speciesSelect).toBeEnabled();
    await user.click(speciesSelect);
    expect(await screen.findByRole("option", { name: "犬" })).toBeInTheDocument();
    await user.click(await screen.findByRole("option", { name: "猫" }));
    expect(speciesSelect).toHaveTextContent("猫");
  });
});

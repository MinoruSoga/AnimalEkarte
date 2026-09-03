import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { createTestWrapper } from "@/testing/TestUtils";
import { useGetOwnerSharedPets } from "../api/get-owner-shared-pets";
import type { PetFormData } from "../types";
import { OwnerPetsSection } from "./OwnerPetsSection";

vi.mock("../api/get-owner-shared-pets", () => ({
  useGetOwnerSharedPets: vi.fn(),
}));

const mockedUseGetOwnerSharedPets = vi.mocked(useGetOwnerSharedPets);

const PET = {
  id: "pet-1",
  petNumber: "P001",
  petName: "ポチ",
  status: "生存",
  species: "犬",
  gender: "雄",
  birthDate: "2020-01-01",
  color: "茶",
  weight: "5",
  environment: "室内",
  remarks: "",
} satisfies PetFormData;

interface RenderSectionOptions {
  pet?: PetFormData;
  pets?: PetFormData[];
  ownerId?: string;
  canEdit?: boolean;
  canCreate?: boolean;
  canDelete?: boolean;
  onAddPet?: () => void;
  onEditPet?: (pet: PetFormData) => void;
  onDeleteRequest?: (id: string, name: string) => void;
}

function renderSection({
  pet = PET,
  pets,
  ownerId = "owner-1",
  canEdit = true,
  canCreate = false,
  canDelete = false,
  onAddPet = vi.fn(),
  onEditPet = vi.fn(),
  onDeleteRequest = vi.fn(),
}: RenderSectionOptions = {}) {
  render(
    <OwnerPetsSection
      pets={pets ?? [pet]}
      ownerId={ownerId}
      canEdit={canEdit}
      canCreate={canCreate}
      canDelete={canDelete}
      onAddPet={onAddPet}
      onEditPet={onEditPet}
      onDeleteRequest={onDeleteRequest}
    />,
    { wrapper: createTestWrapper({ router: true }) },
  );
  return { onAddPet, onEditPet, onDeleteRequest };
}

describe("OwnerPetsSection row actions", () => {
  beforeEach(() => {
    mockedUseGetOwnerSharedPets.mockReset();
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: { shared_pets: [] },
      isError: false,
      isLoading: false,
    } as ReturnType<typeof useGetOwnerSharedPets>);
  });

  it("行は編集せず、権限時だけ固有名の44px native buttonで編集する", async () => {
    const user = userEvent.setup();
    const { onEditPet } = renderSection();

    await user.click(screen.getByText("犬"));
    expect(onEditPet).not.toHaveBeenCalled();

    const editButton = screen.getByRole("button", {
      name: "詳細・編集: ペット ポチ (ID pet-1)",
    });
    expect(editButton.tagName).toBe("BUTTON");
    expect(editButton).toHaveClass("min-h-11", "min-w-11");
    await user.click(editButton);
    expect(onEditPet).toHaveBeenCalledWith(PET);
  });

  it("編集権限がなければ編集buttonを表示しない", () => {
    renderSection({ canEdit: false });
    expect(
      screen.queryByRole("button", { name: "詳細・編集: ペット ポチ (ID pet-1)" }),
    ).not.toBeInTheDocument();
  });

  it("作成権限があれば編集権限なしでもペット追加buttonを表示する", () => {
    renderSection({ canEdit: false, canCreate: true });

    expect(screen.getByRole("button", { name: "ペット追加" })).toBeInTheDocument();
  });

  it("生存ペットは既存の5つの新規業務遷移と削除操作を表示する", async () => {
    const user = userEvent.setup();
    renderSection({ canCreate: true, canDelete: true });

    await user.click(
      screen.getByRole("button", {
        name: "操作メニュー: ペット ポチ (ID pet-1)",
      }),
    );

    for (const name of ["予約作成", "カルテ作成", "トリミング", "入院登録", "会計登録", "削除"]) {
      expect(screen.getByRole("menuitem", { name })).toBeInTheDocument();
    }
  });

  it("死亡ペットは編集を維持しつつ5つの新規業務遷移と破壊的削除を表示しない", async () => {
    const user = userEvent.setup();
    renderSection({
      pet: { ...PET, status: "死亡", deceasedAt: "2026-07-11T00:00:00+09:00" },
      canCreate: true,
      canDelete: true,
    });

    await user.click(
      screen.getByRole("button", {
        name: "操作メニュー: ペット ポチ (ID pet-1)",
      }),
    );

    expect(screen.getByRole("menuitem", { name: "詳細・編集" })).toBeInTheDocument();
    for (const name of ["予約作成", "カルテ作成", "トリミング", "入院登録", "会計登録", "削除"]) {
      expect(screen.queryByRole("menuitem", { name })).not.toBeInTheDocument();
    }
  });

  it("status不明のペットは明示表示し、編集以外の業務遷移と削除をfail-closedにする", async () => {
    const user = userEvent.setup();
    renderSection({
      pet: { ...PET, status: "不明" },
      canCreate: true,
      canDelete: true,
    });

    expect(screen.getByText("不明")).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", {
        name: "操作メニュー: ペット ポチ (ID pet-1)",
      }),
    );

    expect(screen.getByRole("menuitem", { name: "詳細・編集" })).toBeInTheDocument();
    for (const name of ["予約作成", "カルテ作成", "トリミング", "入院登録", "会計登録", "削除"]) {
      expect(screen.queryByRole("menuitem", { name })).not.toBeInTheDocument();
    }
  });
});

describe("OwnerPetsSection shared pets", () => {
  beforeEach(() => {
    mockedUseGetOwnerSharedPets.mockReset();
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: { shared_pets: [] },
      isError: false,
      isLoading: false,
    } as ReturnType<typeof useGetOwnerSharedPets>);
  });

  it("own行の後にshared行を表示し、関係あり/なしのBadgeと既知・未知ラベルを表現する", () => {
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: {
        shared_pets: [
          {
            id: 2,
            pet_number: "P002",
            name: "ハナ",
            status: "deceased",
            gender: "male",
            animal_species: { name: "犬" },
            birth_date: "2021-02-03",
            color: "白",
            weight: 8.5,
            environment: "室内",
            remarks: "共有",
            relationship: "妻",
          },
          {
            id: 3,
            pet_number: "P003",
            name: "ミケ",
            status: "archived",
            gender: "other",
            animal_species: { name: "猫" },
            birth_date: null,
            color: "三毛",
            weight: null,
            environment: "室内",
            remarks: "",
            relationship: "",
          },
        ],
      },
      isError: false,
      isLoading: false,
    } as ReturnType<typeof useGetOwnerSharedPets>);

    renderSection();

    const ownRow = screen.getByText("ポチ").closest("tr");
    const sharedRow = screen.getByText("ハナ").closest("tr");
    const fallbackRow = screen.getByText("ミケ").closest("tr");
    if (!ownRow || !sharedRow || !fallbackRow) {
      throw new Error("expected pet rows");
    }
    expect(
      ownRow.compareDocumentPosition(sharedRow) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    expect(within(sharedRow).getByText("ハナ")).not.toHaveAttribute("data-slot", "badge");
    expect(within(sharedRow).getByText("副飼主")).toHaveAttribute("data-slot", "badge");
    expect(within(sharedRow).getByText("妻")).toBeInTheDocument();
    expect(within(sharedRow).getByText("死亡")).toBeInTheDocument();
    expect(within(sharedRow).getByText("雄")).toBeInTheDocument();
    expect(within(fallbackRow).getByText("ミケ")).not.toHaveAttribute("data-slot", "badge");
    expect(within(fallbackRow).getByText("副飼主")).toHaveAttribute("data-slot", "badge");
    expect(within(fallbackRow).getAllByRole("cell")[1]).toHaveTextContent(/^ミケ副飼主$/);
    expect(within(fallbackRow).getByText("不明")).toBeInTheDocument();
    expect(within(fallbackRow).getByText("other")).toBeInTheDocument();
    expect(
      within(fallbackRow)
        .getAllByRole("cell")
        .map((cell) => cell.textContent),
    ).toEqual(["P003", "ミケ副飼主", "不明", "猫", "other", "", "三毛", "", "室内", "", ""]);
  });

  it("shared行は読み取り専用で操作とown callbackを持たない", async () => {
    const user = userEvent.setup();
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: {
        shared_pets: [
          {
            id: 2,
            pet_number: "P002",
            name: "ハナ",
            status: "alive",
            gender: "female",
            animal_species: { name: "犬" },
            birth_date: null,
            color: "白",
            weight: null,
            environment: "室内",
            remarks: "",
            relationship: "妻",
          },
        ],
      },
      isError: false,
      isLoading: false,
    } as ReturnType<typeof useGetOwnerSharedPets>);
    const onAddPet = vi.fn();
    const onEditPet = vi.fn();
    const onDeleteRequest = vi.fn();

    renderSection({
      pets: [],
      canCreate: true,
      canEdit: true,
      canDelete: true,
      onAddPet,
      onEditPet,
      onDeleteRequest,
    });

    const sharedRow = screen.getByText("ハナ").closest("tr");
    if (!sharedRow) {
      throw new Error("expected shared pet row");
    }
    expect(within(sharedRow).queryAllByRole("button")).toHaveLength(0);
    await user.click(within(sharedRow).getByText("ハナ"));
    expect(onAddPet).not.toHaveBeenCalled();
    expect(onEditPet).not.toHaveBeenCalled();
    expect(onDeleteRequest).not.toHaveBeenCalled();
  });

  it("ownが0件でもsharedがあればempty stateを表示しない", () => {
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: {
        shared_pets: [
          {
            id: 2,
            pet_number: "P002",
            name: "ハナ",
            status: "alive",
            gender: "female",
            animal_species: { name: "犬" },
            birth_date: null,
            color: "",
            weight: null,
            environment: "",
            remarks: "",
            relationship: "",
          },
        ],
      },
      isError: false,
      isLoading: false,
    } as ReturnType<typeof useGetOwnerSharedPets>);

    renderSection({ pets: [] });

    expect(screen.getByText("ハナ")).toBeInTheDocument();
    expect(
      screen.queryByText("ペット情報がありません。「ペット追加」ボタンから追加してください。"),
    ).not.toBeInTheDocument();
  });

  it("ownとsettled sharedがともに0件ならempty stateを表示する", () => {
    renderSection({ pets: [] });

    expect(
      screen.getByText("ペット情報がありません。「ペット追加」ボタンから追加してください。"),
    ).toBeInTheDocument();
  });

  it("shared取得エラーをsection内で通知してもown行を維持する", () => {
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: undefined,
      isError: true,
      isLoading: false,
    } as ReturnType<typeof useGetOwnerSharedPets>);

    renderSection();

    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent("共有ペット情報の取得に失敗しました。");
    expect(
      screen.queryByText("ペット情報がありません。「ペット追加」ボタンから追加してください。"),
    ).not.toBeInTheDocument();
  });
});

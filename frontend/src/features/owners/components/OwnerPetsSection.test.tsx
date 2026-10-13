import { Children } from "react";
import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { createTestWrapper } from "@/testing/TestUtils";
import { DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { useGetOwnerSharedPets } from "../api/get-owner-shared-pets";
import type { OwnerSharedPetsResponse } from "../api/get-owner-shared-pets";
import type { PetFormData } from "../types";
import { OwnerPetsSection } from "./OwnerPetsSection";

const navigateMock = vi.fn();
vi.mock("@/components/ui/dropdown-menu", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/components/ui/dropdown-menu")>();
  return { ...actual, DropdownMenuItem: vi.fn(actual.DropdownMenuItem) };
});
vi.mock("react-router", async (importOriginal) => ({
  ...(await importOriginal<typeof import("react-router")>()),
  useNavigate: () => navigateMock,
}));

vi.mock("../api/get-owner-shared-pets", () => ({
  useGetOwnerSharedPets: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: reservationPermission, canView: reservationViewPermission }),
}));
let reservationPermission = false;
let reservationViewPermission = true;

const mockedUseGetOwnerSharedPets = vi.mocked(useGetOwnerSharedPets);

// useGetOwnerSharedPets の戻り値は UseQueryResult<OwnerSharedPetsResponse, Error>。
// data は実レスポンス型で揃える（空の shared_pets も OwnerSharedPetApiResponse[] として型付け）。
const EMPTY_SHARED_PETS: OwnerSharedPetsResponse = { shared_pets: [] };

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
  const result = render(
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
  return { ...result, onAddPet, onEditPet, onDeleteRequest };
}

describe("OwnerPetsSection row actions", () => {
  beforeEach(() => {
    reservationPermission = false;
    mockedUseGetOwnerSharedPets.mockReset();
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: EMPTY_SHARED_PETS,
      isError: false,
      isLoading: false,
    } as ReturnType<typeof useGetOwnerSharedPets>);
  });

  it("行は編集せず、権限時だけ操作メニューの詳細・編集から編集する", async () => {
    const user = userEvent.setup();
    const { onEditPet } = renderSection();

    await user.click(screen.getByText("犬"));
    expect(onEditPet).not.toHaveBeenCalled();

    // EMR-174: ペット名は deep link に変更。直接編集は操作メニューの詳細・編集で行う。
    await user.click(
      screen.getByRole("button", {
        name: "操作メニュー: ペット ポチ (ID pet-1)",
      }),
    );
    await user.click(screen.getByRole("menuitem", { name: "詳細・編集" }));
    expect(onEditPet).toHaveBeenCalledWith(PET);
  });

  it("編集権限がなければ操作メニューに編集項目を表示しない", async () => {
    const user = userEvent.setup();
    renderSection({ canEdit: false });
    await user.click(
      screen.getByRole("button", {
        name: "操作メニュー: ペット ポチ (ID pet-1)",
      }),
    );
    expect(screen.queryByRole("menuitem", { name: "詳細・編集" })).not.toBeInTheDocument();
  });

  it("作成権限があれば編集権限なしでもペット追加buttonを表示する", () => {
    renderSection({ canEdit: false, canCreate: true });

    expect(screen.getByRole("button", { name: "ペット追加" })).toBeInTheDocument();
  });

  // F004: 右下 Support FAB の占有幅（right-4 16px + size-12 48px =
  // SUPPORT_WIDGET_LAYOUT.footprintPx 64px）がページ右端のヘッダアクションと
  // 重なる。ボタン右側に同等のクリアスペースを確保する（PageLayout pb-20 は
  // スクロール末端のみ保護し、水平方向の重なりは防がない）。
  it("ペット追加ボタンは右下 FAB 占有幅ぶんの右マージンを確保する（F004）", () => {
    renderSection({ canCreate: true });

    expect(screen.getByRole("button", { name: "ペット追加" })).toHaveClass("mr-16");
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

  it.each([
    { status: "死亡", deceasedAt: undefined },
    { status: "生存", deceasedAt: "2026-07-11T00:00:00+09:00" },
    { status: "生存", deceasedAt: "" },
  ])("死亡sentinel %o では編集を維持しつつ新規業務遷移と削除を表示しない", async (sentinel) => {
    const user = userEvent.setup();
    renderSection({
      pet: { ...PET, ...sentinel },
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

  it.each(
    ["予約作成", "カルテ作成", "トリミング", "入院登録", "会計登録", "削除"].flatMap((name) => [
      { name, revoked: false },
      { name, revoked: true },
    ]),
  )(
    "$name の保持済みcallbackも最新の死亡日時または権限取消（$revoked）を見て拒否する",
    async ({ name, revoked }) => {
      const user = userEvent.setup();
      navigateMock.mockReset();
      vi.mocked(DropdownMenuItem).mockClear();
      const onEditPet = vi.fn();
      const onDeleteRequest = vi.fn();
      const { rerender } = renderSection({
        canCreate: true,
        canDelete: true,
        onEditPet,
        onDeleteRequest,
      });
      await user.click(
        screen.getByRole("button", { name: "操作メニュー: ペット ポチ (ID pet-1)" }),
      );
      const oldHandler = vi
        .mocked(DropdownMenuItem)
        .mock.calls.toReversed()
        .find(([props]) => Children.toArray(props.children).includes(name))?.[0].onClick;
      if (!oldHandler) throw new Error(`missing action callback: ${name}`);

      rerender(
        <OwnerPetsSection
          pets={[revoked ? PET : { ...PET, deceasedAt: "2026-10-11T00:00:00+09:00" }]}
          ownerId="owner-1"
          canEdit
          canCreate={!revoked}
          canDelete={!revoked}
          onAddPet={vi.fn()}
          onEditPet={onEditPet}
          onDeleteRequest={onDeleteRequest}
        />,
      );
      // React dispatch supplies the retained handler with a real synthetic event.
      render(
        <div onClick={oldHandler}>
          <button type="button">保持済み操作</button>
        </div>,
      );
      fireEvent.click(screen.getByRole("button", { name: "保持済み操作" }));
      expect(navigateMock).not.toHaveBeenCalled();
      expect(onDeleteRequest).not.toHaveBeenCalled();
      expect(screen.queryByRole("menuitem", { name })).not.toBeInTheDocument();
    },
  );

  it.each([
    ["予約作成", "/reservations"],
    ["カルテ作成", "/medical-records/new"],
    ["トリミング", "/trimming/new"],
    ["入院登録", "/hospitalization/new"],
    ["会計登録", "/accounting/new"],
    ["削除", null],
  ])("生存ペットの %s は既存の操作を実行する", async (name, path) => {
    const user = userEvent.setup();
    navigateMock.mockReset();
    const { onDeleteRequest } = renderSection({ canCreate: true, canDelete: true });
    await user.click(screen.getByRole("button", { name: "操作メニュー: ペット ポチ (ID pet-1)" }));
    await user.click(screen.getByRole("menuitem", { name: name ?? "" }));
    if (path) {
      expect(navigateMock).toHaveBeenCalledWith(`${path}?petId=pet-1`, {
        state: { from: "/owners/owner-1" },
      });
      expect(onDeleteRequest).not.toHaveBeenCalled();
    } else {
      expect(onDeleteRequest).toHaveBeenCalledWith("pet-1", "ポチ");
      expect(navigateMock).not.toHaveBeenCalled();
    }
  });

  it("死亡日時の解除後は生存ペットの操作が戻り、詳細編集も維持する", async () => {
    const user = userEvent.setup();
    const onEditPet = vi.fn();
    const onDeleteRequest = vi.fn();
    const { rerender } = renderSection({
      pet: { ...PET, deceasedAt: "2026-10-11T00:00:00+09:00" },
      canCreate: true,
      canDelete: true,
      onEditPet,
      onDeleteRequest,
    });
    rerender(
      <OwnerPetsSection
        pets={[{ ...PET, deceasedAt: null }]}
        ownerId="owner-1"
        canEdit
        canCreate
        canDelete
        onAddPet={vi.fn()}
        onEditPet={onEditPet}
        onDeleteRequest={onDeleteRequest}
      />,
    );
    await user.click(screen.getByRole("button", { name: "操作メニュー: ペット ポチ (ID pet-1)" }));
    for (const name of ["予約作成", "カルテ作成", "トリミング", "入院登録", "会計登録", "削除"]) {
      expect(screen.getByRole("menuitem", { name })).toBeInTheDocument();
    }
    await user.click(screen.getByRole("menuitem", { name: "詳細・編集" }));
    expect(onEditPet).toHaveBeenCalledWith({ ...PET, deceasedAt: null });
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

// EMR-174: ペット名クリックで飼主詳細のペットモーダル deep link (?pet=) へ遷移する。
describe("OwnerPetsSection pet detail link (EMR-174)", () => {
  beforeEach(() => {
    reservationPermission = false;
    mockedUseGetOwnerSharedPets.mockReset();
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: EMPTY_SHARED_PETS,
      isError: false,
      isLoading: false,
    } as ReturnType<typeof useGetOwnerSharedPets>);
  });

  it("永続化済みペット名は飼主詳細 ?pet= deep link になる（編集権限あり）", () => {
    renderSection({ canEdit: true });

    const link = screen.getByRole("link", { name: /ポチ/ });
    expect(link).toHaveAttribute("href", "/owners/owner-1?pet=pet-1");
  });

  it("閲覧のみのユーザーにもペット名の detail link を表示する", () => {
    renderSection({ canEdit: false });

    const link = screen.getByRole("link", { name: /ポチ/ });
    expect(link).toHaveAttribute("href", "/owners/owner-1?pet=pet-1");
    // mutation 操作は引き続き出さない
    expect(
      screen.queryByRole("button", { name: /詳細・編集: ペット ポチ/ }),
    ).not.toBeInTheDocument();
  });

  it("pending (temp-*) ペットは deep link にしない（ローカル編集ボタンのみ）", async () => {
    const user = userEvent.setup();
    const { onEditPet } = renderSection({
      pet: { ...PET, id: "temp-1710000000000", isPending: true },
      ownerId: undefined,
    });

    expect(screen.queryByRole("link", { name: /ポチ/ })).not.toBeInTheDocument();
    const editButton = screen.getByRole("button", { name: /詳細・編集: ペット ポチ/ });
    await user.click(editButton);
    expect(onEditPet).toHaveBeenCalled();
  });
});

describe("OwnerPetsSection shared pets", () => {
  beforeEach(() => {
    reservationPermission = false;
    mockedUseGetOwnerSharedPets.mockReset();
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: EMPTY_SHARED_PETS,
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

describe("EMR-318 reception entry", () => {
  beforeEach(() => {
    reservationPermission = false;
    reservationViewPermission = true;
    navigateMock.mockReset();
  });
  it("owner create alone does not expose reception", () => {
    renderSection({ pet: { ...PET, id: "10" }, canCreate: true });
    expect(screen.queryByRole("button", { name: "受付: ポチ" })).not.toBeInTheDocument();
  });
  it("reservation create and view expose reception independently from owner create", () => {
    reservationPermission = true;
    renderSection({ pet: { ...PET, id: "10" }, canCreate: false });
    expect(screen.getByRole("button", { name: "受付: ポチ" })).toBeInTheDocument();
  });
  it("reservation create without view does not expose an inaccessible reception flow", () => {
    reservationPermission = true;
    reservationViewPermission = false;
    renderSection({ pet: { ...PET, id: "10" }, canCreate: false });

    expect(screen.queryByRole("button", { name: "受付: ポチ" })).not.toBeInTheDocument();
    expect(navigateMock).not.toHaveBeenCalled();
  });
  it("opens the prefilled reception flow and returns to today's board", async () => {
    reservationPermission = true;
    renderSection({ pet: { ...PET, id: "10" }, canCreate: false });
    await userEvent.click(screen.getByRole("button", { name: "受付: ポチ" }));
    expect(navigateMock).toHaveBeenCalledWith("/reservations?reception=1&petId=10", {
      state: { from: "/" },
    });
  });
  it.each([
    { ...PET, id: "10", status: "死亡" },
    { ...PET, id: "10", deceasedAt: "2026-10-08T00:00:00+09:00" },
    { ...PET, id: "10", deceasedAt: "" },
    { ...PET, id: "temp-1", isPending: true },
  ])("blocks deceased, inconsistent and unsaved pets", (pet) => {
    reservationPermission = true;
    renderSection({ pet });
    expect(screen.queryByRole("button", { name: "受付: ポチ" })).not.toBeInTheDocument();
  });
});

describe("OwnerPetsSection 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
    mockedUseGetOwnerSharedPets.mockReset();
    mockedUseGetOwnerSharedPets.mockReturnValue({
      data: EMPTY_SHARED_PETS,
      isError: false,
      isLoading: false,
    } as ReturnType<typeof useGetOwnerSharedPets>);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月のペット行はペット名セルに Cake バッジを表示する", () => {
    renderSection({ pet: { ...PET, birthDate: "2020-10-05" } });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2020-09-05"],
    ["birthDate 未登録", ""],
  ])("%s のペットには Cake バッジを表示しない", (_label, birthDate) => {
    renderSection({ pet: { ...PET, birthDate } });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=死亡）には誕生月でも Cake バッジを表示しない", () => {
    renderSection({ pet: { ...PET, status: "死亡", birthDate: "2020-10-05" } });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("副飼主ペット行にも誕生月なら Cake バッジを表示する（死亡は抑制）", () => {
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
            birth_date: "2021-10-03",
            color: "白",
            weight: 8.5,
            environment: "室内",
            remarks: "",
            relationship: "妻",
          },
          {
            id: 3,
            pet_number: "P003",
            name: "ミケ",
            status: "deceased",
            gender: "other",
            animal_species: { name: "猫" },
            birth_date: "2020-10-11",
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

    renderSection({ pets: [] });

    const hanaRow = screen.getByText("ハナ").closest("tr");
    const mikeRow = screen.getByText("ミケ").closest("tr");
    if (!hanaRow || !mikeRow) throw new Error("expected shared pet rows");
    expect(within(hanaRow).getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
    expect(within(mikeRow).queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});

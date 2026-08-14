import { act, render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { PetCareSection } from "./PetCareSection";
import type { PetFormData } from "../types";

const { mockMutateAsync, mockMutate } = vi.hoisted(() => ({
  mockMutateAsync: vi.fn(),
  mockMutate: vi.fn(),
}));

vi.mock("@/hooks/use-revoke-pet-death", () => ({
  useRevokePetDeath: () => ({ mutate: mockMutate, isPending: false }),
}));

vi.mock("@/hooks/use-record-pet-death", () => ({
  useRecordPetDeath: () => ({ mutateAsync: mockMutateAsync }),
}));

const basePet: PetFormData = {
  id: "7",
  petNumber: "42-1",
  petName: "ポチ",
  petNameKana: "ぽち",
  status: "生存",
  species: "犬",
  animalSpeciesId: "1",
  gender: "雄",
  birthDate: "2015-04-14",
  breed: "柴犬",
  color: "赤",
  weight: "7.35",
  acquisitionType: "購入",
  dangerLevel: "中",
  food: "療法食",
  environment: "室内",
  remarks: "咬傷注意",
  insuranceId: "",
};

function renderPetCareSection(
  formData: PetFormData,
  setFormData: (updater: (prev: PetFormData) => PetFormData) => void = () => {},
  onPetLifecycleChange?: (result: {
    petId: string;
    status: "死亡" | "生存";
    deceasedAt: string | null;
    deceasedReason?: string | null;
  }) => void,
) {
  return render(
    <MemoryRouter>
      <PetCareSection
        formData={formData}
        setFormData={setFormData}
        insuranceSelectItems={null}
        isLoadingInsurances={false}
        canEdit
        onInsuranceChange={() => {}}
        onPetLifecycleChange={onPetLifecycleChange}
      />
    </MemoryRouter>,
  );
}

// PR#186 P2-2 Bug#1 回帰テスト: 死亡記録された pet の deceased_at が response DTO で
// serialize されなかったため、formData.status === "死亡" のとき常に deceasedAt={null} が
// 渡され、死亡記録の閲覧・解除 UI が一切表示されなかった。
describe("PetCareSection (PR#186 P2-2 Bug#1)", () => {
  it("死亡ステータス + deceasedAt ありのとき、実際の永眠日を含む死亡記録バナーを表示する", () => {
    renderPetCareSection({
      ...basePet,
      status: "死亡",
      deceasedAt: "2026-07-10T12:00:00+09:00",
    });

    expect(screen.getByText(/2026年7月10日 永眠/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "死亡記録を解除" })).toBeInTheDocument();
  });

  it("生存ステータスのとき、死亡を記録するボタンを表示する（バナーではない）", () => {
    renderPetCareSection({ ...basePet, status: "生存", deceasedAt: null });

    expect(screen.getByRole("button", { name: "死亡を記録" })).toBeInTheDocument();
    expect(screen.queryByText(/永眠/)).not.toBeInTheDocument();
  });

  it("BUG-022: pending (temp-*) では死亡記録ボタンを出さない", () => {
    renderPetCareSection({
      ...basePet,
      id: "temp-1710000000000",
      isPending: true,
      status: "生存",
      deceasedAt: null,
    });

    expect(screen.queryByRole("button", { name: "死亡を記録" })).not.toBeInTheDocument();
  });

  it("死亡ステータスだが deceasedAt が未取得のとき、不整合を表示して再登録導線を閉じる", () => {
    renderPetCareSection({ ...basePet, status: "死亡", deceasedAt: undefined });

    expect(
      screen.getByText(
        "生死データに不整合があります（死亡ステータス・死亡日時未登録）。修復は管理者に依頼してください",
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "死亡を記録" }),
    ).not.toBeInTheDocument();
  });
});

// BUG-407: サブダイアログ確定/解除は既にバックエンドへ即時保存されるが、外側
// PetEditModal のローカル formData（生死ラジオ・deceasedAt）へ同期する配線が
// 無いと、次に外側「更新」を押した際に古い status で上書きされる。
// PetCareSection がその同期配線（setFormData 呼び出し）を担うことを固定する。
describe("PetCareSection (BUG-407 outer-form sync)", () => {
  beforeEach(() => {
    mockMutateAsync.mockReset();
    mockMutate.mockReset();
  });

  it("死亡記録成功時に setFormData で status/deceasedAt を即時同期する", async () => {
    mockMutateAsync.mockResolvedValueOnce(undefined);
    let latestFormData: PetFormData = { ...basePet, status: "生存", deceasedAt: null };
    const setFormData = vi.fn((updater: (prev: PetFormData) => PetFormData) => {
      latestFormData = updater(latestFormData);
    });

    renderPetCareSection(latestFormData, setFormData);

    fireEvent.click(screen.getByRole("button", { name: "死亡を記録" }));
    fireEvent.click(screen.getByRole("button", { name: "死亡を記録する" }));

    await waitFor(() => expect(setFormData).toHaveBeenCalled());
    expect(latestFormData.status).toBe("死亡");
    expect(latestFormData.deceasedAt).toBeTruthy();
  });

  it("死亡記録解除成功時に setFormData で status を生存へ即時同期する", async () => {
    mockMutate.mockImplementation((_petId, options) => {
      options?.onSuccess?.();
      options?.onSettled?.();
    });
    let latestFormData: PetFormData = {
      ...basePet,
      status: "死亡",
      deceasedAt: "2026-07-10T12:00:00+09:00",
    };
    const setFormData = vi.fn((updater: (prev: PetFormData) => PetFormData) => {
      latestFormData = updater(latestFormData);
    });

    renderPetCareSection(latestFormData, setFormData);

    fireEvent.click(screen.getByRole("button", { name: "死亡記録を解除" }));
    fireEvent.click(screen.getByRole("button", { name: "解除する" }));

    await waitFor(() => expect(setFormData).toHaveBeenCalled());
    expect(latestFormData.status).toBe("生存");
    expect(latestFormData.deceasedAt).toBeNull();
  });
});

// BUG-409: 生死ラジオ(79-104行目)が deceasedAt と独立に status のみを書き換えられると、
// 外側フォームの「更新」保存で status=死亡 かつ deceasedAt=null（監査ログ・deceasedReason なし）
// という不整合が生じる。生死の変更は監査付きの PetDeceasedRecordButton 経由のみに一本化し、
// ラジオは現在値の表示専用（クリックしても setFormData を呼ばない）でなければならない。
describe("PetCareSection (BUG-409 生死ラジオは二重管理の書込元にならない)", () => {
  it("生存ペットで「死亡」ラジオをクリックしても setFormData は呼ばれない（deceasedAt 無しの不整合書込を防止）", () => {
    const setFormData = vi.fn();
    renderPetCareSection({ ...basePet, status: "生存", deceasedAt: null }, setFormData);

    fireEvent.click(screen.getByRole("radio", { name: "死亡" }));

    expect(setFormData).not.toHaveBeenCalled();
  });

  it("死亡ペットで「生存」ラジオをクリックしても setFormData は呼ばれない（deceasedAt 残存の不整合書込を防止）", () => {
    const setFormData = vi.fn();
    renderPetCareSection(
      { ...basePet, status: "死亡", deceasedAt: "2026-07-10T12:00:00+09:00" },
      setFormData,
    );

    fireEvent.click(screen.getByRole("radio", { name: "生存" }));

    expect(setFormData).not.toHaveBeenCalled();
  });

  // typescript-reviewer 指摘: onChange 不在の間接証拠だけでなく disabled 状態そのものを
  // 直接検証する（将来 disabled だけ外され別経路の書込が復活しても検知できるように）。
  it("両方のラジオが disabled であり、DOM上も編集不能であることを直接確認する", () => {
    renderPetCareSection({ ...basePet, status: "生存", deceasedAt: null });

    expect(screen.getByRole("radio", { name: "生存" })).toBeDisabled();
    expect(screen.getByRole("radio", { name: "死亡" })).toBeDisabled();
  });
});

// BUG-002 focused coverage: 同ファイルの field onChange を最小行使する
describe("PetCareSection (BUG-002 field coverage)", () => {
  it("BUG-002 食べ物・飼育環境・備考の入力で setFormData が呼ばれる", () => {
    let latestFormData: PetFormData = {
      ...basePet,
      id: "pet-synth-1",
      petName: "合成ペット甲",
      food: "",
      environment: "",
      remarks: "",
    };
    const setFormData = vi.fn((updater: (prev: PetFormData) => PetFormData) => {
      latestFormData = updater(latestFormData);
    });

    renderPetCareSection(latestFormData, setFormData);

    fireEvent.change(screen.getByLabelText("食べ物"), {
      target: { value: "合成フード" },
    });
    fireEvent.change(screen.getByLabelText("飼育環境"), {
      target: { value: "合成環境" },
    });
    fireEvent.change(screen.getByLabelText("備考・特記事項"), {
      target: { value: "合成備考" },
    });

    expect(setFormData).toHaveBeenCalled();
    expect(latestFormData.food).toBe("合成フード");
    expect(latestFormData.environment).toBe("合成環境");
    expect(latestFormData.remarks).toBe("合成備考");
  });
});

// BUG-002: モーダル内 setFormData 同期に加え、外側 pets 一覧 owner へ
// 成功時のみ onPetLifecycleChange を1回通知する。失敗時は通知しない。
describe("PetCareSection (BUG-002 outer-list lifecycle notify)", () => {
  beforeEach(() => {
    mockMutateAsync.mockReset();
    mockMutate.mockReset();
  });

  it("BUG-002 死亡記録成功時に onPetLifecycleChange を petId/死亡/提出日で1回呼ぶ", async () => {
    mockMutateAsync.mockResolvedValueOnce(undefined);
    let latestFormData: PetFormData = {
      ...basePet,
      id: "pet-synth-1",
      petName: "合成ペット甲",
      status: "生存",
      deceasedAt: null,
    };
    const setFormData = vi.fn((updater: (prev: PetFormData) => PetFormData) => {
      latestFormData = updater(latestFormData);
    });
    const onPetLifecycleChange = vi.fn();

    renderPetCareSection(latestFormData, setFormData, onPetLifecycleChange);

    fireEvent.click(screen.getByRole("button", { name: "死亡を記録" }));
    fireEvent.click(screen.getByRole("button", { name: "死亡を記録する" }));

    await waitFor(() => expect(setFormData).toHaveBeenCalled());
    await waitFor(() => expect(onPetLifecycleChange).toHaveBeenCalledTimes(1));
    expect(onPetLifecycleChange).toHaveBeenCalledWith({
      petId: "pet-synth-1",
      status: "死亡",
      deceasedAt: expect.any(String),
      deceasedReason: null,
    });
    const [{ deceasedAt }] = onPetLifecycleChange.mock.calls[0] as [
      { petId: string; status: string; deceasedAt: string; deceasedReason: string | null },
    ];
    expect(latestFormData.status).toBe("死亡");
    expect(latestFormData.deceasedAt).toBe(deceasedAt);
  });

  it("BUG-002 cross-pet: 死亡記録の完了前に別ペットへ切り替えても表示中フォームを上書きしない", async () => {
    let resolveRecord: (() => void) | undefined;
    mockMutateAsync.mockImplementationOnce(
      () => new Promise<void>((resolve) => {
        resolveRecord = resolve;
      }),
    );
    let latestFormData: PetFormData = {
      ...basePet,
      id: "pet-synth-1",
      petName: "合成ペット甲",
      status: "生存",
      deceasedAt: null,
    };
    const setFormData = vi.fn((updater: (prev: PetFormData) => PetFormData) => {
      latestFormData = updater(latestFormData);
    });
    const onPetLifecycleChange = vi.fn();
    const view = renderPetCareSection(
      latestFormData,
      setFormData,
      onPetLifecycleChange,
    );

    fireEvent.click(screen.getByRole("button", { name: "死亡を記録" }));
    fireEvent.click(screen.getByRole("button", { name: "死亡を記録する" }));
    await waitFor(() => expect(mockMutateAsync).toHaveBeenCalledTimes(1));

    const switchedPet: PetFormData = {
      ...basePet,
      id: "pet-synth-2",
      petName: "合成ペット乙",
      status: "生存",
      deceasedAt: null,
    };
    latestFormData = switchedPet;
    view.rerender(
      <MemoryRouter>
        <PetCareSection
          formData={switchedPet}
          setFormData={setFormData}
          insuranceSelectItems={null}
          isLoadingInsurances={false}
          canEdit
          onInsuranceChange={() => {}}
          onPetLifecycleChange={onPetLifecycleChange}
        />
      </MemoryRouter>,
    );

    expect(resolveRecord).toBeDefined();
    await act(async () => resolveRecord?.());

    await waitFor(() => expect(onPetLifecycleChange).toHaveBeenCalledTimes(1));
    expect(latestFormData).toBe(switchedPet);
    expect(onPetLifecycleChange).toHaveBeenCalledWith({
      petId: "pet-synth-1",
      status: "死亡",
      deceasedAt: expect.any(String),
      deceasedReason: null,
    });
  });

  it("BUG-002 死亡記録失敗時は onPetLifecycleChange を呼ばない", async () => {
    mockMutateAsync.mockRejectedValueOnce(new Error("network error"));
    const setFormData = vi.fn();
    const onPetLifecycleChange = vi.fn();

    renderPetCareSection(
      {
        ...basePet,
        id: "pet-synth-1",
        petName: "合成ペット甲",
        status: "生存",
        deceasedAt: null,
      },
      setFormData,
      onPetLifecycleChange,
    );

    fireEvent.click(screen.getByRole("button", { name: "死亡を記録" }));
    fireEvent.click(screen.getByRole("button", { name: "死亡を記録する" }));

    await screen.findByText("死亡の記録に失敗しました");
    expect(onPetLifecycleChange).not.toHaveBeenCalled();
    expect(setFormData).not.toHaveBeenCalled();
  });

  it("BUG-002 死亡記録解除成功時に onPetLifecycleChange を 生存+null で1回呼ぶ", async () => {
    mockMutate.mockImplementation((_petId, options) => {
      options?.onSuccess?.();
      options?.onSettled?.();
    });
    let latestFormData: PetFormData = {
      ...basePet,
      id: "pet-synth-1",
      petName: "合成ペット甲",
      status: "死亡",
      deceasedAt: "2026-07-10T12:00:00+09:00",
    };
    const setFormData = vi.fn((updater: (prev: PetFormData) => PetFormData) => {
      latestFormData = updater(latestFormData);
    });
    const onPetLifecycleChange = vi.fn();

    renderPetCareSection(latestFormData, setFormData, onPetLifecycleChange);

    fireEvent.click(screen.getByRole("button", { name: "死亡記録を解除" }));
    fireEvent.click(screen.getByRole("button", { name: "解除する" }));

    await waitFor(() => expect(setFormData).toHaveBeenCalled());
    await waitFor(() => expect(onPetLifecycleChange).toHaveBeenCalledTimes(1));
    expect(onPetLifecycleChange).toHaveBeenCalledWith({
      petId: "pet-synth-1",
      status: "生存",
      deceasedAt: null,
      deceasedReason: null,
    });
    expect(latestFormData.status).toBe("生存");
    expect(latestFormData.deceasedAt).toBeNull();
  });

  it("BUG-002 cross-pet: 死亡解除の完了前に別ペットへ切り替えても表示中フォームを上書きしない", async () => {
    let revokeOnSuccess: (() => void) | undefined;
    mockMutate.mockImplementation((_petId, options) => {
      revokeOnSuccess = options?.onSuccess;
    });
    let latestFormData: PetFormData = {
      ...basePet,
      id: "pet-synth-1",
      petName: "合成ペット甲",
      status: "死亡",
      deceasedAt: "2026-07-10T12:00:00+09:00",
    };
    const setFormData = vi.fn((updater: (prev: PetFormData) => PetFormData) => {
      latestFormData = updater(latestFormData);
    });
    const onPetLifecycleChange = vi.fn();
    const view = renderPetCareSection(
      latestFormData,
      setFormData,
      onPetLifecycleChange,
    );

    fireEvent.click(screen.getByRole("button", { name: "死亡記録を解除" }));
    fireEvent.click(screen.getByRole("button", { name: "解除する" }));
    await waitFor(() => expect(mockMutate).toHaveBeenCalledTimes(1));

    const switchedPet: PetFormData = {
      ...basePet,
      id: "pet-synth-2",
      petName: "合成ペット乙",
      status: "死亡",
      deceasedAt: "2026-07-11T12:00:00+09:00",
    };
    latestFormData = switchedPet;
    view.rerender(
      <MemoryRouter>
        <PetCareSection
          formData={switchedPet}
          setFormData={setFormData}
          insuranceSelectItems={null}
          isLoadingInsurances={false}
          canEdit
          onInsuranceChange={() => {}}
          onPetLifecycleChange={onPetLifecycleChange}
        />
      </MemoryRouter>,
    );

    expect(revokeOnSuccess).toBeDefined();
    revokeOnSuccess?.();

    await waitFor(() => expect(onPetLifecycleChange).toHaveBeenCalledTimes(1));
    expect(latestFormData).toBe(switchedPet);
    expect(onPetLifecycleChange).toHaveBeenCalledWith({
      petId: "pet-synth-1",
      status: "生存",
      deceasedAt: null,
      deceasedReason: null,
    });
  });
});

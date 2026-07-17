import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { PetCareSection } from "./PetCareSection";
import type { PetFormData } from "../types";

vi.mock("@/hooks/use-revoke-pet-death", () => ({
  useRevokePetDeath: () => ({ mutate: vi.fn(), isPending: false }),
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

function renderPetCareSection(formData: PetFormData) {
  return render(
    <MemoryRouter>
      <PetCareSection
        formData={formData}
        setFormData={() => {}}
        insuranceSelectItems={null}
        isLoadingInsurances={false}
        canEdit
        onInsuranceChange={() => {}}
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

  it("死亡ステータスだが deceasedAt が未取得のとき、誤って生存扱いの記録ボタンにフォールバックする", () => {
    renderPetCareSection({ ...basePet, status: "死亡", deceasedAt: undefined });

    expect(screen.getByRole("button", { name: "死亡を記録" })).toBeInTheDocument();
    expect(screen.queryByText(/永眠/)).not.toBeInTheDocument();
  });
});

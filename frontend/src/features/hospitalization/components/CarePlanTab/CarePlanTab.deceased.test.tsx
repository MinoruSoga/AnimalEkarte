import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { CarePlanTab } from "./CarePlanTab";
import type { CarePlanItem } from "../../api/care-plan-items";

const mocks = vi.hoisted(() => ({ canCreate: true, canEdit: true, canDelete: true }));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: mocks.canCreate,
    canEdit: mocks.canEdit,
    canDelete: mocks.canDelete,
  }),
}));

const item: CarePlanItem = {
  id: "item-1",
  hospitalization_id: "hosp-1",
  type: "instruction",
  name: "既存項目",
  description: "",
  timing: ["morning"],
  status: "active",
  notes: "",
  unit_price: 0,
  category: "",
  manual: false,
  other_reason: "",
  sort_order: 0,
  created_at: "",
  updated_at: "",
};

vi.mock("../../api/care-plan-items", () => ({
  useGetCarePlanItems: () => ({ data: [item], isLoading: false }),
  useCreateCarePlanItem: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateCarePlanItem: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteCarePlanItem: () => ({ mutate: vi.fn(), isPending: false }),
}));

function renderTab(petIsDeceased: boolean, isDischarged = false) {
  return render(
    <CarePlanTab
      hospitalizationId="hosp-1"
      petIsDeceased={petIsDeceased}
      isDischarged={isDischarged}
    />,
  );
}

describe("FE-RC-002: CarePlanTab — 死亡ペットの render 側防壁", () => {
  beforeEach(() => {
    mocks.canCreate = true;
    mocks.canEdit = true;
    mocks.canDelete = true;
  });

  it("petIsDeceased=false では追加・編集・削除操作を表示する", () => {
    renderTab(false);
    expect(screen.getByRole("button", { name: "編集" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "削除" })).toBeInTheDocument();
    expect(screen.getByText("新しいケアプラン項目を追加")).toBeInTheDocument();
  });

  it("petIsDeceased=true では追加・編集・削除操作を一切表示せず理由を表示する", () => {
    renderTab(true);
    expect(screen.queryByRole("button", { name: "編集" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
    expect(screen.queryByText("新しいケアプラン項目を追加")).not.toBeInTheDocument();
    expect(
      screen.getByText("死亡したペットのため、ケアプランの追加・編集・削除はできません"),
    ).toBeInTheDocument();
  });

  // EMR-227: 退院済み入院への書込みは死亡ペットと同じ二重防壁（render 側 + callback 側）で閉じる。
  it("isDischarged=true では追加・編集・削除操作を一切表示せず理由を表示する", () => {
    renderTab(false, true);
    expect(screen.queryByRole("button", { name: "編集" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
    expect(screen.queryByText("新しいケアプラン項目を追加")).not.toBeInTheDocument();
    expect(
      screen.getByText("退院済みの入院のため、ケアプランの追加・編集・削除はできません"),
    ).toBeInTheDocument();
  });

  it("petIsDeceased=true かつ isDischarged=true では死亡センチネルを優先する", () => {
    renderTab(true, true);
    expect(
      screen.getByText("死亡したペットのため、ケアプランの追加・編集・削除はできません"),
    ).toBeInTheDocument();
    expect(
      screen.queryByText("退院済みの入院のため、ケアプランの追加・編集・削除はできません"),
    ).not.toBeInTheDocument();
  });
});

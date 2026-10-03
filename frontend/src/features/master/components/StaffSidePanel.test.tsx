import type { ReactNode } from "react";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { StaffSidePanel } from "./StaffSidePanel";
import type { Staff } from "../api/staffs";
import type { ReservationType } from "../api/reservation-types";

// ──────────────────────────────────────────────────────────
// MasterSidePanel は薄くモックし、保存ハンドラだけを捕捉する。
// StaffBasicInfoSection が使う PropertyRow / StatusToggleButton もここで供給する。
// ──────────────────────────────────────────────────────────
let capturedPanelProps: { onSave?: () => Promise<void> | void } | null = null;
vi.mock("@/components/shared/SidePeek", () => ({
  MasterSidePanel: (props: { children: ReactNode; onSave?: () => Promise<void> | void }) => {
    capturedPanelProps = props;
    return <div>{props.children}</div>;
  },
  StatusToggleButton: () => null,
  PropertyRow: ({ label, children }: { label?: ReactNode; children: ReactNode }) => (
    <div>
      {label}
      {children}
    </div>
  ),
}));

// server の capableIds。テスト毎に差し替える。
const capableQuery = vi.hoisted(() => ({ data: ["1"] as string[] | undefined }));
vi.mock("../api/staffs", () => ({
  useGetStaffCapableReservationTypes: () => ({ data: capableQuery.data }),
  useGetStaffClinics: () => ({ data: [] }),
  useGetStaffPermissionGroups: () => ({ data: [] }),
  useAttachStaffAccount: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { isSystemAdmin: false } }),
}));

function makeStaff(overrides: Partial<Staff> = {}): Staff {
  return {
    id: "1",
    clinicId: "1",
    name: "既存 太郎",
    isActive: true,
    occupationId: null,
    occupationName: null,
    licenseNumber: "",
    sortOrder: 1,
    email: "existing@example.com",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    staffType: "doctor",
    reservationDisplayName: "",
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    ...overrides,
  };
}

function reservationType(overrides: Partial<ReservationType>): ReservationType {
  return {
    id: "1",
    clinicId: "1",
    name: "一般診療",
    color: "#111111",
    isActive: true,
    description: "",
    sortOrder: 1,
    groupId: undefined,
    createdAt: "2026-05-29T00:00:00Z",
    updatedAt: "2026-05-29T00:00:00Z",
    reservationDisplayName: "",
    durationMinutes: 30,
    shortName: "",
    showShortName: false,
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    reservationDayOption: "none",
    isInternal: false,
    category: "general",
    ...overrides,
  };
}

const TYPES = [
  reservationType({ id: "1", name: "一般診療", category: "general" }),
  reservationType({ id: "2", name: "シャンプー", category: "trimming" }),
];

function renderPanel({
  item = makeStaff(),
  onSave = vi.fn(async () => true),
  allReservationTypes = TYPES,
  onDirtyChange,
}: {
  item?: Staff | null;
  onSave?: () => Promise<boolean>;
  allReservationTypes?: ReservationType[];
  onDirtyChange?: (dirty: boolean) => void;
} = {}) {
  const onSaveGroups = vi.fn();
  const onSaveClinics = vi.fn();
  const onSaveCapableReservationTypes = vi.fn();
  render(
    <MemoryRouter>
      <StaffSidePanel
        item={item}
        onClose={() => {}}
        onSave={onSave}
        allOccupations={[]}
        allGroups={[]}
        onSaveGroups={onSaveGroups}
        allClinics={[]}
        onSaveClinics={onSaveClinics}
        allReservationTypes={allReservationTypes}
        onSaveCapableReservationTypes={onSaveCapableReservationTypes}
        onDirtyChange={onDirtyChange}
      />
    </MemoryRouter>,
  );
  return { onSaveGroups, onSaveClinics, onSaveCapableReservationTypes, onSave };
}

describe("StaffSidePanel 対応区分（capableIds 単一書き込みモデル）", () => {
  beforeEach(() => {
    capturedPanelProps = null;
    capableQuery.data = ["1"];
  });

  it("対応可能/対応不可 両 view での編集は1つの capableIds へ集約され、保存で1回だけ送られる", async () => {
    const { onSaveCapableReservationTypes } = renderPanel();
    const user = userEvent.setup({ delay: null });

    // 対応可能 view で シャンプー を対応可能に追加 → capableIds = ["1", "2"]
    await user.click(screen.getByRole("checkbox", { name: /シャンプー/ }));

    // 対応不可 view で 一般診療 にチェック（= 対応できない） → capableIds = ["2"]
    await user.click(screen.getByRole("radio", { name: "対応不可" }));
    await user.click(screen.getByRole("checkbox", { name: /一般診療/ }));

    await act(async () => {
      await capturedPanelProps?.onSave?.();
    });

    expect(onSaveCapableReservationTypes).toHaveBeenCalledTimes(1);
    expect(onSaveCapableReservationTypes).toHaveBeenCalledWith("1", ["2"]);
  });

  it("view 切替だけでは dirty にならない（表示のみの操作で書き込み状態を作らない）", async () => {
    const onDirtyChange = vi.fn();
    renderPanel({ onDirtyChange });
    const user = userEvent.setup({ delay: null });

    await user.click(screen.getByRole("radio", { name: "対応不可" }));
    await user.click(screen.getByRole("radio", { name: "対応可能" }));

    expect(onDirtyChange).not.toHaveBeenCalledWith(true);
  });

  it("onSave が失敗（false）を返した場合は capableIds の保存を発行しない", async () => {
    const { onSaveCapableReservationTypes } = renderPanel({
      onSave: vi.fn(async () => false),
    });
    const user = userEvent.setup({ delay: null });

    await user.click(screen.getByRole("checkbox", { name: /シャンプー/ }));
    await act(async () => {
      await capturedPanelProps?.onSave?.();
    });

    expect(onSaveCapableReservationTypes).not.toHaveBeenCalled();
  });

  it("予約区分一覧が未取得でも、保存は逆算せずサーバー取得済みの capableIds をそのまま送る", async () => {
    const { onSaveCapableReservationTypes } = renderPanel({ allReservationTypes: [] });

    expect(screen.getByText("予約区分が登録されていません")).toBeInTheDocument();

    await act(async () => {
      await capturedPanelProps?.onSave?.();
    });

    // 全有効区分が未取得の状態で「全件 − 除外選択」の逆算保存をしていれば [] になる。
    // 単一モデル方式では編集経路が存在しないためサーバー値が素通りする。
    expect(onSaveCapableReservationTypes).toHaveBeenCalledTimes(1);
    expect(onSaveCapableReservationTypes).toHaveBeenCalledWith("1", ["1"]);
  });

  it("新規スタッフでは「スタッフ登録後に設定できます」を表示し、保存しても capableIds は送らない", async () => {
    const { onSaveCapableReservationTypes } = renderPanel({ item: null });

    expect(screen.getByText("スタッフ登録後に設定できます")).toBeInTheDocument();

    await act(async () => {
      await capturedPanelProps?.onSave?.();
    });

    expect(onSaveCapableReservationTypes).not.toHaveBeenCalled();
  });
});

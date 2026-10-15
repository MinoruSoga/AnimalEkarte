import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { formatJSTWallDate, formatJSTWallTime } from "@/lib/jst-date";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { usePermission } from "@/hooks/use-permission";
import type { Reservation } from "@/lib/transforms/reservation";

import { NextVisitButton } from "./NextVisitButton";

const setMutateAsync = vi.hoisted(() => vi.fn());
const clearMutateAsync = vi.hoisted(() => vi.fn());
const confirmMutateAsync = vi.hoisted(() => vi.fn());
const modalPropsHolder = vi.hoisted(() => ({
  current: null as Record<string, unknown> | null,
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

vi.mock("@/hooks/use-get-reservations", () => ({
  useGetReservations: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(),
}));

// API 層（axios + invalidateQueries）はモックし、注入した AsyncMutation が
// フック経由で呼ばれることを検証する。
vi.mock("../api/next-visit-intent", () => ({
  useSetNextVisitIntent: () => ({ mutateAsync: setMutateAsync }),
  useClearNextVisitIntent: () => ({ mutateAsync: clearMutateAsync }),
}));

// EMR-308 A1: confirm-tentative mutation は共有 hooks。QueryClient 不要にするためモック化。
vi.mock("@/hooks/use-confirm-tentative", () => ({
  useConfirmTentative: () => ({ mutateAsync: confirmMutateAsync }),
}));

// ReservationFormModal は lazy import。同一 specifier を vi.mock で差し替え、
// tentative 確定モードの props（tentativeConfirmId / onConfirmTentative / initialData）を捕捉する。
vi.mock("@/components/shared/ReservationFormModal/ReservationFormModal", () => ({
  ReservationFormModal: (props: Record<string, unknown>) => {
    modalPropsHolder.current = props;
    return props.isOpen ? <div data-testid="confirm-modal" /> : null;
  },
}));

function wallDateInDays(days: number, hour = 10, minute = 0): Date {
  const d = new Date();
  d.setDate(d.getDate() + days);
  d.setHours(hour, minute, 0, 0);
  return d;
}

/** 表示対象の status/start/id と clear の更新revisionを持つ最小予約。 */
function reservationOf(status: string, start: Date, id = "r-1"): Reservation {
  return {
    id,
    status,
    start,
    updatedAt: "2026-10-11T00:00:00.123456789Z",
  } as unknown as Reservation;
}

function mockReservations(items: Reservation[]) {
  vi.mocked(useGetReservations).mockReturnValue({
    data: items,
  } as unknown as ReturnType<typeof useGetReservations>);
}

function mockPermission(canCreate: boolean, canEdit = true) {
  vi.mocked(usePermission).mockReturnValue({
    canView: true,
    canCreate,
    canEdit,
    canDelete: false,
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  setMutateAsync.mockResolvedValue({});
  clearMutateAsync.mockResolvedValue(undefined);
  confirmMutateAsync.mockResolvedValue({});
  modalPropsHolder.current = null;
  mockReservations([]);
  mockPermission(true);
});

describe("NextVisitButton — 表示優先度（EMR-308 A1）", () => {
  it("予約が無いときはアイコンのみ（次回予定を追加）", () => {
    render(<NextVisitButton petId="10" medicalRecordId="99" />);

    expect(screen.getByRole("button", { name: "次回予定を追加" })).toBeEnabled();
    expect(screen.getByRole("tooltip", { hidden: true })).toHaveTextContent("次回予定を追加");
  });

  it("未来の pending/confirmed 実予約は日時（日付+時刻）で表示する", () => {
    const start = wallDateInDays(30, 14, 30);
    mockReservations([reservationOf("confirmed", start)]);
    render(<NextVisitButton petId="10" medicalRecordId="99" />);

    const label = `${formatJSTWallDate(start)} ${formatJSTWallTime(start)}`;
    expect(screen.getByText(label)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "次回予定を変更" })).toBeEnabled();
  });

  it("tentative のみなら意図日を日付のみで表示する", async () => {
    const start = wallDateInDays(45, 0, 0);
    mockReservations([reservationOf("tentative", start)]);
    // tentative 行があると lazy ReservationFormModal がマウントされるため、
    // import 解決を act で flush する。
    await act(async () => {
      render(<NextVisitButton petId="10" medicalRecordId="99" />);
    });

    expect(screen.getByText(formatJSTWallDate(start))).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "次回予定を変更" })).toBeEnabled();
  });
});

describe("NextVisitButton — Popover での tentative 操作（EMR-308 A2）", () => {
  it("クイック選択（3ヶ月後）は PUT 経路に intended_date + source_medical_record_id で送る", async () => {
    const user = userEvent.setup();
    render(<NextVisitButton petId="10" medicalRecordId="99" />);

    await user.click(screen.getByRole("button", { name: "次回予定を追加" }));
    await user.click(screen.getByRole("button", { name: "3ヶ月後" }));

    expect(setMutateAsync).toHaveBeenCalledWith({
      petId: "10",
      intendedDate: expect.stringMatching(/^\d{4}-\d{2}-\d{2}$/),
      medicalRecordId: "99",
    });
    expect(clearMutateAsync).not.toHaveBeenCalled();
  });

  it("手入力の日付も PUT 経路に送る", async () => {
    const user = userEvent.setup();
    render(<NextVisitButton petId="10" medicalRecordId="99" />);

    await user.click(screen.getByRole("button", { name: "次回予定を追加" }));
    const input = screen.getByLabelText("次回来院推奨日");
    const d = new Date();
    d.setMonth(d.getMonth() + 2);
    const date = formatJSTWallDate(d);
    await user.clear(input);
    await user.type(input, date.replaceAll("-", "/"));
    await user.tab();

    // Codex PR606 #11: mutation は FIFO chain 上の microtask で発行されるため
    // 同期 expect ではなく waitFor で待つ
    await waitFor(() =>
      expect(setMutateAsync).toHaveBeenCalledWith({
        petId: "10",
        intendedDate: date,
        medicalRecordId: "99",
      }),
    );
  });

  it("クリアは DELETE 経路を呼び PUT は呼ばない", async () => {
    const user = userEvent.setup();
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0))]);
    await act(async () => {
      render(<NextVisitButton petId="10" medicalRecordId="99" />);
    });

    await user.click(screen.getByRole("button", { name: "次回予定を変更" }));
    await user.click(screen.getByRole("button", { name: "クリア" }));

    expect(clearMutateAsync).toHaveBeenCalledWith({
      petId: "10",
      expectedReservationId: "r-1",
      expectedUpdatedAt: "2026-10-11T00:00:00.123456789Z",
    });
    expect(setMutateAsync).not.toHaveBeenCalled();
  });

  it("1ヶ月後/3ヶ月後/6ヶ月後/1年後/手入力/クリアの選択肢を Popover に保持する", async () => {
    const user = userEvent.setup();
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0))]);
    await act(async () => {
      render(<NextVisitButton petId="10" medicalRecordId="99" />);
    });

    await user.click(screen.getByRole("button", { name: "次回予定を変更" }));

    expect(screen.getByRole("button", { name: "1ヶ月後" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "3ヶ月後" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "6ヶ月後" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "1年後" })).toBeEnabled();
    expect(screen.getByLabelText("次回来院推奨日")).toBeEnabled();
    expect(screen.getByRole("button", { name: "クリア" })).toBeEnabled();
  });
});

describe("NextVisitButton — ガード", () => {
  it("reservations:create/edit ともに無いとボタンは無効で mutation も発行しない", async () => {
    mockPermission(false, false);
    const user = userEvent.setup();
    render(<NextVisitButton petId="10" medicalRecordId="99" />);

    const trigger = screen.getByRole("button", { name: "次回予定を追加" });
    expect(trigger).toBeDisabled();
    await user.click(trigger);

    expect(setMutateAsync).not.toHaveBeenCalled();
    expect(clearMutateAsync).not.toHaveBeenCalled();
  });

  it("死亡ペットではボタンは無効", () => {
    render(<NextVisitButton petId="10" medicalRecordId="99" isPetDeceased={true} />);

    expect(screen.getByRole("button", { name: "次回予定を追加" })).toBeDisabled();
  });
});

describe("NextVisitButton — 権限分割（PUT=create / DELETE=edit）", () => {
  it("create のみ（edit 無し）: 既存 tentative の変更・クリアは不可（trigger 自体が無効）", async () => {
    // BE は既存 intent の PUT upsert と DELETE に edit を要求するため、
    // create-only ユーザーに actionable な操作が残らない → trigger を disabled にする。
    mockPermission(true, false);
    const user = userEvent.setup();
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0))]);
    await act(async () => {
      render(<NextVisitButton petId="10" medicalRecordId="99" />);
    });

    const trigger = screen.getByRole("button", { name: "次回予定を変更" });
    expect(trigger).toBeDisabled();
    await user.click(trigger);

    expect(setMutateAsync).not.toHaveBeenCalled();
    expect(clearMutateAsync).not.toHaveBeenCalled();
  });

  it("edit のみ（create 無し）: クリアは DELETE を発行するが日付設定は PUT を発行しない", async () => {
    mockPermission(false, true);
    const user = userEvent.setup();
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0))]);
    await act(async () => {
      render(<NextVisitButton petId="10" medicalRecordId="99" />);
    });

    // クリア（DELETE = reservations:edit）は許可される
    await user.click(screen.getByRole("button", { name: "次回予定を変更" }));
    await user.click(screen.getByRole("button", { name: "クリア" }));
    expect(clearMutateAsync).toHaveBeenCalledWith({
      petId: "10",
      expectedReservationId: "r-1",
      expectedUpdatedAt: "2026-10-11T00:00:00.123456789Z",
    });

    // 日付設定（PUT = reservations:create）は発行しない
    const input = screen.getByLabelText("次回来院推奨日");
    const d = new Date();
    d.setMonth(d.getMonth() + 2);
    fireEvent.change(input, { target: { value: formatJSTWallDate(d) } });
    expect(setMutateAsync).not.toHaveBeenCalled();
  });
});

describe("NextVisitButton — 予約を確定（EMR-308 A1）", () => {
  it("tentative 表示中は「予約を確定」ボタンを出す", async () => {
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0), "t-1")]);
    await act(async () => {
      render(<NextVisitButton petId="10" medicalRecordId="99" />);
    });

    expect(screen.getByRole("button", { name: "次回予定を確定" })).toBeEnabled();
  });

  it("実予約がある場合は確定導線を出さない（二重予約防止）", () => {
    mockReservations([reservationOf("confirmed", wallDateInDays(30))]);
    render(<NextVisitButton petId="10" medicalRecordId="99" />);

    expect(screen.queryByRole("button", { name: "次回予定を確定" })).not.toBeInTheDocument();
  });

  it("tentative が無い（未設定）ときは出さない", () => {
    render(<NextVisitButton petId="10" medicalRecordId="99" />);

    expect(screen.queryByRole("button", { name: "次回予定を確定" })).not.toBeInTheDocument();
  });

  it("reservations:edit が無い場合は出さない（fail-closed）", async () => {
    mockPermission(true, false);
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0), "t-1")]);
    await act(async () => {
      render(<NextVisitButton petId="10" medicalRecordId="99" />);
    });

    expect(screen.queryByRole("button", { name: "次回予定を確定" })).not.toBeInTheDocument();
  });

  it("クリックで確定モーダルが tentativeConfirmId 付きで開き、onConfirmTentative が POST body 用データを渡す", async () => {
    const user = userEvent.setup();
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0), "t-1")]);
    await act(async () => {
      render(<NextVisitButton petId="10" medicalRecordId="99" />);
    });

    await user.click(screen.getByRole("button", { name: "次回予定を確定" }));

    // lazy モーダル（mock）が tentative 確定モードで開く
    await waitFor(() => expect(screen.getByTestId("confirm-modal")).toBeInTheDocument());
    expect(modalPropsHolder.current?.tentativeConfirmId).toBe("t-1");
    expect(modalPropsHolder.current?.initialData).toMatchObject({ petId: "10" });
    // 別行作成ではなく同一行昇格のため onSave ではなく onConfirmTentative を使う
    expect(modalPropsHolder.current?.onSave).toBeUndefined();

    const onConfirmTentative = modalPropsHolder.current?.onConfirmTentative as (
      id: string,
      data: Record<string, unknown>,
    ) => Promise<string | null>;
    const formData = {
      start: wallDateInDays(45, 10, 0),
      end: wallDateInDays(45, 11, 0),
      status: "confirmed",
    };
    // 成功時に setConfirmOpen(false) が走るため act で包む
    let confirmResult: string | null | undefined;
    await act(async () => {
      confirmResult = await onConfirmTentative("t-1", formData);
    });
    expect(confirmResult).toBeNull();
    expect(confirmMutateAsync).toHaveBeenCalledWith({
      reservationId: "t-1",
      formData,
    });
    // 成功でモーダルを閉じる
    await waitFor(() => expect(screen.queryByTestId("confirm-modal")).not.toBeInTheDocument());
  });

  it("確定失敗時はインラインエラー文字列を返してモーダルを開いたままにする", async () => {
    const user = userEvent.setup();
    confirmMutateAsync.mockRejectedValue(new Error("api failed"));
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0), "t-1")]);
    await act(async () => {
      render(<NextVisitButton petId="10" medicalRecordId="99" />);
    });

    await user.click(screen.getByRole("button", { name: "次回予定を確定" }));
    await waitFor(() => expect(screen.getByTestId("confirm-modal")).toBeInTheDocument());

    const onConfirmTentative = modalPropsHolder.current?.onConfirmTentative as (
      id: string,
      data: Record<string, unknown>,
    ) => Promise<string | null>;
    let result: string | null | undefined;
    await act(async () => {
      result = await onConfirmTentative("t-1", {
        start: wallDateInDays(45, 10, 0),
        end: wallDateInDays(45, 11, 0),
      });
    });

    expect(result).toBe("予約の確定に失敗しました。入力内容を確認して再度お試しください。");
    expect(screen.getByTestId("confirm-modal")).toBeInTheDocument();
  });
});

import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Pet } from "@/types";
import { DangerLevelHigh, PetStatusAlive, PetStatusDeceased } from "@/types/generated/models";

import { useReceptionModalHandlers } from "./use-reception-modal-handlers";
import type { ReceptionAppointment } from "../api/types";

const { toastSuccessMock, updateReservationMock } = vi.hoisted(() => ({
  toastSuccessMock: vi.fn(),
  updateReservationMock: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: toastSuccessMock },
}));

vi.mock("@/hooks/use-update-reservation", () => ({
  useUpdateReservation: () => ({
    mutate: updateReservationMock,
  }),
}));

const baseAppointment: ReceptionAppointment = {
  id: "101",
  time: "09:45",
  visitDate: "2026-06-01",
  end: new Date(2026, 5, 1, 10, 5, 0),
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "一般診察",
  reservationTypeId: "1",
  reservationCategory: "general",
  isDesignated: false,
  doctor: "担当者A",
  doctorId: "33",
  petId: "10",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
};

const baseSelectedPet = {
  id: "11",
  clinicId: "1",
  ownerId: "21",
  ownerNumber: 21,
  ownerName: "佐藤",
  ownerNameKana: undefined,
  address: undefined,
  phone: "",
  petNumber: "P-11",
  name: "ミケ",
  petNameKana: undefined,
  species: "猫",
  animalSpeciesId: "2",
  breed: "",
  color: "",
  bloodType: undefined,
  microchipNumber: undefined,
  gender: "雌",
  status: "生存",
  birthDate: undefined,
  neuteredDate: undefined,
  weight: undefined,
  food: "",
  environment: "",
  acquisitionType: undefined,
  dangerLevel: "低",
  dangerReason: undefined,
  lastVisit: undefined,
  insuranceId: undefined,
  insuranceName: undefined,
  insuranceDetails: undefined,
  remarks: "",
  deceasedAt: undefined,
} satisfies Pet;

interface HandlerProps {
  canEditReservation?: boolean;
  canDeleteReservation?: boolean;
}

function renderHandlers(
  updateAppointment = vi.fn(),
  initialPermissions: HandlerProps = {
    canEditReservation: true,
    canDeleteReservation: true,
  },
) {
  const advanceStatus = vi.fn();
  const cancelAppointment = vi.fn();
  const view = renderHook(
    (permissions: HandlerProps) =>
      useReceptionModalHandlers({
        advanceStatus,
        cancelAppointment,
        updateAppointment,
        ...permissions,
      }),
    { initialProps: initialPermissions },
  );
  return { ...view, advanceStatus, cancelAppointment };
}

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

describe("useReceptionModalHandlers", () => {
  beforeEach(() => {
    toastSuccessMock.mockReset();
    updateReservationMock.mockReset();
    updateReservationMock.mockImplementation((_payload, options?: { onSuccess?: () => void }) => {
      options?.onSuccess?.();
    });
  });

  it("受付予約編集では appointment の visitDate と time から開始日時を作る", () => {
    const { result } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });

    expect(result.current.editingAppointment?.start?.getFullYear()).toBe(2026);
    expect(result.current.editingAppointment?.start?.getMonth()).toBe(5);
    expect(result.current.editingAppointment?.start?.getDate()).toBe(1);
    expect(result.current.editingAppointment?.start?.getHours()).toBe(9);
    expect(result.current.editingAppointment?.start?.getMinutes()).toBe(45);
    expect(result.current.editingAppointment?.end?.getHours()).toBe(10);
    expect(result.current.editingAppointment?.end?.getMinutes()).toBe(5);
    expect(result.current.editingAppointment?.type).toBe("1");
    expect(result.current.editingAppointment?.doctor).toBe("33");
    expect(result.current.editingAppointment?.status).toBe("checked_in");
  });

  it("編集保存では予約更新 API を呼び、成功後にローカル appointment も更新する", () => {
    const updateAppointment = vi.fn();
    const { result } = renderHandlers(updateAppointment);

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [],
      );
    });

    expect(updateReservationMock).toHaveBeenCalledWith(
      {
        id: "101",
        req: expect.objectContaining({
          pet_id: 10,
          owner_id: 20,
          status: "checked_in",
        }),
      },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
    const req = (updateReservationMock.mock.calls[0]?.[0] as { req: Record<string, unknown> }).req;
    expect(req).not.toHaveProperty("start_time");
    expect(req).not.toHaveProperty("end_time");
    expect(updateAppointment).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "101",
        time: "09:45",
        visitDate: "2026-06-01",
      }),
    );
  });

  it("新しく選択した生存・高危険ペットの sentinel を楽観更新へ渡す", () => {
    const updateAppointment = vi.fn();
    const { result } = renderHandlers(updateAppointment);

    act(() => {
      result.current.handleEditAppointment({
        ...baseAppointment,
        petDangerLevel: "low",
        petDangerReason: "旧ペットの理由",
      });
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [
          {
            ...baseSelectedPet,
            dangerLevel: "高",
            dangerReason: "保定時に噛む",
          },
        ],
      );
    });

    expect(updateAppointment).toHaveBeenCalledWith(
      expect.objectContaining({
        petId: "11",
        petStatus: PetStatusAlive,
        petDangerLevel: DangerLevelHigh,
        petDangerReason: "保定時に噛む",
      }),
    );
  });

  it("新しく選択したペットの危険度が不明なら API 成功後も server refetch を待つ", () => {
    const updateAppointment = vi.fn();
    const { result } = renderHandlers(updateAppointment);

    act(() => {
      result.current.handleEditAppointment({
        ...baseAppointment,
        petDangerLevel: DangerLevelHigh,
        petDangerReason: "旧ペットの理由",
      });
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [{ ...baseSelectedPet, dangerLevel: undefined }],
      );
    });

    expect(updateReservationMock).toHaveBeenCalledOnce();
    expect(updateAppointment).not.toHaveBeenCalled();
  });

  it("ペット選択を変更しない編集では既存 sentinel を楽観更新で保持する", () => {
    const updateAppointment = vi.fn();
    const { result } = renderHandlers(updateAppointment);

    act(() => {
      result.current.handleEditAppointment({
        ...baseAppointment,
        petStatus: PetStatusAlive,
        petDangerLevel: DangerLevelHigh,
        petDangerReason: "保定時に噛む",
      });
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [],
      );
    });

    expect(updateAppointment).toHaveBeenCalledWith(
      expect.objectContaining({
        petStatus: PetStatusAlive,
        petDangerLevel: DangerLevelHigh,
        petDangerReason: "保定時に噛む",
      }),
    );
  });

  it("編集保存では予約区分IDと担当者IDが不正値なら undefined にする", () => {
    const { result } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "一般診察",
          doctor: "担当者A",
        },
        [],
      );
    });

    expect(updateReservationMock).toHaveBeenCalledOnce();
    const req = (updateReservationMock.mock.calls[0]?.[0] as { req: Record<string, unknown> }).req;
    expect(req).not.toHaveProperty("reservation_type_id");
    expect(req).not.toHaveProperty("doctor_id");
  });

  it("編集対象に新しく選択されたペットが死亡の場合は予約更新を拒否する", () => {
    const { result } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [
          {
            ...baseSelectedPet,
            status: "死亡",
          },
        ],
      );
    });

    expect(updateReservationMock).not.toHaveBeenCalled();
  });

  it("受付予約カラムの編集では pending 表示を confirmed として保存する", () => {
    const { result } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment({ ...baseAppointment, status: "pending" });
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [],
      );
    });

    expect(updateReservationMock).toHaveBeenCalledWith(
      {
        id: "101",
        req: expect.objectContaining({
          status: "confirmed",
        }),
      },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
  });

  it("死亡 appointment の card click・status・edit・cancel callback を拒否する", () => {
    const deceasedAppointment = {
      ...baseAppointment,
      petStatus: PetStatusDeceased,
    };
    const { result, advanceStatus, cancelAppointment } = renderHandlers();

    act(() => {
      result.current.handleCardClick(deceasedAppointment);
      result.current.handleAdvanceStatus();
      result.current.handleEditAppointment(deceasedAppointment);
      result.current.handleCancelAppointment(deceasedAppointment);
      result.current.executeCancel();
    });

    expect(result.current.modalOpen).toBe(false);
    expect(result.current.isEditModalOpen).toBe(false);
    expect(result.current.cancelConfirmOpen).toBe(false);
    expect(advanceStatus).not.toHaveBeenCalled();
    expect(cancelAppointment).not.toHaveBeenCalled();
    expect(updateReservationMock).not.toHaveBeenCalled();
  });

  it("captured status callback は commit 後の最新 edit 権限を再確認する", () => {
    const { result, rerender, advanceStatus } = renderHandlers();

    act(() => {
      result.current.handleCardClick(baseAppointment);
    });
    const capturedAdvanceStatus = result.current.handleAdvanceStatus;

    rerender({ canEditReservation: false, canDeleteReservation: true });
    act(() => {
      capturedAdvanceStatus();
    });

    expect(advanceStatus).not.toHaveBeenCalled();
  });

  it("captured edit save callback は commit 後の最新 edit 権限を再確認する", () => {
    const { result, rerender } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });
    const capturedEditSave = result.current.handleEditSave;

    rerender({ canEditReservation: undefined, canDeleteReservation: true });
    act(() => {
      capturedEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [],
      );
    });

    expect(updateReservationMock).not.toHaveBeenCalled();
  });

  it("captured cancel callback は commit 後の最新 delete 権限を再確認する", () => {
    const { result, rerender, cancelAppointment } = renderHandlers();

    act(() => {
      result.current.handleCancelAppointment(baseAppointment);
    });
    const capturedExecuteCancel = result.current.executeCancel;

    rerender({ canEditReservation: true, canDeleteReservation: false });
    act(() => {
      capturedExecuteCancel();
    });

    expect(cancelAppointment).not.toHaveBeenCalled();
  });

  it("cancel API pending 中は成功 toast を表示しない", async () => {
    const pendingCancel = createDeferred<boolean>();
    const { result, cancelAppointment } = renderHandlers();
    cancelAppointment.mockReturnValueOnce(pendingCancel.promise);

    act(() => {
      result.current.handleCancelAppointment(baseAppointment);
    });
    act(() => {
      void result.current.executeCancel();
    });

    expect(cancelAppointment).toHaveBeenCalledWith(baseAppointment.id);
    expect(toastSuccessMock).not.toHaveBeenCalled();

    await act(async () => {
      pendingCancel.resolve(true);
      await pendingCancel.promise;
    });
  });

  it("cancel API 失敗時は成功 toast を表示せず確認画面を維持する", async () => {
    const { result, cancelAppointment } = renderHandlers();
    cancelAppointment.mockResolvedValueOnce(false);

    act(() => {
      result.current.handleCancelAppointment(baseAppointment);
    });
    await act(async () => {
      await result.current.executeCancel();
    });

    expect(toastSuccessMock).not.toHaveBeenCalled();
    expect(result.current.cancelConfirmOpen).toBe(true);
  });

  it("cancel API 成功後にだけ成功 toast を表示して確認画面を閉じる", async () => {
    const { result, cancelAppointment } = renderHandlers();
    cancelAppointment.mockResolvedValueOnce(true);

    act(() => {
      result.current.handleCancelAppointment(baseAppointment);
    });
    await act(async () => {
      await result.current.executeCancel();
    });

    expect(toastSuccessMock).toHaveBeenCalledWith("予約を取り消しました");
    expect(result.current.cancelConfirmOpen).toBe(false);
  });

  describe("受付編集の終了時刻復元 (BUG-002)", () => {
    const restoredEnd = new Date(2026, 5, 1, 10, 5, 0);
    const appointmentWithRealEnd = {
      ...baseAppointment,
      end: restoredEnd,
      notes: "元メモ",
    };

    it("handleEditAppointment は appointment.end の実終了時刻をフォームへ復元する", () => {
      const { result } = renderHandlers();

      act(() => {
        result.current.handleEditAppointment(appointmentWithRealEnd);
      });

      expect(result.current.editingAppointment?.end?.getFullYear()).toBe(2026);
      expect(result.current.editingAppointment?.end?.getMonth()).toBe(5);
      expect(result.current.editingAppointment?.end?.getDate()).toBe(1);
      expect(result.current.editingAppointment?.end?.getHours()).toBe(10);
      expect(result.current.editingAppointment?.end?.getMinutes()).toBe(5);
    });

    it("復元した終了時刻のままメモだけ変更した保存では start_time/end_time を送らない", () => {
      const { result } = renderHandlers();

      act(() => {
        result.current.handleEditAppointment(appointmentWithRealEnd);
      });
      act(() => {
        result.current.handleEditSave(
          {
            start: new Date(2026, 5, 1, 9, 45, 0),
            end: restoredEnd,
            visitType: "revisit",
            type: "1",
            doctor: "33",
            notes: "メモだけ変更",
          },
          [],
        );
      });

      expect(updateReservationMock).toHaveBeenCalledOnce();
      const payload = updateReservationMock.mock.calls[0]?.[0] as {
        id: string;
        req: Record<string, unknown>;
      };
      expect(payload.req).not.toHaveProperty("start_time");
      expect(payload.req).not.toHaveProperty("end_time");
      expect(payload.req).not.toHaveProperty("doctor_id");
      expect(payload.req).not.toHaveProperty("reservation_type_id");
      expect(payload.req.notes).toBe("メモだけ変更");
    });
  });
});

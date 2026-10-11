import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import { useGetReservations } from "@/hooks/use-get-reservations";
import { useSetNextVisitIntent, useClearNextVisitIntent } from "@/hooks/use-next-visit-intent";
import { axios } from "@/lib/axios";
import { formatJSTWallDate, toJSTWallDate } from "@/lib/jst-date";
import type { Reservation } from "@/lib/transforms/reservation";

import { useNextVisitIntent } from "./use-next-visit-intent";

vi.mock("sonner", () => ({ toast: { success: vi.fn() } }));
vi.mock("@/lib/axios", () => ({ axios: { put: vi.fn(), delete: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));
vi.mock("@/hooks/use-get-reservations", () => ({ useGetReservations: vi.fn() }));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: true, canEdit: true }),
}));

function futureDate(days: number) {
  const date = toJSTWallDate(new Date());
  date.setDate(date.getDate() + days);
  return formatJSTWallDate(date);
}

function tentative(date: string): Reservation {
  return {
    id: `intent-${date}`,
    status: "tentative",
    intendedDate: date,
    start: new Date(`${date}T00:00:00`),
  } as unknown as Reservation;
}

function deferred() {
  let resolve!: () => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<void>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}

const rows = new Map<string, Reservation[]>();
const scopeKey = (petId: string, clinicId?: string) => JSON.stringify([petId, clinicId]);

beforeEach(() => {
  vi.clearAllMocks();
  rows.clear();
  vi.mocked(useGetReservations).mockImplementation(
    (options = {}) =>
      ({
        data: rows.get(scopeKey(options.petId ?? "", options.clinicId)) ?? [],
      }) as ReturnType<typeof useGetReservations>,
  );
});

function renderScopedIntent(petId = "10", clinicId = "1") {
  const setMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
  const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
  const hook = renderHook(
    (scope: { petId: string; clinicId: string }) =>
      useNextVisitIntent({
        ...scope,
        isPetDeceased: false,
        setIntentMutation: setMutation,
        clearIntentMutation: clearMutation,
      }),
    { initialProps: { petId, clinicId } },
  );
  return { ...hook, setMutation, clearMutation };
}

describe("useNextVisitIntent — pet/clinic scope isolation", () => {
  it.each([
    { petId: "20", clinicId: "1" },
    { petId: "10", clinicId: "2" },
  ])("同じサーバー日付でもscope $petId/$clinicIdへの切替で旧draftを捨てる", (next) => {
    const date = futureDate(30);
    rows.set(scopeKey("10", "1"), [tentative(date)]);
    rows.set(scopeKey(next.petId, next.clinicId), [tentative(date)]);
    const { result, rerender, setMutation } = renderScopedIntent();
    act(() => result.current.handleIntentDateChange("2099-01-01"));
    expect(result.current.intentFieldValue).toBe("2099-01-01");
    rerender(next);
    expect(result.current.intentFieldValue).toBe(date);
    expect(setMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it.each(["success", "failure"])(
    "旧scope mutationの%sは新scopeのdraft・pending・queueに影響しない",
    async (outcome) => {
      const old = deferred();
      const current = deferred();
      const oldDate = futureDate(30);
      const serverDate = futureDate(45);
      const draftDate = futureDate(60);
      const refetchedDate = futureDate(75);
      rows.set(scopeKey("20", "2"), [tentative(serverDate)]);
      const { result, rerender, setMutation } = renderScopedIntent();
      setMutation.mutateAsync
        .mockImplementationOnce(() => old.promise)
        .mockImplementationOnce(() => current.promise);
      await act(async () => result.current.handleIntentDateChange(oldDate));
      rerender({ petId: "20", clinicId: "2" });
      expect(result.current.intentFieldValue).toBe(serverDate);
      await act(async () => result.current.handleIntentDateChange(draftDate));
      await waitFor(() => expect(setMutation.mutateAsync).toHaveBeenCalledTimes(2));
      expect(setMutation.mutateAsync).toHaveBeenLastCalledWith({
        petId: "20",
        clinicId: "2",
        intendedDate: draftDate,
        medicalRecordId: undefined,
      });
      rows.set(scopeKey("20", "2"), [tentative(refetchedDate)]);
      rerender({ petId: "20", clinicId: "2" });
      expect(result.current.intentFieldValue).toBe(draftDate);
      await act(async () => {
        if (outcome === "success") old.resolve();
        else old.reject(new Error("old request failed"));
      });
      expect(result.current.intentFieldValue).toBe(draftDate);
      expect(toast.success).not.toHaveBeenCalled();
      // pending中のrefetchは、追加の日付変更なしでBの完了時に反映する。
      await act(async () => current.resolve());
      expect(result.current.intentFieldValue).toBe(refetchedDate);
    },
  );

  it("旧callbackを切替後に呼んでも新scopeのdraftや書込を変更しない", () => {
    const { result, rerender, setMutation } = renderScopedIntent();
    const oldChange = result.current.handleIntentDateChange;
    rerender({ petId: "20", clinicId: "1" });
    act(() => oldChange(futureDate(30)));
    expect(result.current.intentFieldValue).toBe("");
    expect(setMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it.each(["success", "failure"])(
    "A→B→Aの再訪後も最初のAの%sを現在のAへ適用しない",
    async (outcome) => {
      const old = deferred();
      const current = deferred();
      const date = futureDate(30);
      const draft = futureDate(60);
      const refreshed = futureDate(75);
      const { result, rerender, setMutation } = renderScopedIntent();
      setMutation.mutateAsync
        .mockImplementationOnce(() => old.promise)
        .mockImplementationOnce(() => current.promise);
      await act(async () => result.current.handleIntentDateChange(date));
      const oldReset = result.current.resetIntentDraft;
      rerender({ petId: "20", clinicId: "1" });
      rerender({ petId: "10", clinicId: "1" });
      await act(async () => result.current.handleIntentDateChange(draft));
      // UIのpendingはfreshでも、同一petへのwriteは旧Aの後に実行する。
      expect(setMutation.mutateAsync).toHaveBeenCalledTimes(1);
      rows.set(scopeKey("10", "1"), [tentative(refreshed)]);
      rerender({ petId: "10", clinicId: "1" });
      act(() => oldReset());
      await act(async () => {
        if (outcome === "success") old.resolve();
        else old.reject(new Error("first A failed"));
      });
      expect(result.current.intentFieldValue).toBe(draft);
      expect(toast.success).not.toHaveBeenCalled();
      expect(setMutation.mutateAsync).toHaveBeenCalledTimes(2);
      await act(async () => current.resolve());
      expect(result.current.intentFieldValue).toBe(refreshed);
    },
  );

  it.each(["success", "failure"])("旧scopeのclear%sでも新scopeのsetを妨げない", async (outcome) => {
    const old = deferred();
    const current = deferred();
    const draft = futureDate(60);
    const { result, rerender, setMutation, clearMutation } = renderScopedIntent();
    clearMutation.mutateAsync.mockImplementationOnce(() => old.promise);
    setMutation.mutateAsync.mockImplementationOnce(() => current.promise);
    await act(async () => result.current.handleIntentDateChange(""));
    rerender({ petId: "20", clinicId: "2" });
    await act(async () => result.current.handleIntentDateChange(draft));
    expect(clearMutation.mutateAsync).toHaveBeenCalledWith({ petId: "10", clinicId: "1" });
    expect(setMutation.mutateAsync).toHaveBeenCalledWith({
      petId: "20",
      clinicId: "2",
      intendedDate: draft,
      medicalRecordId: undefined,
    });
    await act(async () => {
      if (outcome === "success") old.resolve();
      else old.reject(new Error("old clear failed"));
    });
    expect(result.current.intentFieldValue).toBe(draft);
    expect(toast.success).not.toHaveBeenCalled();
    await act(async () => current.resolve());
    expect(toast.success).toHaveBeenCalledTimes(1);
  });

  it("同じpetの医院切替でもdraftはfreshにしwrite順序は維持する", async () => {
    const old = deferred();
    const current = deferred();
    const draft = futureDate(60);
    const { result, rerender, setMutation } = renderScopedIntent();
    setMutation.mutateAsync
      .mockImplementationOnce(() => old.promise)
      .mockImplementationOnce(() => current.promise);
    await act(async () => result.current.handleIntentDateChange(futureDate(30)));
    rerender({ petId: "10", clinicId: "2" });
    expect(result.current.intentFieldValue).toBe("");
    await act(async () => result.current.handleIntentDateChange(draft));
    expect(setMutation.mutateAsync).toHaveBeenCalledTimes(1);
    await act(async () => old.resolve());
    expect(setMutation.mutateAsync).toHaveBeenCalledTimes(2);
    expect(setMutation.mutateAsync).toHaveBeenLastCalledWith({
      petId: "10",
      clinicId: "2",
      intendedDate: draft,
      medicalRecordId: undefined,
    });
    expect(result.current.intentFieldValue).toBe(draft);
    expect(toast.success).not.toHaveBeenCalled();
    await act(async () => current.resolve());
  });

  it.each(["success", "failure"])(
    "unmount後の旧mutation%sで成功通知を出さない",
    async (outcome) => {
      const pending = deferred();
      const { result, unmount, setMutation } = renderScopedIntent();
      setMutation.mutateAsync.mockImplementationOnce(() => pending.promise);
      await act(async () => result.current.handleIntentDateChange(futureDate(30)));
      unmount();
      await act(async () => {
        if (outcome === "success") pending.resolve();
        else pending.reject(new Error("unmounted request failed"));
      });
      expect(toast.success).not.toHaveBeenCalled();
    },
  );

  it.each(["set", "clear"])(
    "TanStack options更新後も旧scopeのqueued %sは要求元の医院へ送る",
    async (operation) => {
      const pending = deferred();
      const date = futureDate(30);
      const nextDate = futureDate(60);
      const response = {
        data: { id: 1, start_time: `${date}T00:00:00+09:00`, end_time: `${date}T00:00:00+09:00` },
      };
      vi.mocked(axios.put)
        .mockImplementationOnce(async () => {
          await pending.promise;
          return response;
        })
        .mockResolvedValue(response);
      vi.mocked(axios.delete).mockResolvedValue({});
      const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
      const wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
      );
      const { result, rerender, unmount } = renderHook(
        (scope: { petId: string; clinicId: string }) => {
          const setIntentMutation = useSetNextVisitIntent(scope.clinicId);
          const clearIntentMutation = useClearNextVisitIntent(scope.clinicId);
          return useNextVisitIntent({
            ...scope,
            isPetDeceased: false,
            setIntentMutation,
            clearIntentMutation,
          });
        },
        { wrapper, initialProps: { petId: "10", clinicId: "1" } },
      );
      await act(async () => result.current.handleIntentDateChange(date));
      await act(async () =>
        result.current.handleIntentDateChange(operation === "set" ? nextDate : ""),
      );
      expect(axios.put).toHaveBeenCalledTimes(1);
      expect(axios.delete).not.toHaveBeenCalled();
      rerender({ petId: "20", clinicId: "2" });
      await act(async () => result.current.handleIntentDateChange(nextDate));
      await waitFor(() => expect(axios.put).toHaveBeenCalledTimes(2));
      expect(axios.put).toHaveBeenNthCalledWith(
        2,
        "/v1/reservations/pets/20/next-visit-intent",
        { intended_date: nextDate },
        { headers: { "X-Clinic-ID": "2" } },
      );
      await act(async () => pending.resolve());
      if (operation === "set") {
        await waitFor(() => expect(axios.put).toHaveBeenCalledTimes(3));
        expect(axios.put).toHaveBeenNthCalledWith(
          3,
          "/v1/reservations/pets/10/next-visit-intent",
          { intended_date: nextDate },
          { headers: { "X-Clinic-ID": "1" } },
        );
      } else {
        await waitFor(() => expect(axios.delete).toHaveBeenCalledTimes(1));
        expect(axios.delete).toHaveBeenCalledWith("/v1/reservations/pets/10/next-visit-intent", {
          headers: { "X-Clinic-ID": "1" },
        });
      }
      unmount();
      client.clear();
    },
  );
});

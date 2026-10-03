import type { ComponentProps } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import { createWrapper, noop } from "./ReservationFormModal.test-helpers";

/**
 * EMR-116 (SLACK-STAFF-SELECT): iPad/一部PCで担当者(獣医師)が選べない申告の
 * stub-free 再現ハーネス。
 *
 * 既存 ReservationFormModal.staff-candidates.test.tsx は SearchableSelect を
 * vi.mock で差し替えているため、Dialog×portaled Popover×cmdk の実経路は未検証だった。
 * 本ファイルは製品コンポーネントを一切 stub せず、実物の親 Dialog + Popover +
 * SearchableSelect + cmdk を mount して、
 *   候補表示 → 選択 → フォーム値 → 保存 payload → 再表示
 * を fine pointer(mouse)/keyboard/coarse pointer(touch) の 3 経路で比較する。
 *
 * 「候補 0 件 / 読込中 / 取得失敗」と「候補は出るのに選択できない」を分離するため、
 * 各テストは option の render 有無と選択後の trigger 表示・保存 payload を別々に検査する。
 */

const CAPABLE_STAFF = { id: 11, name: "鈴木" } as const;
const INCAPABLE_STAFF = { id: 10, name: "三井" } as const;

const STAFF_SELECT_HANDLERS = [
  http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
  http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
  http.get("/api/v1/masters/animal-species", () =>
    HttpResponse.json([{ id: 1, name: "犬", is_active: true }]),
  ),
  http.get("/api/v1/masters/staffs", () =>
    HttpResponse.json([
      {
        id: INCAPABLE_STAFF.id,
        name: INCAPABLE_STAFF.name,
        is_active: true,
        clinic_assignments: [{ clinic_id: 1, is_main: true }],
      },
      {
        id: CAPABLE_STAFF.id,
        name: CAPABLE_STAFF.name,
        is_active: true,
        clinic_assignments: [{ clinic_id: 1, is_main: true }],
      },
    ]),
  ),
  http.get("/api/v1/shifts/on-duty-staffs", () =>
    HttpResponse.json([
      { id: INCAPABLE_STAFF.id, name: INCAPABLE_STAFF.name },
      { id: CAPABLE_STAFF.id, name: CAPABLE_STAFF.name },
    ]),
  ),
  http.get("/api/v1/clinics/1/reservation-staffs", () =>
    HttpResponse.json([
      { id: INCAPABLE_STAFF.id, name: INCAPABLE_STAFF.name, is_active: true, capable_courses: [] },
      {
        id: CAPABLE_STAFF.id,
        name: CAPABLE_STAFF.name,
        is_active: true,
        capable_courses: [{ id: 5, name: "診察" }],
      },
    ]),
  ),
  http.get("/api/v1/masters/reservation-types/5/unavailable-times", () =>
    HttpResponse.json({ data: [] }),
  ),
  http.get("/api/v1/reservations/available-times", () =>
    HttpResponse.json([{ start_time: "1000", end_time: "1100" }]),
  ),
  http.get("/api/v1/masters/reservation-types", () =>
    HttpResponse.json([
      {
        id: 5,
        name: "診察",
        color: "#111111",
        is_active: true,
        duration_minutes: 60,
        sort_order: 1,
        is_internal: false,
        category: "medical",
        group_id: null,
        group: null,
      },
    ]),
  ),
];

/** 仮端末条件: matchMedia("(pointer: coarse)") を true にする環境エミュレーション。 */
function emulateCoarsePointer() {
  const original = window.matchMedia;
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    value: (query: string) => ({
      matches: query === "(pointer: coarse)",
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    }),
  });
  return () => {
    Object.defineProperty(window, "matchMedia", { writable: true, value: original });
  };
}

/**
 * touch tap のブラウザ pointer イベント列をそのまま流す coarse-pointer 補助経路。
 * jsdom は TouchEvent.touches を構築できず react-remove-scroll が touches[0] で
 * 落ちるため、Radix/cmdk が実際に消費する pointer* + click 列だけを発火する
 * (pointerType=touch は iPad の DismissableLayer/focus 判定が見る値)。
 */
async function touchTap(target: Element) {
  fireEvent.pointerOver(target, { pointerType: "touch" });
  fireEvent.pointerDown(target, { pointerType: "touch" });
  fireEvent.pointerUp(target, { pointerType: "touch" });
  fireEvent.click(target);
}

function renderReservationModal(
  onSave: ComponentProps<typeof ReservationFormModal>["onSave"] = noop,
) {
  const start = new Date();
  start.setDate(start.getDate() + 1);
  start.setHours(10, 0, 0, 0);
  const end = new Date(start);
  end.setHours(11, 0, 0, 0);
  return render(
    <ReservationFormModal
      isOpen={true}
      onClose={noop}
      onSave={onSave}
      initialData={{ start, end }}
      canCreate={true}
      canEdit={false}
    />,
    { wrapper: createWrapper() },
  );
}

async function fillNewOwnerBasics(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByTestId("mode-new"));
  fireEvent.change(screen.getByTestId("new-owner-name"), { target: { value: "山田太郎" } });
  fireEvent.change(screen.getByTestId("new-owner-phone"), {
    target: { value: "+81 90 1234 5678" },
  });
  fireEvent.change(screen.getByTestId("new-owner-pet-name"), { target: { value: "ポチ" } });
  fireEvent.change(screen.getByTestId("new-owner-chief-complaint"), {
    target: { value: "食欲不振" },
  });
  const speciesTrigger = screen.getByTestId("new-owner-species");
  await waitFor(() => {
    expect(speciesTrigger).toBeEnabled();
    expect(speciesTrigger).not.toHaveTextContent("読み込み中");
  });
  await user.click(speciesTrigger);
  await user.click(await screen.findByRole("option", { name: "犬" }));
}

async function selectReservationType() {
  await userEvent.setup({ delay: null }).click(screen.getByTestId("res-type-trigger"));
  fireEvent.click(await screen.findByTestId("res-type-card-5"));
  await waitFor(() => {
    expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("診察");
  });
}

/** 担当者ポップオーバーを開き、capability フィルタ後の候補 render を検査して返す。 */
async function openStaffPopoverAndExpectCandidates() {
  const trigger = screen.getByTestId("res-staff-trigger");
  expect(trigger).toHaveAttribute("aria-expanded", "true");
  // 候補 present vs empty/loading/failure の分離: option role の実在を先に確かめる。
  const option = await screen.findByRole("option", { name: CAPABLE_STAFF.name });
  expect(screen.queryByRole("option", { name: INCAPABLE_STAFF.name })).not.toBeInTheDocument();
  expect(screen.queryByText("スタッフ候補を読み込み中です")).not.toBeInTheDocument();
  expect(screen.queryByText("スタッフ候補の取得に失敗しました")).not.toBeInTheDocument();
  return { trigger, option };
}

async function expectStaffSelectedAndSaved(onSave: ReturnType<typeof vi.fn>) {
  await waitFor(() => {
    expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent(CAPABLE_STAFF.name);
  });
  await userEvent.setup({ delay: null }).click(screen.getByRole("button", { name: "予約を確定" }));
  await waitFor(() => {
    expect(onSave).toHaveBeenCalledOnce();
  });
  expect(onSave.mock.calls[0][0]).toMatchObject({
    doctor: String(CAPABLE_STAFF.id),
    type: "5",
  });
}

afterEach(() => {
  server.resetHandlers();
  localStorage.removeItem("auth_current_clinic:v1");
});

describe("ReservationFormModal 担当者選択 — stub-free device matrix (EMR-116)", () => {
  it("mouse: 候補 render → click 選択 → フォーム値 → 保存 payload", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(...STAFF_SELECT_HANDLERS);
    const onSave = vi.fn();
    const user = userEvent.setup({ delay: null });

    renderReservationModal(onSave);
    await fillNewOwnerBasics(user);
    await selectReservationType();

    await user.click(screen.getByTestId("res-staff-trigger"));
    const { option } = await openStaffPopoverAndExpectCandidates();
    await user.click(option);
    await expectStaffSelectedAndSaved(onSave);
  }, 20000);

  it("keyboard: focus → Enter で開き cmdk ArrowDown/Enter で選択 → 保存 payload", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(...STAFF_SELECT_HANDLERS);
    const onSave = vi.fn();
    const user = userEvent.setup({ delay: null });

    renderReservationModal(onSave);
    await fillNewOwnerBasics(user);
    await selectReservationType();

    const trigger = screen.getByTestId("res-staff-trigger");
    trigger.focus();
    await user.keyboard("{Enter}");
    await openStaffPopoverAndExpectCandidates();
    // cmdk は開放時に先頭候補を highlight する。ArrowDown/Enter で明示選択する。
    // 選択成立で Popover は閉じ option は unmount されるため、選択後は trigger 表示を見る。
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{ArrowUp}");
    await user.keyboard("{Enter}");
    await waitFor(() => {
      expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent(CAPABLE_STAFF.name);
    });
    await user.click(screen.getByRole("button", { name: "予約を確定" }));
    await waitFor(() => {
      expect(onSave).toHaveBeenCalledOnce();
    });
    expect(onSave.mock.calls[0][0]).toMatchObject({
      doctor: String(CAPABLE_STAFF.id),
      type: "5",
    });
  }, 20000);

  it("coarse pointer: (pointer: coarse) + touch tap で候補選択 → 保存 payload", async () => {
    const restorePointer = emulateCoarsePointer();
    try {
      localStorage.setItem("auth_current_clinic:v1", "1");
      server.use(...STAFF_SELECT_HANDLERS);
      const onSave = vi.fn();
      const user = userEvent.setup({ delay: null });

      renderReservationModal(onSave);
      await fillNewOwnerBasics(user);
      await selectReservationType();

      // iPad 系 touch tap: pointerType=touch の pointer 列 + click。
      await touchTap(screen.getByTestId("res-staff-trigger"));
      const { option } = await openStaffPopoverAndExpectCandidates();
      // coarse では onOpenAutoFocus が抑制され、検索 input へ強制 focus しない。
      expect(document.activeElement).not.toBe(
        document.querySelector('[data-slot="command-input"]'),
      );
      await touchTap(option);
      await expectStaffSelectedAndSaved(onSave);
    } finally {
      restorePointer();
    }
  }, 20000);

  it("user.pointer TouchA: user-event の touch デバイス経路でも選択できる", async () => {
    const restorePointer = emulateCoarsePointer();
    try {
      localStorage.setItem("auth_current_clinic:v1", "1");
      server.use(...STAFF_SELECT_HANDLERS);
      const user = userEvent.setup({ delay: null });

      renderReservationModal(noop);
      await selectReservationType();

      await user.pointer([{ keys: "[TouchA]", target: screen.getByTestId("res-staff-trigger") }]);
      const { option } = await openStaffPopoverAndExpectCandidates();
      await user.pointer([{ keys: "[TouchA]", target: option }]);
      await waitFor(() => {
        expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent(CAPABLE_STAFF.name);
      });
    } finally {
      restorePointer();
    }
  }, 20000);

  it("候補 0 件は empty message で示され、present-but-unselectable と区別できる", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(...STAFF_SELECT_HANDLERS);
    // msw: 後から use した handler が同一路径を上書きする。
    server.use(
      http.get("/api/v1/clinics/1/reservation-staffs", () =>
        HttpResponse.json([
          {
            id: INCAPABLE_STAFF.id,
            name: INCAPABLE_STAFF.name,
            is_active: true,
            capable_courses: [],
          },
          {
            id: CAPABLE_STAFF.id,
            name: CAPABLE_STAFF.name,
            is_active: true,
            capable_courses: [],
          },
        ]),
      ),
    );
    const user = userEvent.setup({ delay: null });

    renderReservationModal(noop);
    await selectReservationType();

    await user.click(screen.getByTestId("res-staff-trigger"));
    await waitFor(() => {
      expect(screen.getByText("この条件で対応可能なスタッフがいません")).toBeInTheDocument();
    });
    expect(screen.queryByRole("option")).not.toBeInTheDocument();
  }, 20000);

  it("capability metadata 未取得は「読み込み中」で示され失敗/空とは区別される", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    let release: (() => void) | undefined;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    server.use(...STAFF_SELECT_HANDLERS);
    server.use(
      http.get("/api/v1/clinics/1/reservation-staffs", async () => {
        await gate;
        return HttpResponse.json([
          { id: 11, name: "鈴木", is_active: true, capable_courses: [{ id: 5, name: "診察" }] },
        ]);
      }),
    );
    const user = userEvent.setup({ delay: null });

    try {
      renderReservationModal(noop);
      await selectReservationType();

      await user.click(screen.getByTestId("res-staff-trigger"));
      await waitFor(() => {
        expect(screen.getByText("スタッフ候補を読み込み中です")).toBeInTheDocument();
      });
    } finally {
      // gate を必ず解放し、テスト失敗時も pending handler を残さない。
      release?.();
    }
  }, 20000);

  it("reservation-staffs 取得失敗は error message で示される", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(...STAFF_SELECT_HANDLERS);
    server.use(
      http.get("/api/v1/clinics/1/reservation-staffs", () =>
        HttpResponse.json({ message: "boom" }, { status: 500 }),
      ),
    );
    const user = userEvent.setup({ delay: null });

    renderReservationModal(noop);
    await selectReservationType();

    await user.click(screen.getByTestId("res-staff-trigger"));
    await waitFor(() => {
      expect(screen.getByText("スタッフ候補の取得に失敗しました")).toBeInTheDocument();
    });
    expect(screen.queryByRole("option", { name: CAPABLE_STAFF.name })).not.toBeInTheDocument();
  }, 20000);

  it("再表示: 保存済み担当者を含む編集初期表示が名前を維持する", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      ...STAFF_SELECT_HANDLERS,
      http.get("/api/v1/reservations/99", () =>
        HttpResponse.json({
          id: 99,
          start_time: "1000",
          end_time: "1100",
          reservation_route: null,
        }),
      ),
    );

    const start = new Date();
    start.setDate(start.getDate() + 1);
    start.setHours(10, 0, 0, 0);
    const end = new Date(start);
    end.setHours(11, 0, 0, 0);

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={{
          id: "99",
          start,
          end,
          type: "5",
          doctor: String(CAPABLE_STAFF.id),
          visitType: "revisit",
          status: "confirmed",
        }}
        canCreate={false}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    await waitFor(() => {
      expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent(CAPABLE_STAFF.name);
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("診察");
    });
    // re-display は保存値の書き換え・placeholder 落ちをしない。
    expect(screen.getByTestId("res-staff-trigger")).not.toHaveTextContent("選択してください");
  }, 20000);
});

import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import type { Reservation } from "@/types";
import type { Pet } from "@/types";
import { http, HttpResponse } from "msw";
import { createWrapper, silentApiHandlers, noop } from "./ReservationFormModal.test-helpers";

// type が初期値にあると available-times クエリが走る。空配列 = unset 応答で黙らせる。
const confirmModeHandlers = [
  ...silentApiHandlers,
  http.get("/api/v1/reservations/available-times", () => HttpResponse.json([])),
];

// SearchableSelect は Radix Popover を Radix Dialog の内側で開く。jsdom + カバレッジ計装下では
// Dialog の FocusScope が focus を掴み直し続け、開いた Popover が focus-outside 判定で即座に
// 閉じるため、CI でのみ aria-expanded が false のまま option が現れない。2026-08-23 に CI 上で
// 実測して確認した（click は届いており、body/trigger の pointer-events も disabled も正常。
// fireEvent.click でも開かない）。本テストの対象は Popover の開閉実装ではないので、開閉の
// 意味論だけを保った素の実装へ差し替え、Dialog×Popover の相互作用を構造的に取り除く。
vi.mock("@/components/ui/searchable-select", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/components/ui/searchable-select")>();
  const { useState } = await import("react");
  type Props = Parameters<typeof actual.SearchableSelect>[0];
  function SearchableSelectStub(props: Props) {
    const [open, setOpen] = useState(false);
    const flat = props.groups ? props.groups.flatMap((g) => g.options) : (props.options ?? []);
    const selected = flat.find((o) => o.value === props.value);
    return (
      <div>
        <button
          type="button"
          role="combobox"
          id={props.id}
          aria-label={props.ariaLabel}
          aria-expanded={open}
          aria-invalid={props.ariaInvalid}
          aria-describedby={props.ariaDescribedBy}
          disabled={props.disabled}
          data-testid={props.triggerTestId}
          className={props.className}
          onClick={() => setOpen((v) => !v)}
        >
          {selected ? selected.label : props.placeholder}
        </button>
        {open
          ? flat.map((o) => (
              <button
                key={o.value}
                type="button"
                role="option"
                aria-selected={o.value === props.value}
                disabled={o.disabled}
                onClick={() => {
                  props.onValueChange(o.value);
                  setOpen(false);
                }}
              >
                {o.label}
              </button>
            ))
          : null}
      </div>
    );
  }
  return { ...actual, SearchableSelect: SearchableSelectStub };
});

// tentative 確定モードは患者固定: initialData.petId → useGetPet が対象患者を返す。
// 検索用 useGetPets は使われないが、モジュールモックのため両方をスタブする。
const LOCKED_PET = {
  id: "10",
  name: "ポチ",
  species: "犬",
  ownerId: "20",
  ownerName: "山田",
  status: "生存",
  deceasedAt: null,
} as unknown as Pet;

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({ data: LOCKED_PET }),
  useGetPets: () => ({ data: [] }),
}));

afterEach(() => {
  server.resetHandlers();
  localStorage.removeItem("auth_current_clinic:v1");
});

const futureStart = () => new Date(2999, 5, 20, 10, 0); // 2999-06-20 10:00（過去日バリデーション回避）

const confirmInitialData: Partial<Reservation> = {
  petId: "10",
  start: futureStart(),
  end: new Date(2999, 5, 20, 11, 0),
  visitType: "revisit",
  type: "1",
  doctor: "",
  isDesignated: false,
  status: "confirmed",
};

describe("ReservationFormModal — tentative 確定モード (EMR-308)", () => {
  it("タイトルは「次回予定の確定」、患者は固定表示、送信は 予約を確定", () => {
    server.use(...confirmModeHandlers);
    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onConfirmTentative={noop}
        tentativeConfirmId="t-1"
        initialData={confirmInitialData}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    expect(screen.getByRole("heading", { name: "次回予定の確定" })).toBeInTheDocument();
    // 患者固定（検索UIは出さない）
    expect(screen.getByText("対象患者")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "予約を確定" })).toBeInTheDocument();
  }, 15000);

  it("送信で onConfirmTentative(tentativeId, formData) が呼ばれ onSave は呼ばれない", async () => {
    server.use(...confirmModeHandlers);
    const user = userEvent.setup();
    const onConfirmTentative = vi.fn().mockResolvedValue(null);
    const onSave = vi.fn();

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={onSave}
        onConfirmTentative={onConfirmTentative}
        tentativeConfirmId="t-1"
        initialData={confirmInitialData}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    // initialData.petId → useGetPet で対象患者が選択済みになるのを待つ
    // （LockedPatientNotice と SelectedPetChip の2箇所に名前が出るため getAllByText）
    await waitFor(() => expect(screen.getAllByText("ポチ").length).toBeGreaterThan(0));

    await user.click(screen.getByRole("button", { name: "予約を確定" }));

    await waitFor(() => expect(onConfirmTentative).toHaveBeenCalledTimes(1));
    const [id, formData] = onConfirmTentative.mock.calls[0] as [
      string,
      Partial<Reservation>,
      ...unknown[],
    ];
    expect(id).toBe("t-1");
    expect(formData.start).toEqual(futureStart());
    expect(formData.type).toBe("1");
    expect(onSave).not.toHaveBeenCalled();
  }, 20000);

  it("onConfirmTentative がエラー文字列を返すとインライン表示して開いたまま", async () => {
    server.use(...confirmModeHandlers);
    const user = userEvent.setup();
    const onConfirmTentative = vi
      .fn()
      .mockResolvedValue("予約の確定に失敗しました。入力内容を確認して再度お試しください。");

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onConfirmTentative={onConfirmTentative}
        tentativeConfirmId="t-1"
        initialData={confirmInitialData}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(screen.getAllByText("ポチ").length).toBeGreaterThan(0));
    await user.click(screen.getByRole("button", { name: "予約を確定" }));

    await waitFor(() =>
      expect(screen.getByRole("alert")).toHaveTextContent("予約の確定に失敗しました"),
    );
    // モーダルは開いたまま
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  }, 20000);

  it("canEdit=false では確定ボタンを出さない（fail-closed）", () => {
    server.use(...confirmModeHandlers);
    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onConfirmTentative={noop}
        tentativeConfirmId="t-1"
        initialData={confirmInitialData}
        canEdit={false}
      />,
      { wrapper: createWrapper() },
    );

    expect(screen.queryByRole("button", { name: "予約を確定" })).not.toBeInTheDocument();
  }, 15000);
});

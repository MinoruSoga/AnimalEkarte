import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Reservation } from "../types";
import type { ReservationTypeColor } from "@/hooks/use-reservation-type-color-map";
import { ReservationDetailModal } from "./ReservationDetailModal";
import { useGetReservation } from "@/hooks/use-get-reservation";

// PR#614 指摘: modal はクリック時 snapshot を GET /reservations/:id の最新応答で
// 上書き追従する。既定は未取得（snapshot のまま表示）。
const { mockGetReservation, mockGetOwnerLineTags } = vi.hoisted(() => ({
  mockGetReservation: { current: { data: undefined as unknown } },
  mockGetOwnerLineTags: vi.fn((_ownerId: string) => ({ data: undefined as unknown })),
}));
vi.mock("@/hooks/use-get-reservation", () => ({
  useGetReservation: vi.fn(() => mockGetReservation.current),
}));
vi.mock("@/hooks/use-owner-line-tags", () => ({
  useGetOwnerLineTags: mockGetOwnerLineTags,
}));

// EMR-227/EMR-301: 死亡区分は wire 予約サマリの petStatus/petDeceasedAt で判定する
//（/pets/:id 取得は owners:view 権限依存のため廃止）。
const getColorMock = vi.fn((): ReservationTypeColor => ({
  style: { backgroundColor: "rgb(0, 0, 0)" },
  dotStyle: { backgroundColor: "rgb(0, 0, 0)" },
  hex: "#000000",
}));

vi.mock("../hooks/use-reservation-type-color-map", () => ({
  useReservationTypeColorMap: () => ({
    getColor: getColorMock,
  }),
}));

const baseReservation: Reservation = {
  id: "88",
  start: new Date("2026-05-29T03:30:00.000Z"),
  end: new Date("2026-05-29T04:00:00.000Z"),
  ownerName: "山田花子",
  petName: "ポチ",
  petType: "トイプードル",
  visitType: "revisit",
  type: "一般診察",
  category: "general",
  reservationTypeId: "1",
  doctor: "担当者A",
  doctorId: "33",
  isDesignated: false,
  status: "checked_in",
  notes: undefined,
  petId: "10",
  ownerId: "20",
  source: "manual",
  reservationRoute: "reception",
  clinicId: "1",
  isTentative: false,
  intendedDate: undefined,
  hadIntendedDate: false,
  sourceMedicalRecordId: undefined,
};

describe("ReservationDetailModal", () => {
  it("予約内容として既存notesを表示する", () => {
    render(
      <ReservationDetailModal
        isOpen
        onClose={vi.fn()}
        reservation={{ ...baseReservation, notes: "既存の予約内容" }}
      />,
    );
    expect(screen.getByText("予約内容")).toBeInTheDocument();
    expect(screen.getByText("既存の予約内容")).toBeInTheDocument();
    expect(screen.queryByText("メモ")).not.toBeInTheDocument();
  });

  beforeEach(() => {
    vi.mocked(useGetReservation).mockClear();
    mockGetReservation.current = { data: undefined };
    mockGetOwnerLineTags.mockClear();
    getColorMock.mockReset();
    getColorMock.mockReturnValue({
      style: { backgroundColor: "rgb(0, 0, 0)" },
      dotStyle: { backgroundColor: "rgb(0, 0, 0)" },
      hex: "#000000",
    });
  });

  it("通常診療予約ではカルテ作成ボタンを表示する", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={baseReservation}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: /カルテ作成/ })).toBeInTheDocument();
  });

  it("別医院の予約詳細とLINEタグを行の医院で取得し、最新死亡状態に追従する", () => {
    mockGetReservation.current = {
      data: { ...baseReservation, clinicId: "2", petStatus: "deceased" },
    };
    render(
      <ReservationDetailModal
        isOpen
        onClose={vi.fn()}
        reservation={{ ...baseReservation, clinicId: "2" }}
        onCreateRecord={vi.fn()}
      />,
    );
    expect(useGetReservation).toHaveBeenCalledWith("88", "2");
    expect(mockGetOwnerLineTags).toHaveBeenLastCalledWith("20", "2");
    expect(screen.queryByRole("button", { name: /カルテ作成/ })).not.toBeInTheDocument();
  });

  it("トリミングカテゴリの予約では区分名に依存せずトリミング記録作成ボタンを表示する", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{
          ...baseReservation,
          type: "シャンプーコース",
          category: "trimming",
          reservationTypeId: "9",
        }}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: /トリミング記録作成/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /カルテ作成/ })).not.toBeInTheDocument();
  });

  it("ペットホテル予約ではマスタ名の部分一致で入院・ホテル登録ボタンを表示する", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{
          ...baseReservation,
          type: "ペットホテル",
          category: "general",
          reservationTypeId: "12",
        }}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: /入院・ホテル登録/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /カルテ作成/ })).not.toBeInTheDocument();
  });

  it("無効区分では DialogTitle と予約区分に（無効）を付ける (BUG-016)", () => {
    getColorMock.mockReturnValue({
      style: { backgroundColor: "rgba(255, 102, 0, 0.1)" },
      dotStyle: { backgroundColor: "#ff6600" },
      hex: "#ff6600",
      isInactive: true,
    });

    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={baseReservation}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("heading", { name: "一般診察（無効）" })).toBeInTheDocument();
    const typeRow = screen.getByText("予約区分").parentElement;
    expect(typeRow?.textContent).toContain("一般診察（無効）");
  });

  it("有効区分では（無効）を付けない (BUG-016)", () => {
    getColorMock.mockReturnValue({
      style: { backgroundColor: "rgba(17, 17, 17, 0.1)" },
      dotStyle: { backgroundColor: "#111111" },
      hex: "#111111",
      isInactive: false,
    });

    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={baseReservation}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("heading", { name: "一般診察" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /（無効）/ })).not.toBeInTheDocument();
    const typeRow = screen.getByText("予約区分").parentElement;
    expect(typeRow?.textContent).toContain("一般診察");
    expect(typeRow?.textContent).not.toContain("（無効）");
  });

  // PR#614 指摘: reservation prop はクリック時 snapshot で、死亡登録後も生存時の
  // sentinel を保持する。detail 取得の最新応答が返ったら表示・導線をそちらに追従させる。
  it("snapshot が生存でも detail 応答が死亡ならカルテ作成を隠す", () => {
    mockGetReservation.current = {
      data: { ...baseReservation, petStatus: "deceased" },
    };
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{ ...baseReservation, petBirthDate: "2018-05-10" }}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.queryByRole("button", { name: /カルテ作成/ })).not.toBeInTheDocument();
  });

  it("detail 応答が未着なら snapshot のまま表示する（fail しない）", () => {
    mockGetReservation.current = { data: undefined };
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={baseReservation}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: /カルテ作成/ })).toBeInTheDocument();
  });

  // PR#614 再レビュー指摘2: LINE タグは表示・操作対象と同一 owner を見るため、
  // snapshot ではなく effective 予約（fresh GET 優先・fallback snapshot）の
  // ownerId で引く。別 owner を引くと警告が誤った飼主を指す。
  it("detail 応答の ownerId が snapshot と異なる場合 LINE タグは detail 側 ownerId で引く", () => {
    mockGetReservation.current = {
      data: { ...baseReservation, ownerId: "owner-b", ownerName: "佐藤一郎" },
    };
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{ ...baseReservation, ownerId: "owner-a" }}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(mockGetOwnerLineTags).toHaveBeenLastCalledWith("owner-b", "1");
  });

  it("detail 応答が未着なら snapshot の ownerId で LINE タグを引く（fallback）", () => {
    mockGetReservation.current = { data: undefined };
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{ ...baseReservation, ownerId: "owner-a" }}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(mockGetOwnerLineTags).toHaveBeenLastCalledWith("owner-a", "1");
  });

  it("reservation 未確定では ownerId 空を渡して LINE タグ query を無効化する", () => {
    render(<ReservationDetailModal isOpen={true} onClose={vi.fn()} reservation={null} />);

    expect(mockGetOwnerLineTags).toHaveBeenLastCalledWith("", undefined);
  });
});

describe("ReservationDetailModal — tentative (EMR-308)", () => {
  // tentative は start/end が意図日 00:00–00:15 JST の sentinel。実時刻ではないので
  // 時刻表示・ステータスセレクト・カルテ作成・編集の各導線を出さず、
  // 「予約を確定」（confirm-tentative 経路）のみを出す。
  // 意図日は未来に固定する — 過去日だと isExpiredTentativeReservation が真になり
  // ラベルが 期限超過・未対応 に変わるため。
  const tentativeReservation: Reservation = {
    ...baseReservation,
    id: "tentative-1",
    start: new Date("2999-06-29T00:00:00.000Z"),
    end: new Date("2999-06-29T00:15:00.000Z"),
    status: "tentative" as Reservation["status"],
    isTentative: true,
    intendedDate: "2999-06-29",
  };

  it("次回予定・未予約 ラベルと 時刻未設定 を表示する", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={tentativeReservation}
        onCreateRecord={vi.fn()}
        onStatusChange={vi.fn()}
        onEdit={vi.fn()}
        onConfirmTentative={vi.fn()}
      />,
    );

    expect(screen.getByText("次回予定・未予約")).toBeInTheDocument();
    expect(screen.getByText("時刻未設定（予定日のみ）")).toBeInTheDocument();
  });

  it("tentative 行ではカルテ作成・編集・ステータス変更・削除を出さない", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={tentativeReservation}
        onCreateRecord={vi.fn()}
        onStatusChange={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onConfirmTentative={vi.fn()}
      />,
    );

    // カルテ作成は appointmentId 付き遷移で sentinel 行に直接紐づき昇格しないため出さない
    expect(screen.queryByRole("button", { name: /カルテ作成/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /編集/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
  });

  it("意図日が過去の tentative は 期限超過・未対応 ラベルになる", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{
          ...tentativeReservation,
          start: new Date("2000-01-01T00:00:00.000Z"),
          end: new Date("2000-01-01T00:15:00.000Z"),
          intendedDate: "2000-01-01",
        }}
        onConfirmTentative={vi.fn()}
      />,
    );

    expect(screen.getByText("期限超過・未対応")).toBeInTheDocument();
    expect(screen.queryByText("次回予定・未予約")).not.toBeInTheDocument();
  });

  it("onConfirmTentative 指定時は予約を確定ボタンで予約を渡す", async () => {
    const user = userEvent.setup();
    const onConfirmTentative = vi.fn();
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={tentativeReservation}
        onConfirmTentative={onConfirmTentative}
      />,
    );

    await user.click(screen.getByRole("button", { name: "予約を確定" }));
    expect(onConfirmTentative).toHaveBeenCalledWith(tentativeReservation);
  });

  it("onConfirmTentative 省略時は確定ボタンを出さない（fail-closed）", () => {
    render(
      <ReservationDetailModal isOpen={true} onClose={vi.fn()} reservation={tentativeReservation} />,
    );

    expect(screen.queryByRole("button", { name: "予約を確定" })).not.toBeInTheDocument();
  });

  it("死亡ペットの tentative では確定ボタンを出さない（EMR-177 positive-match ゲート）", () => {
    // 確定は実予約化する臨床 write — BE 側 ConfirmTentative も死亡ガード済み。
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{ ...tentativeReservation, petStatus: "deceased" }}
        onConfirmTentative={vi.fn()}
      />,
    );

    expect(screen.queryByRole("button", { name: "予約を確定" })).not.toBeInTheDocument();
  });

  it("deceased_at のみの半死亡ペットでも確定ボタンを出さない（死亡OR契約）", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{ ...tentativeReservation, petDeceasedAt: "2026-07-10T12:00:00+09:00" }}
        onConfirmTentative={vi.fn()}
      />,
    );

    expect(screen.queryByRole("button", { name: "予約を確定" })).not.toBeInTheDocument();
  });

  it("pet サマリ未取得（不明）の tentative では確定ボタンを出す（BE が最終ゲート）", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={tentativeReservation}
        onConfirmTentative={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: "予約を確定" })).toBeInTheDocument();
  });

  it("昇格済み（hadIntendedDate のみ残る）行は通常予約として扱う", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{
          ...baseReservation,
          hadIntendedDate: true,
          intendedDate: "2026-05-29",
        }}
        onCreateRecord={vi.fn()}
        onEdit={vi.fn()}
        onStatusChange={vi.fn()}
        onConfirmTentative={vi.fn()}
      />,
    );

    // isTentative=false のため tentative ラベルも確定ボタンも出ず、カルテ作成は出る
    expect(screen.queryByText("次回予定・未予約")).not.toBeInTheDocument();
    expect(screen.queryByText("期限超過・未対応")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "予約を確定" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /カルテ作成/ })).toBeInTheDocument();
  });
});

describe("ReservationDetailModal 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    // 現在JST月を 2026-10 に固定して決定的にする
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  function renderModal(reservation: Reservation) {
    render(<ReservationDetailModal isOpen={true} onClose={vi.fn()} reservation={reservation} />);
  }

  it("今月(JST)誕生月のペットはペット名行に Cake バッジを表示する", () => {
    renderModal({ ...baseReservation, petBirthDate: "2018-10-05" });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05"],
    ["birth_date 未登録", undefined],
  ])("%s のペットには Cake バッジを表示しない", (_label, petBirthDate) => {
    renderModal({ ...baseReservation, petBirthDate });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("petStatus=deceased（ワイヤ sentinel）のペットには誕生月でも表示しない", () => {
    renderModal({
      ...baseReservation,
      petBirthDate: "2018-10-05",
      petStatus: "deceased",
    });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("petDeceasedAt のみの半死亡ペットには誕生月でも表示しない（死亡OR契約）", () => {
    renderModal({
      ...baseReservation,
      petBirthDate: "2018-10-05",
      petDeceasedAt: "2026-07-10T12:00:00+09:00",
    });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});

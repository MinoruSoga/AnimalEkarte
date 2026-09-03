import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import type { Reservation } from "@/types";
import { createWrapper, silentApiHandlers, noop } from "./ReservationFormModal.test-helpers";

// SearchableSelect は Radix Popover を Radix Dialog の内側で開く。jsdom + カバレッジ計装下では
// Dialog の FocusScope が focus を掴み直し続け、開いた Popover が focus-outside 判定で即座に
// 閉じるため、CI でのみ aria-expanded が false のまま option が現れない。2026-08-23 に CI 上で
// 実測して確認した（click は届いており、body/trigger の pointer-events も disabled も正常。
// fireEvent.click でも開かない）。本テストの対象は Popover の開閉実装ではないので、開閉の
// 意味論だけを保った素の実装へ差し替え、Dialog×Popover の相互作用を構造的に取り除く。
vi.mock("@/components/ui/searchable-select", async (importOriginal) => {
  const actual =
    await importOriginal<typeof import("@/components/ui/searchable-select")>();
  const { useState } = await import("react");
  type Props = Parameters<typeof actual.SearchableSelect>[0];
  function SearchableSelectStub(props: Props) {
    const [open, setOpen] = useState(false);
    const flat = props.groups
      ? props.groups.flatMap((g) => g.options)
      : (props.options ?? []);
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

afterEach(() => {
  server.resetHandlers();
  localStorage.removeItem("auth_current_clinic:v1");
});

// ─────────────────────────────────────────────────────────────
// Issue #51: 新規飼主インラインフォーム (既存飼主/新規飼主切替)
// ─────────────────────────────────────────────────────────────

describe("ReservationFormModal — 新規飼主モード (Issue #51)", () => {
  it("デフォルト（既存飼主）モードで患者検索UIが表示される", () => {
    server.use(...silentApiHandlers);

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    // 既存飼主モードでは患者検索ラベルが表示される
    expect(screen.getByText("患者検索")).toBeInTheDocument();
    // 新規飼主フォームは表示されない
    expect(screen.queryByTestId("new-owner-name")).not.toBeInTheDocument();
  });

  it("既存飼主モードで患者と予約区分が未選択なら保存しない", async () => {
    server.use(...silentApiHandlers);
    const onSave = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={onSave}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    await user.click(screen.getByRole("button", { name: "予約を確定" }));

    expect(await screen.findByText("患者を選択してください")).toBeInTheDocument();
    expect(screen.getByText("予約区分を選択してください")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("「新規飼主」ボタンをクリックすると4項目の入力フォームが表示される", () => {
    server.use(...silentApiHandlers);

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    const newOwnerBtn = screen.getByTestId("mode-new");
    fireEvent.click(newOwnerBtn);

    // 新規飼主フォームの4項目が表示される
    expect(screen.getByTestId("new-owner-name")).toBeInTheDocument();
    expect(screen.getByTestId("new-owner-phone")).toBeInTheDocument();
    expect(screen.getByTestId("new-owner-pet-name")).toBeInTheDocument();
    expect(screen.getByTestId("new-owner-chief-complaint")).toBeInTheDocument();
    // 患者検索は非表示になる
    expect(screen.queryByText("患者検索")).not.toBeInTheDocument();
  });

  it("電話番号が10桁未満の場合BE整合フォーマットエラーが表示される", async () => {
    server.use(...silentApiHandlers);
    const user = userEvent.setup({ delay: null });

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    await user.click(screen.getByTestId("mode-new"));
    fireEvent.change(screen.getByTestId("new-owner-name"), { target: { value: "山田太郎" } });
    // 数字が10桁未満 — Backend contractに通らない形式
    fireEvent.change(screen.getByTestId("new-owner-phone"), { target: { value: "1234-5678" } });

    fireEvent.click(screen.getByRole("button", { name: "予約を確定" }));

    await waitFor(() => {
      expect(
        screen.getByText("電話番号の形式が正しくありません（例：090-1234-5678 または +81 90 1234 5678）")
      ).toBeInTheDocument();
    });
  });

  it("BUG-020: 電話番号を正しい形式へ直すと電話エラーだけが消える", async () => {
    server.use(...silentApiHandlers);
    const onSave = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={onSave}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    await user.click(screen.getByTestId("mode-new"));
    fireEvent.change(screen.getByTestId("new-owner-name"), { target: { value: "合成飼主" } });
    fireEvent.change(screen.getByTestId("new-owner-phone"), { target: { value: "1234-5678" } });
    await user.click(screen.getByRole("button", { name: "予約を確定" }));

    const phoneError =
      "電話番号の形式が正しくありません（例：090-1234-5678 または +81 90 1234 5678）";
    const phoneErrorElement = await screen.findByText(phoneError);
    expect(phoneErrorElement).toBeVisible();
    expect(phoneErrorElement).toHaveAttribute("role", "alert");
    expect(screen.getByText("ペット名を入力してください")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();

    fireEvent.change(screen.getByTestId("new-owner-phone"), {
      target: { value: "090-0000" },
    });
    expect(screen.getByText(phoneError)).toBeVisible();
    expect(onSave).not.toHaveBeenCalled();

    fireEvent.change(screen.getByTestId("new-owner-phone"), {
      target: { value: "090-0000-0000" },
    });

    await waitFor(() => {
      expect(screen.queryByText(phoneError)).not.toBeInTheDocument();
    });
    expect(screen.getByText("ペット名を入力してください")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  // Radix Select x2 + MSW + waitFor x3 → Docker jsdom で累計 5 秒超えが稀に発生するため 15s に設定
  it("全フィールド入力後に onSave が newOwnerData を含む引数で呼ばれる", async () => {
    server.use(
      http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
      http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
      http.get("/api/v1/masters/staffs", () => HttpResponse.json([])),
      http.get("/api/v1/shifts/on-duty-staffs", () => HttpResponse.json([])),
      http.get("/api/v1/reservations/available-times", () => HttpResponse.json([])),
      // 動物種: 犬 1件
      http.get("/api/v1/masters/animal-species", () =>
        HttpResponse.json([{ id: 1, name: "犬", is_active: true }])
      ),
      // 予約区分: 一般診療 1件（group なし → "その他" グループ）
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          { id: 5, name: "一般診療", color: "#000000", is_active: true, group_id: null, group: null },
        ])
      )
    );

    const onSave = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={onSave}
        initialData={null}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() }
    );

    // 新規飼主モードに切り替え（act() ラップ必要なため user.click を維持）
    await user.click(screen.getByTestId("mode-new"));

    // テキストフィールドを入力（fireEvent.change で1イベント完結、タイムアウト防止）
    fireEvent.change(screen.getByTestId("new-owner-name"), { target: { value: "山田太郎" } });
    fireEvent.change(screen.getByTestId("new-owner-phone"), { target: { value: "+81 90 1234 5678" } });
    fireEvent.change(screen.getByTestId("new-owner-pet-name"), { target: { value: "ポチ" } });
    fireEvent.change(screen.getByTestId("new-owner-chief-complaint"), { target: { value: "食欲不振" } });

    // SearchableSelect は種取得中 disabled。enabled になるまで待ってから開く（disabled click は no-op）
    const speciesTrigger = screen.getByTestId("new-owner-species");
    await waitFor(() => {
      expect(speciesTrigger).toBeEnabled();
      expect(speciesTrigger).not.toHaveTextContent("読み込み中");
    });
    await user.click(speciesTrigger);
    await user.click(await screen.findByRole("option", { name: "犬" }));
    await waitFor(() => {
      expect(speciesTrigger).toHaveTextContent("犬");
      expect(screen.queryByRole("option", { name: "犬" })).not.toBeInTheDocument();
    });

    // 予約区分: 種リストが閉じた後にサブダイアログを開きカードで選択(id 5 = 一般診療)
    await user.click(screen.getByTestId("res-type-trigger"));
    await user.click(await screen.findByTestId("res-type-card-5"));
    await waitFor(() => {
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("一般診療");
    });

    // 保存を実行
    await user.click(screen.getByRole("button", { name: "予約を確定" }));

    await waitFor(() => {
      expect(onSave).toHaveBeenCalledOnce();
    });

    const [, , newOwnerArg] = onSave.mock.calls[0];
    expect(newOwnerArg).toMatchObject({
      ownerName: "山田太郎",
      phone: "+81 90 1234 5678",
      petName: "ポチ",
      chiefComplaint: "食欲不振",
      animalSpeciesId: 1,
    });
  }, 15000);
});

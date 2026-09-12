import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { ReservationFormModal } from "./ReservationFormModal";
import { createWrapper, noop } from "./ReservationFormModal.test-helpers";

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
    const label =
      selected?.label ??
      (props.value ? (props.fallbackLabel ?? props.placeholder) : props.placeholder);
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
          {label}
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

describe("ReservationFormModal — 担当者候補", () => {
  it("対応可能コースを持つスタッフだけを担当者候補に残す（肯定形 capability）", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
      http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
      http.get("/api/v1/masters/animal-species", () => HttpResponse.json([])),
      http.get("/api/v1/masters/staffs", () =>
        HttpResponse.json([
          {
            id: 10,
            name: "非対応スタッフ",
            is_active: true,
            clinic_assignments: [{ clinic_id: 1, is_main: true }],
          },
          {
            id: 11,
            name: "対応スタッフ",
            is_active: true,
            clinic_assignments: [{ clinic_id: 1, is_main: true }],
          },
        ]),
      ),
      http.get("/api/v1/shifts/on-duty-staffs", () =>
        HttpResponse.json([
          { id: 10, name: "非対応スタッフ" },
          { id: 11, name: "対応スタッフ" },
        ]),
      ),
      http.get("/api/v1/clinics/1/reservation-staffs", () =>
        HttpResponse.json([
          {
            id: 10,
            name: "非対応スタッフ",
            is_active: true,
            capable_courses: [],
          },
          {
            id: 11,
            name: "対応スタッフ",
            is_active: true,
            capable_courses: [{ id: 5, name: "トリミング" }],
          },
        ]),
      ),
      http.get("/api/v1/masters/reservation-types/5/unavailable-times", () =>
        HttpResponse.json({ data: [] }),
      ),
      http.get("/api/v1/reservations/available-times", () =>
        HttpResponse.json([
          { start_time: "0945", end_time: "1045" },
          { start_time: "1230", end_time: "1330" },
        ]),
      ),
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          {
            id: 5,
            name: "トリミング",
            color: "#111111",
            is_active: true,
            duration_minutes: 60,
            sort_order: 1,
            is_internal: false,
            category: "trimming",
            group_id: null,
            group: null,
          },
        ]),
      ),
    );

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
      { wrapper: createWrapper() },
    );

    await user.click(screen.getByTestId("res-type-trigger"));
    // 予約区分サブダイアログでカード選択(id 5 = トリミング)
    fireEvent.click(await screen.findByTestId("res-type-card-5"));

    await user.click(screen.getByTestId("res-staff-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "対応スタッフ" })).toBeInTheDocument();
    });
    expect(screen.queryByRole("option", { name: "非対応スタッフ" })).not.toBeInTheDocument();
  }, 15000);

  const ORPHAN_STAFF_HANDLERS = [
    http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
    http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
    http.get("/api/v1/masters/animal-species", () =>
      HttpResponse.json([{ id: 1, name: "犬", is_active: true }]),
    ),
    http.get("/api/v1/masters/staffs", () =>
      HttpResponse.json([
        {
          id: 10,
          name: "三井",
          is_active: true,
          clinic_assignments: [{ clinic_id: 1, is_main: true }],
        },
        {
          id: 11,
          name: "鈴木",
          is_active: true,
          clinic_assignments: [{ clinic_id: 1, is_main: true }],
        },
      ]),
    ),
    http.get("/api/v1/shifts/on-duty-staffs", () =>
      HttpResponse.json([
        { id: 10, name: "三井" },
        { id: 11, name: "鈴木" },
      ]),
    ),
    http.get("/api/v1/clinics/1/reservation-staffs", () =>
      HttpResponse.json([
        { id: 10, name: "三井", is_active: true, capable_courses: [] },
        {
          id: 11,
          name: "鈴木",
          is_active: true,
          capable_courses: [
            { id: 5, name: "診察" },
            { id: 6, name: "トリミング" },
          ],
        },
      ]),
    ),
    http.get("/api/v1/masters/reservation-types/5/unavailable-times", () =>
      HttpResponse.json({ data: [] }),
    ),
    http.get("/api/v1/masters/reservation-types/6/unavailable-times", () =>
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
        {
          id: 6,
          name: "トリミング",
          color: "#222222",
          is_active: true,
          duration_minutes: 60,
          sort_order: 2,
          is_internal: false,
          category: "trimming",
          group_id: null,
          group: null,
        },
      ]),
    ),
  ];

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

  it("A1: 有効な担当選択は名前表示と同一 id の送信値になる", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(...ORPHAN_STAFF_HANDLERS);

    const onSave = vi.fn();
    const user = userEvent.setup({ delay: null });
    const start = new Date();
    start.setDate(start.getDate() + 1);
    start.setHours(10, 0, 0, 0);
    const end = new Date(start);
    end.setHours(11, 0, 0, 0);

    render(
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

    await fillNewOwnerBasics(user);
    await user.click(screen.getByTestId("res-type-trigger"));
    fireEvent.click(await screen.findByTestId("res-type-card-5"));
    await waitFor(() => {
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("診察");
    });

    await user.click(screen.getByTestId("res-staff-trigger"));
    await user.click(await screen.findByRole("option", { name: "鈴木" }));
    await waitFor(() => {
      expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("鈴木");
    });

    await user.click(screen.getByRole("button", { name: "予約を確定" }));
    await waitFor(() => {
      expect(onSave).toHaveBeenCalledOnce();
    });
    expect(onSave.mock.calls[0][0]).toMatchObject({ doctor: "11", type: "5" });
  }, 20000);

  it("A2: 区分変更で担当が候補外になったら理由表示し、解除まで送信しない", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(...ORPHAN_STAFF_HANDLERS);

    const onSave = vi.fn();
    const user = userEvent.setup({ delay: null });
    const start = new Date();
    start.setDate(start.getDate() + 1);
    start.setHours(10, 0, 0, 0);
    const end = new Date(start);
    end.setHours(11, 0, 0, 0);

    render(
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

    await fillNewOwnerBasics(user);

    // type unset: select Mitsui while still on-duty candidate
    await user.click(screen.getByTestId("res-staff-trigger"));
    await user.click(await screen.findByRole("option", { name: "三井" }));
    await waitFor(() => {
      expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("三井");
    });

    await user.click(screen.getByTestId("res-type-trigger"));
    fireEvent.click(await screen.findByTestId("res-type-card-5"));
    await waitFor(() => {
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("診察");
    });

    // Keep name (not silent placeholder unset) and show orphan reason
    await waitFor(() => {
      expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("三井");
      expect(
        screen.getByText(
          "この条件では指定できない担当者です。解除するか、対応可能な担当者を選び直してください。",
        ),
      ).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "予約を確定" }));
    await waitFor(() => {
      expect(
        screen.getByText(
          "この条件では指定できない担当者です。解除するか、対応可能な担当者を選び直してください。",
        ),
      ).toBeInTheDocument();
    });
    expect(onSave).not.toHaveBeenCalled();

    await user.click(screen.getByTestId("res-staff-clear"));
    await waitFor(() => {
      expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("選択してください");
      expect(
        screen.queryByText(
          "この条件では指定できない担当者です。解除するか、対応可能な担当者を選び直してください。",
        ),
      ).not.toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "予約を確定" }));
    await waitFor(() => {
      expect(onSave).toHaveBeenCalledOnce();
    });
    expect(onSave.mock.calls[0][0]).toMatchObject({ doctor: "", type: "5" });
  }, 20000);

  it("A2: 候補 metadata 取得中は orphan 扱いにせず選択名を保持する", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    let releaseStaffs: (() => void) | undefined;
    const staffsGate = new Promise<void>((resolve) => {
      releaseStaffs = resolve;
    });

    server.use(
      http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
      http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
      http.get("/api/v1/masters/animal-species", () =>
        HttpResponse.json([{ id: 1, name: "犬", is_active: true }]),
      ),
      http.get("/api/v1/masters/staffs", () =>
        HttpResponse.json([
          {
            id: 10,
            name: "三井",
            is_active: true,
            clinic_assignments: [{ clinic_id: 1, is_main: true }],
          },
          {
            id: 11,
            name: "鈴木",
            is_active: true,
            clinic_assignments: [{ clinic_id: 1, is_main: true }],
          },
        ]),
      ),
      http.get("/api/v1/shifts/on-duty-staffs", () =>
        HttpResponse.json([
          { id: 10, name: "三井" },
          { id: 11, name: "鈴木" },
        ]),
      ),
      http.get("/api/v1/clinics/1/reservation-staffs", async () => {
        await staffsGate;
        return HttpResponse.json([
          { id: 10, name: "三井", is_active: true, capable_courses: [] },
          {
            id: 11,
            name: "鈴木",
            is_active: true,
            capable_courses: [{ id: 5, name: "診察" }],
          },
        ]);
      }),
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
    );

    const user = userEvent.setup({ delay: null });
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
        initialData={{ start, end }}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() },
    );

    await user.click(screen.getByTestId("res-staff-trigger"));
    await user.click(await screen.findByRole("option", { name: "三井" }));
    await waitFor(() => {
      expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("三井");
    });

    await user.click(screen.getByTestId("res-type-trigger"));
    fireEvent.click(await screen.findByTestId("res-type-card-5"));

    // Pending metadata: keep label, do not show confirmed-orphan reason yet
    await waitFor(() => {
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("診察");
    });
    expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("三井");
    expect(
      screen.queryByText(
        "この条件では指定できない担当者です。解除するか、対応可能な担当者を選び直してください。",
      ),
    ).not.toBeInTheDocument();

    releaseStaffs?.();
    await waitFor(() => {
      expect(
        screen.getByText(
          "この条件では指定できない担当者です。解除するか、対応可能な担当者を選び直してください。",
        ),
      ).toBeInTheDocument();
    });
    expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("三井");
  }, 20000);

  it("A2: reservation-staffs 取得失敗では orphan クリアせず選択を保持する", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
      http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
      http.get("/api/v1/masters/animal-species", () =>
        HttpResponse.json([{ id: 1, name: "犬", is_active: true }]),
      ),
      http.get("/api/v1/masters/staffs", () =>
        HttpResponse.json([
          {
            id: 10,
            name: "三井",
            is_active: true,
            clinic_assignments: [{ clinic_id: 1, is_main: true }],
          },
          {
            id: 11,
            name: "鈴木",
            is_active: true,
            clinic_assignments: [{ clinic_id: 1, is_main: true }],
          },
        ]),
      ),
      http.get("/api/v1/shifts/on-duty-staffs", () =>
        HttpResponse.json([
          { id: 10, name: "三井" },
          { id: 11, name: "鈴木" },
        ]),
      ),
      http.get("/api/v1/clinics/1/reservation-staffs", () =>
        HttpResponse.json({ message: "boom" }, { status: 500 }),
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
    );

    const user = userEvent.setup({ delay: null });
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
        initialData={{ start, end }}
        canCreate={true}
        canEdit={false}
      />,
      { wrapper: createWrapper() },
    );

    await user.click(screen.getByTestId("res-staff-trigger"));
    await user.click(await screen.findByRole("option", { name: "三井" }));
    await waitFor(() => {
      expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("三井");
    });

    await user.click(screen.getByTestId("res-type-trigger"));
    fireEvent.click(await screen.findByTestId("res-type-card-5"));
    await waitFor(() => {
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("診察");
    });

    expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("三井");
    expect(
      screen.queryByText(
        "この条件では指定できない担当者です。解除するか、対応可能な担当者を選び直してください。",
      ),
    ).not.toBeInTheDocument();
    expect(screen.queryByTestId("res-staff-clear")).not.toBeInTheDocument();
  }, 20000);

  it("A3: 編集初期表示は保存済み担当を書き換えず名前を表示する", async () => {
    localStorage.setItem("auth_current_clinic:v1", "1");
    server.use(
      ...ORPHAN_STAFF_HANDLERS,
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
          doctor: "10",
          visitType: "revisit",
          status: "confirmed",
        }}
        canCreate={false}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    await waitFor(() => {
      expect(screen.getByTestId("res-staff-trigger")).toHaveTextContent("三井");
      expect(screen.getByTestId("res-type-trigger")).toHaveTextContent("診察");
    });
    // Display convenience must not wipe or replace the saved doctor id.
    expect(screen.getByTestId("res-staff-trigger")).not.toHaveTextContent("選択してください");
    expect(
      screen.getByText(
        "この条件では指定できない担当者です。解除するか、対応可能な担当者を選び直してください。",
      ),
    ).toBeInTheDocument();
  }, 20000);
});

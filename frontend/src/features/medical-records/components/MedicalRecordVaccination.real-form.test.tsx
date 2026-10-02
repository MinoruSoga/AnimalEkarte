import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import { axios } from "@/lib/axios";
import { MedicalRecordVaccination } from "./MedicalRecordVaccination";

// BUG-MR-VACCINE-FORM-NESTED (EMR-69): real-component path. The sibling
// MedicalRecordVaccination.test.tsx mocks ./VaccinationForm, so it cannot prove
// the REAL form stays un-nested and the REAL SubmitButton dispatches the
// vaccination action. This file keeps VaccinationForm, VaccinationHistory,
// SubmitButton, useCreateVaccination and useGetPetVaccinations real; only leaf
// input controls and the axios boundary are mocked.

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllVaccinesMaster: () => ({
    data: [
      { id: "7", name: "混合ワクチン", isActive: true },
      { id: "8", name: "狂犬病ワクチン", isActive: true },
    ],
  }),
}));

vi.mock("@/components/ui/searchable-select", () => ({
  SearchableSelect: ({
    value,
    onValueChange,
    options,
  }: {
    value: string;
    onValueChange: (value: string) => void;
    options?: { value: string; label: string }[];
  }) => (
    <select
      aria-label="予防接種名"
      value={value}
      onChange={(event) => onValueChange(event.target.value)}
    >
      <option value="">ワクチンを選択</option>
      {(options ?? []).map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  ),
}));

// VaccinationForm imports "@/components/shared/DatePicker" while
// VaccinationHistory imports the deep "@/components/shared/DatePicker/DatePicker";
// both resolve to the same component, so mock both module paths.
const { DatePickerStub } = vi.hoisted(() => {
  const Stub = ({
    value,
    onChange,
    placeholder,
  }: {
    value?: string;
    onChange: (value: string) => void;
    placeholder?: string;
  }) => (
    <input
      aria-label={placeholder ?? "予防接種日"}
      value={value}
      onChange={(event) => onChange(event.target.value)}
    />
  );
  return { DatePickerStub: Stub };
});
vi.mock("@/components/shared/DatePicker", () => ({ DatePicker: DatePickerStub }));
vi.mock("@/components/shared/DatePicker/DatePicker", () => ({
  DatePicker: DatePickerStub,
}));

// Keep the pure helpers (calculateNextDate / resolveScheduleTypeAfterManualDate)
// that use-medical-record-vaccination-form imports from the same module; replace
// only the interactive field component.
vi.mock("@/components/shared/NextScheduleField", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/components/shared/NextScheduleField")>();
  return {
    ...actual,
    NextScheduleField: ({
      scheduleType,
      nextDate,
      onScheduleTypeChange,
      onNextDateChange,
    }: {
      scheduleType: string;
      nextDate: string;
      onScheduleTypeChange: (value: string) => void;
      onNextDateChange: (value: string) => void;
    }) => (
      <div>
        <input
          aria-label="次回予定種別"
          value={scheduleType}
          onChange={(event) => onScheduleTypeChange(event.target.value)}
        />
        <input
          aria-label="次回接種予定日"
          value={nextDate}
          onChange={(event) => onNextDateChange(event.target.value)}
        />
      </div>
    ),
  };
});

vi.mock("@/components/shared/MasterLink", () => ({
  MasterLink: () => null,
}));

// BUG-501 規約: 実施日は JST 当日デフォルト。他の JST helpers (formatJSTDate /
// toJSTWallDate) は履歴 transform が使うので実装のまま残す。
vi.mock("@/lib/jst-date", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/lib/jst-date")>()),
  todayJSTISO: () => "2026-08-29",
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Simulated server-side vaccination store. axios.post echoes the posted body
// into the store so the invalidation-driven refetch (axios.get) returns the
// actual per-registration payload — this is what makes the B4 assertion
// end-to-end instead of a manual rerender, and what lets the EMR-105 same-day
// test observe whether sequential registrations keep their own fields.
const historyItems: Record<string, unknown>[] = [];
let nextVaccinationId = 12;

const vaccineNamesById: Record<number, string> = {
  7: "新混合ワクチン",
  8: "狂犬病ワクチン",
};

const seedVaccination = {
  id: 11,
  pet_id: 1,
  medical_record_id: 98,
  vaccine_id: 5,
  date: "2026-08-01T00:00:00+09:00",
  next_date: null,
  vaccine: { name: "既存ワクチン" },
};

const mockGet = vi.mocked(axios.get);
const mockPost = vi.mocked(axios.post);

beforeEach(() => {
  historyItems.length = 0;
  historyItems.push(seedVaccination);
  nextVaccinationId = 12;
  mockGet.mockReset();
  mockPost.mockReset();
  mockGet.mockImplementation(() =>
    Promise.resolve({ data: { data: historyItems.map((item) => ({ ...item })) } }),
  );
  mockPost.mockImplementation((_url: string, body: unknown) => {
    const posted = (body ?? {}) as Record<string, unknown>;
    const vaccineId = Number(posted.vaccine_id);
    const rec = {
      ...posted,
      id: nextVaccinationId++,
      vaccine: { name: vaccineNamesById[vaccineId] ?? `ワクチン(ID:${vaccineId})` },
    };
    historyItems.push(rec);
    return Promise.resolve({ data: rec });
  });
  vi.mocked(toast.success).mockClear();
});

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
}

// EMR-212: production always renders the section inside the page-level
// medical-record <form action> (カルテ保存). Reproduce that nesting context.
function renderPanel(outerAction: (payload: FormData) => void = () => {}) {
  const queryClient = makeQueryClient();
  const view = render(
    <QueryClientProvider client={queryClient}>
      <form action={outerAction}>
        <MedicalRecordVaccination petId="1" medicalRecordId="99" />
      </form>
    </QueryClientProvider>,
  );
  return { queryClient, ...view };
}

async function openAddForm() {
  fireEvent.click(await screen.findByRole("button", { name: "記録を追加" }));
  // Real VaccinationForm is open once its real SubmitButton label appears.
  await screen.findByRole("button", { name: "接種記録を追加" });
}

async function submitAddForm() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: "接種記録を追加" }));
}

describe("MedicalRecordVaccination real VaccinationForm — nested-form regression (EMR-69)", () => {
  it("追加フォーム表示中も合成 DOM 内の form 要素は外側のカルテ保存 1 つのみ", async () => {
    const { container } = renderPanel();

    expect(container.querySelectorAll("form")).toHaveLength(1);
    await openAddForm();

    // 実コンポーネントが開いていることを実フィールドで確認してから構造を検査
    expect(screen.getByLabelText("予防接種名")).toBeInTheDocument();
    expect(screen.getByLabelText("予防接種日")).toBeInTheDocument();
    expect(container.querySelectorAll("form")).toHaveLength(1);
  });

  it("実 SubmitButton がワクチン登録 action を発火し、外側のカルテ保存 action は呼ばれない", async () => {
    const outerAction = vi.fn();
    renderPanel(outerAction);
    await openAddForm();

    fireEvent.change(screen.getByLabelText("予防接種名"), { target: { value: "7" } });
    fireEvent.change(screen.getByLabelText("予防接種日"), { target: { value: "2026-07-20" } });
    await submitAddForm();

    await waitFor(() => {
      expect(mockPost).toHaveBeenCalledWith(
        "/v1/vaccinations",
        expect.objectContaining({
          pet_id: 1,
          medical_record_id: 99,
          vaccine_id: 7,
          date: "2026-07-20",
        }),
      );
    });
    expect(outerAction).not.toHaveBeenCalled();
  });

  it("ワクチン未選択・接種日クリアでは実 FormFieldError が出て POST しない", async () => {
    renderPanel();
    await openAddForm();

    // 実施日は JST 当日デフォルトのため未入力検証は明示クリアが必要
    fireEvent.change(screen.getByLabelText("予防接種日"), { target: { value: "" } });
    await submitAddForm();

    const alerts = await screen.findAllByRole("alert");
    expect(alerts.map((alert) => alert.textContent)).toEqual(
      expect.arrayContaining(["ワクチン種別を選択してください", "接種日を入力してください"]),
    );
    expect(mockPost).not.toHaveBeenCalled();
  });

  it("登録成功で vaccinations 履歴 query が無効化・再取得され一覧に新規行が描画される", async () => {
    renderPanel();
    // 初期取得が終わり既存履歴が見えるまで待つ
    expect(await screen.findAllByText("既存ワクチン")).not.toHaveLength(0);

    await openAddForm();
    fireEvent.change(screen.getByLabelText("予防接種名"), { target: { value: "7" } });
    fireEvent.change(screen.getByLabelText("予防接種日"), { target: { value: "2026-07-20" } });
    await submitAddForm();

    // create 成功 → useCreateVaccination.onSuccess が ["vaccinations"] を invalidate
    // → ["vaccinations","pet","1"] が prefix 一致で再取得 → 新規レコードが可視リストに出る
    await waitFor(() => {
      expect(mockPost).toHaveBeenCalledWith(
        "/v1/vaccinations",
        expect.objectContaining({ vaccine_id: 7, date: "2026-07-20" }),
      );
    });
    await waitFor(() => {
      expect(mockGet).toHaveBeenCalledTimes(2);
    });
    expect(await screen.findAllByText("新混合ワクチン")).not.toHaveLength(0);
    expect(await screen.findAllByText("既存ワクチン")).not.toHaveLength(0);
    expect(toast.success).toHaveBeenCalledWith("接種記録を追加しました");

    // resetForm: 追加フォームは閉じて一覧表示に戻る
    expect(await screen.findByRole("button", { name: "記録を追加" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "接種記録を追加" })).not.toBeInTheDocument();
  });
});

// EMR-105 (S29 step 2-5, in-chart entry path): 保存済みカルテ内フォームから
// 同日に複数ワクチンを順次登録しても、各 POST が独立し、再取得された履歴行の
// lot / 次回予定 / ワクチン種別 / medical_record 紐付けが混ざらないこと。
// 単独フォーム経路は 2026-09-23 実保存証拠済み（reports/uat-2026-09-23/S29）のため
// ここではカルテ内経路のみを実フォームで検証する。
describe("MedicalRecordVaccination EMR-105 same-day multi-vaccine (in-chart path)", () => {
  it("保存済みカルテから同日2件を順次登録し、各 POST と再表示行が混ざらない", async () => {
    renderPanel();
    // 既存履歴（前日の seed 行）が見えるまで初期取得を待つ
    expect(await screen.findAllByText("既存ワクチン")).not.toHaveLength(0);

    // ── 1件目: vaccine 7 / LOT-A / 次回 2026-09-26 ──
    // 接種日はカルテ内フォームの JST 当日 default（2026-08-29）をそのまま使う = 同日接種。
    await openAddForm();
    fireEvent.change(screen.getByLabelText("予防接種名"), { target: { value: "7" } });
    const lot1Input = screen.getByText("LOT1").parentElement?.querySelector("input");
    expect(lot1Input).not.toBeNull();
    fireEvent.change(lot1Input as Element, { target: { value: "LOT-A" } });
    fireEvent.change(screen.getByLabelText("次回接種予定日"), {
      target: { value: "2026-09-26" },
    });
    await submitAddForm();
    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledTimes(1);
    });
    // 成功後フォームは閉じて一覧に戻る（resetForm）— 2件目は開き直す。
    expect(await screen.findByRole("button", { name: "記録を追加" })).toBeInTheDocument();

    // ── 2件目: vaccine 8 / LOT-B / 次回 2026-10-10、同日（当日 default）──
    await openAddForm();
    fireEvent.change(screen.getByLabelText("予防接種名"), { target: { value: "8" } });
    const lot1Input2 = screen.getByText("LOT1").parentElement?.querySelector("input");
    expect(lot1Input2).not.toBeNull();
    fireEvent.change(lot1Input2 as Element, { target: { value: "LOT-B" } });
    fireEvent.change(screen.getByLabelText("次回接種予定日"), {
      target: { value: "2026-10-10" },
    });
    await submitAddForm();
    await waitFor(() => {
      expect(mockPost).toHaveBeenCalledTimes(2);
    });

    // 各 POST は同じ pet_id / medical_record_id / date を持つが、
    // vaccine_id・lot1・next_date は登録ごとの値でなければならない。
    expect(mockPost).toHaveBeenNthCalledWith(
      1,
      "/v1/vaccinations",
      expect.objectContaining({
        pet_id: 1,
        medical_record_id: 99,
        vaccine_id: 7,
        date: "2026-08-29",
        lot1: "LOT-A",
        next_date: "2026-09-26",
      }),
    );
    expect(mockPost).toHaveBeenNthCalledWith(
      2,
      "/v1/vaccinations",
      expect.objectContaining({
        pet_id: 1,
        medical_record_id: 99,
        vaccine_id: 8,
        date: "2026-08-29",
        lot1: "LOT-B",
        next_date: "2026-10-10",
      }),
    );

    // 再表示（invalidation 再取得）: seed + 2件の同日行が各自行として描画される。
    await waitFor(() => {
      expect(mockGet).toHaveBeenCalledTimes(3);
    });
    expect(await screen.findAllByText("既存ワクチン")).not.toHaveLength(0);
    expect(await screen.findAllByText("新混合ワクチン")).not.toHaveLength(0);
    expect(await screen.findAllByText("狂犬病ワクチン")).not.toHaveLength(0);
    // 次予定列も行ごとに独立（同日だからといって共有・上書きされない）。
    expect(await screen.findAllByText("26/9/26")).not.toHaveLength(0);
    expect(await screen.findAllByText("26/10/10")).not.toHaveLength(0);
  });
});

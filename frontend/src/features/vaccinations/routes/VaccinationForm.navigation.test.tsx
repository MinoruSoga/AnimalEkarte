import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import {
  createMemoryRouter,
  Outlet,
  RouterProvider,
  useLocation,
  type InitialEntry,
} from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { server } from "@/testing/mocks/node";

import { VaccinationForm } from "./VaccinationForm";
import { VaccinationList } from "./VaccinationList";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  })),
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { clinics: [{ clinicId: "clinic-1" }] },
    currentClinicId: "clinic-1",
  }),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: vi.fn(),
    markClean: vi.fn(),
  }),
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("@/components/shared/PatientInfoCard", () => ({
  PatientInfoCard: () => null,
  formatPatientPetDetails: () => "",
}));

vi.mock("@/components/shared/ConfirmDialog/ConfirmDialog", () => ({
  ConfirmDialog: () => null,
}));

const PET_ID = "14";
const TARGET_ID = "25";
const TARGET_DETAIL_URL = `/vaccinations/${TARGET_ID}`;
const LIST_URL = `/vaccinations?petId=${PET_ID}&page=2`;
const RECEPTION_STATE = { from: "/?date=2026-10-15" };

interface WireVaccination {
  id: number;
  clinic_id: number;
  pet_id?: number;
  vaccine_id: number;
  date: string;
  doctor_id?: number;
  next_date?: string;
  next_schedule_type?: string;
  supplemental: string;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  remarks: string;
  version: number;
  created_at: string;
  updated_at: string;
  pet?: { id: number; name: string; owner: { name: string } };
  vaccine?: { id: number; name: string };
  doctor?: { id: number; name: string };
}

let vaccinationStore: WireVaccination[];
let medicalRecordWrites: number;

/** 合成 wire フィクスチャ。overrides で検証対象フィールドだけを差し替える。 */
function wireVaccination(index: number, overrides: Partial<WireVaccination> = {}): WireVaccination {
  return {
    id: index,
    clinic_id: 1,
    pet_id: 14,
    vaccine_id: 1,
    date: "2026-01-10",
    next_date: "2027-01-10",
    supplemental: "",
    lot1: "",
    lot2: "",
    lot3: "",
    lot4: "",
    remarks: "",
    version: 1,
    created_at: "2026-01-10T00:00:00Z",
    updated_at: "2026-01-10T00:00:00Z",
    pet: { id: 14, name: "ポチ", owner: { name: "山田太郎" } },
    vaccine: { id: 1, name: "混合ワクチン" },
    doctor: { id: 3, name: "佐藤医師" },
    ...overrides,
  };
}

/** page2 に対象行が来る 26 行（他ペット行も混ぜ、pet_id 絞込の欠落を検出する）。 */
function seedStore(): WireVaccination[] {
  const firstPageRows = Array.from({ length: 24 }, (_v, i) => {
    const day = String(i + 1).padStart(2, "0");
    return wireVaccination(i + 1, { date: `2026-01-${day}`, next_date: `2026-02-${day}` });
  });
  return [
    ...firstPageRows,
    wireVaccination(Number(TARGET_ID), { date: "2026-01-25", next_date: undefined }),
    wireVaccination(900, {
      pet_id: 15,
      date: "2026-01-25",
      pet: { id: 15, name: "ミケ", owner: { name: "鈴木花子" } },
    }),
  ];
}

/** MSW 合成ハンドラ。予防接種の読取/PATCH と表示用フィクスチャを合成し、medical-records への書込みは counter で捕捉する（実 backend ではない）。 */
function useApiHandlers() {
  server.use(
    http.get("*/v1/vaccinations", ({ request }) => {
      const url = new URL(request.url);
      const petId = url.searchParams.get("pet_id");
      const rows = petId
        ? vaccinationStore.filter((row) => String(row.pet_id) === petId)
        : vaccinationStore;
      return HttpResponse.json({ data: rows });
    }),
    http.get(`*/v1/vaccinations/${TARGET_ID}`, () => {
      const row = vaccinationStore.find((r) => String(r.id) === TARGET_ID);
      return row
        ? HttpResponse.json(row)
        : HttpResponse.json({ message: "not found" }, { status: 404 });
    }),
    http.patch(`*/v1/vaccinations/${TARGET_ID}`, async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>;
      const row = vaccinationStore.find((r) => String(r.id) === TARGET_ID);
      if (!row) return HttpResponse.json({ message: "not found" }, { status: 404 });
      const updated: WireVaccination = {
        ...row,
        remarks: body.remarks !== undefined ? String(body.remarks) : row.remarks,
        next_date: typeof body.next_date === "string" ? body.next_date : row.next_date,
        version: row.version + 1,
      };
      vaccinationStore = vaccinationStore.map((r) => (String(r.id) === TARGET_ID ? updated : r));
      return HttpResponse.json(updated);
    }),
    http.get(`*/v1/pets/${PET_ID}`, () =>
      HttpResponse.json({
        id: 14,
        clinic_id: 1,
        owner_id: 9,
        name: "ポチ",
        status: "alive",
        owner: { name: "山田太郎" },
        animal_species: { name: "犬" },
      }),
    ),
    http.get("*/v1/masters/vaccines", () =>
      HttpResponse.json([{ id: 1, name: "混合ワクチン", is_active: true, interval: "1year" }]),
    ),
    http.all("*/v1/medical-records*", () => {
      medicalRecordWrites += 1;
      return HttpResponse.json({ message: "unexpected" }, { status: 500 });
    }),
  );
}

/** 現在の router URL/state を data-testid で露出し、遷移後の復元を検証する。 */
function LocationProbe() {
  const location = useLocation();
  return (
    <>
      <output data-testid="nav-location">{`${location.pathname}${location.search}`}</output>
      <output data-testid="nav-state">{JSON.stringify(location.state ?? null)}</output>
      <Outlet />
    </>
  );
}

/** テストごとにキャッシュを分離し retry を切る。 */
function newQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

/** 実 VaccinationList / VaccinationForm を memory router で通す。initialEntries は呼び出し側が履歴・state を指定する。 */
function renderFlow(initialEntries: InitialEntry[]) {
  const queryClient = newQueryClient();
  const router = createMemoryRouter(
    [
      {
        element: <LocationProbe />,
        children: [
          { path: "/", element: <div>受付</div> },
          { path: "/vaccinations", element: <VaccinationList /> },
          { path: "/vaccinations/:id", element: <VaccinationForm /> },
        ],
      },
    ],
    { initialEntries },
  );
  const view = render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return { router, view, queryClient };
}

/** probe が捉えた現在の pathname+search。 */
function navLocation() {
  return screen.getByTestId("nav-location").textContent;
}

/** probe が捉えた現在の location.state（JSON パース済み）。 */
function navState(): unknown {
  return JSON.parse(screen.getByTestId("nav-state").textContent ?? "null");
}

/** 一覧の実行リンクから対象予防接種の編集フォームへ遷移する。 */
async function openTargetFromList(user: ReturnType<typeof userEvent.setup>) {
  const link = await screen.findByRole("link", { name: new RegExp(`ID ${TARGET_ID}$`) });
  await user.click(link);
  await screen.findByRole("heading", { name: "予防接種編集" });
}

/** 指定した戻り系ボタンを押し、一覧 URL へ戻るまで待つ。 */
async function backToList(user: ReturnType<typeof userEvent.setup>, buttonName: string) {
  await user.click(await screen.findByRole("button", { name: buttonName }));
  await waitFor(() => expect(navLocation()).toBe(LIST_URL));
}

beforeEach(() => {
  vaccinationStore = seedStore();
  medicalRecordWrites = 0;
  useApiHandlers();
});

describe("VaccinationForm ナビゲーション state 保持（EMR-281）", () => {
  it.each(["戻る", "キャンセル"])(
    "ペット絞込一覧→詳細→%s で petId/page と listReturnState が維持される",
    async (buttonName) => {
      const user = userEvent.setup();
      renderFlow([
        { pathname: "/vaccinations", search: `?petId=${PET_ID}&page=2`, state: RECEPTION_STATE },
      ]);

      const link = await screen.findByRole("link", { name: new RegExp(`ID ${TARGET_ID}$`) });
      expect(navLocation()).toBe(LIST_URL);
      expect(screen.queryByRole("link", { name: /ミケ/ })).not.toBeInTheDocument();

      await user.click(link);
      await screen.findByRole("heading", { name: "予防接種編集" });
      expect(navLocation()).toBe(TARGET_DETAIL_URL);
      expect(navState()).toEqual({ from: LIST_URL, listReturnState: RECEPTION_STATE });

      await backToList(user, buttonName);
      expect(navState()).toEqual(RECEPTION_STATE);
      await screen.findByRole("link", { name: new RegExp(`ID ${TARGET_ID}$`) });

      expect(medicalRecordWrites).toBe(0);
    },
  );

  it("備考保存→一覧→再表示→state維持リロード→戻る でも URL/state/保存値が維持される", async () => {
    const user = userEvent.setup();
    const first = renderFlow([
      { pathname: "/vaccinations", search: `?petId=${PET_ID}&page=2`, state: RECEPTION_STATE },
    ]);

    await openTargetFromList(user);

    const remarks = await screen.findByLabelText("備考");
    await user.clear(remarks);
    await user.type(remarks, "EDIT_20261018");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(navLocation()).toBe(LIST_URL));
    expect(navState()).toEqual(RECEPTION_STATE);

    await openTargetFromList(user);
    await waitFor(() => expect(screen.getByLabelText("備考")).toHaveValue("EDIT_20261018"));

    const captured = first.router.state.location;
    expect(captured.state).toEqual({ from: LIST_URL, listReturnState: RECEPTION_STATE });
    const reloadEntry: InitialEntry = {
      pathname: captured.pathname,
      search: captured.search,
      hash: captured.hash,
      state: captured.state,
    };
    first.view.unmount();
    first.queryClient.clear();

    renderFlow([reloadEntry]);
    await screen.findByRole("heading", { name: "予防接種編集" });
    await waitFor(() => expect(screen.getByLabelText("備考")).toHaveValue("EDIT_20261018"));

    await backToList(user, "戻る");
    expect(navState()).toEqual(RECEPTION_STATE);
    expect(medicalRecordWrites).toBe(0);
  });

  it("state なしで詳細へ直接入った場合は安全に /vaccinations へフォールバックする", async () => {
    const user = userEvent.setup();
    renderFlow([{ pathname: TARGET_DETAIL_URL }]);

    await screen.findByRole("heading", { name: "予防接種編集" });
    await user.click(await screen.findByRole("button", { name: "戻る" }));

    await waitFor(() => expect(navLocation()).toBe("/vaccinations"));
    expect(medicalRecordWrites).toBe(0);
  });

  it("外部 URL の from は open redirect として拒否され /vaccinations へフォールバックする", async () => {
    const user = userEvent.setup();
    renderFlow([{ pathname: TARGET_DETAIL_URL, state: { from: "https://evil.example.com/x" } }]);

    await screen.findByRole("heading", { name: "予防接種編集" });
    await user.click(await screen.findByRole("button", { name: "戻る" }));

    await waitFor(() => expect(navLocation()).toBe("/vaccinations"));
  });
});

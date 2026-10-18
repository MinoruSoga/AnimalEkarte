import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, renderHook, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { server } from "@/testing/mocks/node";
import { queryKeys } from "@/lib/query-keys";
import { useGetAllConsultations, useGetAllProcedures } from "@/hooks/use-treatment-master";

import { TreatmentsTab } from "./TreatmentsTab";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }] },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));

const RECORD_ID = "10";
const CLINIC1 = "1";
const CLINIC2 = "2";

const CLINIC1_CONSULTATIONS = [{ id: 101, name: "診察C1", price: 1000, is_active: true }];
const CLINIC1_PROCEDURES = [{ id: 102, name: "処置C1", price: 2000, is_active: true }];
const CLINIC2_CONSULTATIONS = [{ id: 201, name: "診察C2", price: 1100, is_active: true }];
const CLINIC2_PROCEDURES = [{ id: 202, name: "処置C2", price: 2200, is_active: true }];

const EXPECTED_PRICE_BY_ID: Record<number, number> = {
  101: 1000,
  102: 2000,
  201: 1100,
  202: 2200,
};

const ALLOWED_REF_BY_CLINIC: Record<string, { consultation_id: number; procedure_id: number }> = {
  "1": { consultation_id: 101, procedure_id: 102 },
  "2": { consultation_id: 201, procedure_id: 202 },
};

interface PostCall {
  header: string | null;
  body: Record<string, unknown>;
}

interface CategoryFixture {
  category: "consultations" | "procedures";
  globalName: string;
  scopedName: string;
  refKey: "consultation_id" | "procedure_id";
  refId: number;
  price: string;
}

const CATEGORIES: CategoryFixture[] = [
  {
    category: "consultations",
    globalName: "診察C1",
    scopedName: "診察C2",
    refKey: "consultation_id",
    refId: 201,
    price: "¥1,100",
  },
  {
    category: "procedures",
    globalName: "処置C1",
    scopedName: "処置C2",
    refKey: "procedure_id",
    refId: 202,
    price: "¥2,200",
  },
];

let acceptedPosts: PostCall[];
let rejectedPosts: PostCall[];
let getHeadersByEndpoint: Record<"consultations" | "procedures", Array<string | null>>;
let clinic2GateByCategory: Partial<Record<"consultations" | "procedures", Promise<void>>>;

/** リクエストの X-Clinic-ID を返す。ヘッダ欠落時のみフィクスチャの clinic1 扱い。 */
function clinicOf(request: Request): string {
  return request.headers.get("X-Clinic-ID") ?? CLINIC1;
}

/**
 * MSW 合成ハンドラ。マスタ GET はヘッダのクリニックで 2 拠点フィクスチャを返し、
 * POST はヘッダと参照 ID の一致を合成判定して accepted/rejected に記録する
 * （実 DB・実テナント判定ではない）。
 */
function useBaseHandlers() {
  server.use(
    http.get(`*/v1/medical-records/${RECORD_ID}/treatments`, () => HttpResponse.json([])),
    http.get(`*/v1/medical-records/${RECORD_ID}/vitals`, () => HttpResponse.json([])),
    http.get("*/v1/masters/consultations", async ({ request }) => {
      const clinic = clinicOf(request);
      getHeadersByEndpoint.consultations = [
        ...getHeadersByEndpoint.consultations,
        request.headers.get("X-Clinic-ID"),
      ];
      const gate = clinic2GateByCategory.consultations;
      if (clinic === CLINIC2 && gate) await gate;
      return HttpResponse.json(clinic === CLINIC2 ? CLINIC2_CONSULTATIONS : CLINIC1_CONSULTATIONS);
    }),
    http.get("*/v1/masters/procedures", async ({ request }) => {
      const clinic = clinicOf(request);
      getHeadersByEndpoint.procedures = [
        ...getHeadersByEndpoint.procedures,
        request.headers.get("X-Clinic-ID"),
      ];
      const gate = clinic2GateByCategory.procedures;
      if (clinic === CLINIC2 && gate) await gate;
      return HttpResponse.json(clinic === CLINIC2 ? CLINIC2_PROCEDURES : CLINIC1_PROCEDURES);
    }),
    http.get("*/v1/masters/vaccines", () => HttpResponse.json([])),
    http.get("*/v1/masters/checkup-types", () => HttpResponse.json([])),
    http.get("*/v1/masters/medicines", () => HttpResponse.json([])),
    http.post(`*/v1/medical-records/${RECORD_ID}/treatments`, async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>;
      const header = request.headers.get("X-Clinic-ID");
      const call: PostCall = { header, body };
      const allowed = header !== null ? ALLOWED_REF_BY_CLINIC[header] : undefined;
      if (!allowed) {
        rejectedPosts = [...rejectedPosts, call];
        return HttpResponse.json({ message: "unknown clinic" }, { status: 400 });
      }
      for (const key of ["consultation_id", "procedure_id"] as const) {
        const value = body[key];
        if (value !== undefined && value !== null && value !== allowed[key]) {
          rejectedPosts = [...rejectedPosts, call];
          return HttpResponse.json(
            { message: `${key} does not belong to clinic ${header}` },
            { status: 400 },
          );
        }
      }
      acceptedPosts = [...acceptedPosts, call];
      return HttpResponse.json({ id: "new-row" }, { status: 201 });
    }),
  );
}

/** 呼び出し側の QueryClient を共有する wrapper。キャッシュ境界の観測用。 */
function makeWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

/** テストごとにキャッシュを分離し retry を切って終端 error を即時にする。 */
function newQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

/** 実 TreatmentsTab を呼び出し側 QueryClient で描画し、実ダイアログを開く。 */
async function renderTreatmentsTab(recordClinicId: string | undefined, queryClient: QueryClient) {
  const view = render(
    <TreatmentsTab medicalRecordId={RECORD_ID} recordClinicId={recordClinicId} />,
    { wrapper: makeWrapper(queryClient) },
  );
  const user = userEvent.setup();
  await user.click(await screen.findByRole("button", { name: /マスタから追加/ }));
  return view;
}

beforeEach(() => {
  acceptedPosts = [];
  rejectedPosts = [];
  getHeadersByEndpoint = { consultations: [], procedures: [] };
  clinic2GateByCategory = {};
  useBaseHandlers();
});

describe("clinic-scoped マスタ hooks（実 axios + MSW 境界, PR700）", () => {
  it.each(CATEGORIES)(
    "単一 QueryClient で明示 clinic1/clinic2 の $category は別データを保持する",
    async ({ category, refId }) => {
      const queryClient = newQueryClient();
      const wrapper = makeWrapper(queryClient);
      const hook = category === "consultations" ? useGetAllConsultations : useGetAllProcedures;
      const c1 = renderHook(() => hook({ clinicId: CLINIC1 }), { wrapper });
      const c2 = renderHook(() => hook({ clinicId: CLINIC2 }), { wrapper });
      await waitFor(() => {
        expect(c1.result.current.isSuccess).toBe(true);
        expect(c2.result.current.isSuccess).toBe(true);
      });
      expect(c1.result.current.data?.[0]?.id).toBe(String(refId - 100));
      expect(c2.result.current.data?.[0]?.id).toBe(String(refId));
      expect(c2.result.current.data?.[0]?.price).toBe(EXPECTED_PRICE_BY_ID[refId]);
      expect(c1.result.current.data?.[0]?.price).toBe(EXPECTED_PRICE_BY_ID[refId - 100]);
      const keys = queryClient
        .getQueryCache()
        .getAll()
        .map((q) => q.queryKey);
      expect(keys).toContainEqual(queryKeys.masters.categoryForClinic(category, CLINIC1));
      expect(keys).toContainEqual(queryKeys.masters.categoryForClinic(category, CLINIC2));
      expect(getHeadersByEndpoint[category]).toEqual([CLINIC1, CLINIC2]);
    },
  );

  it.each(["consultations", "procedures"] as const)(
    "省略と空文字の clinicId は旧グローバルキーと暗黙ヘッダを維持する (%s)",
    async (category) => {
      const queryClient = newQueryClient();
      const wrapper = makeWrapper(queryClient);
      const hook = category === "consultations" ? useGetAllConsultations : useGetAllProcedures;
      const a = renderHook(() => hook(), { wrapper });
      const b = renderHook(() => hook({ clinicId: "" }), { wrapper });
      await waitFor(() => {
        expect(a.result.current.isSuccess).toBe(true);
        expect(b.result.current.isSuccess).toBe(true);
      });
      const keys = queryClient
        .getQueryCache()
        .getAll()
        .map((q) => q.queryKey);
      expect(keys).toContainEqual(queryKeys.masters.category(category));
      expect(keys).not.toContainEqual(queryKeys.masters.categoryForClinic(category, CLINIC2));
      expect(getHeadersByEndpoint[category].every((h) => h !== CLINIC2)).toBe(true);
    },
  );

  it("useGetAllProcedures({enabled:false, clinicId:'2'}) は一切リクエストしない", async () => {
    const queryClient = newQueryClient();
    const { result } = renderHook(
      () => useGetAllProcedures({ enabled: false, clinicId: CLINIC2 }),
      { wrapper: makeWrapper(queryClient) },
    );
    expect(result.current.fetchStatus).toBe("idle");
    await Promise.resolve();
    expect(getHeadersByEndpoint.procedures).toHaveLength(0);
  });

  it("category prefix invalidation は同一カテゴリの scoped1/scoped2 双方を refetch する", async () => {
    const queryClient = newQueryClient();
    const wrapper = makeWrapper(queryClient);
    const c1 = renderHook(() => useGetAllConsultations({ clinicId: CLINIC1 }), { wrapper });
    const c2 = renderHook(() => useGetAllConsultations({ clinicId: CLINIC2 }), { wrapper });
    await waitFor(() => {
      expect(c1.result.current.isSuccess).toBe(true);
      expect(c2.result.current.isSuccess).toBe(true);
    });
    expect(getHeadersByEndpoint.consultations).toEqual([CLINIC1, CLINIC2]);
    await queryClient.invalidateQueries({
      queryKey: queryKeys.masters.category("consultations"),
    });
    await waitFor(() =>
      expect(getHeadersByEndpoint.consultations).toEqual([CLINIC1, CLINIC2, CLINIC1, CLINIC2]),
    );
  });
});

describe("TreatmentsTab マスタ選択のレコードクリニックスコープ (PR700)", () => {
  it.each(CATEGORIES)(
    "recordClinicId=2 の $category 選択は clinic2 マスタを表示し scoped ID で POST する",
    async ({ globalName, scopedName, refKey, refId, price }) => {
      const queryClient = newQueryClient();
      const wrapper = makeWrapper(queryClient);

      const warmConsultations = renderHook(() => useGetAllConsultations(), { wrapper });
      const warmProcedures = renderHook(() => useGetAllProcedures(), { wrapper });
      await waitFor(() => {
        expect(warmConsultations.result.current.isSuccess).toBe(true);
        expect(warmProcedures.result.current.isSuccess).toBe(true);
      });
      expect(warmConsultations.result.current.data?.[0]?.id).toBe("101");
      expect(warmProcedures.result.current.data?.[0]?.id).toBe("102");

      await renderTreatmentsTab(CLINIC2, queryClient);

      expect(await screen.findByText(scopedName)).toBeInTheDocument();
      expect(screen.getByText(price)).toBeInTheDocument();
      expect(screen.queryByText(globalName)).not.toBeInTheDocument();

      await userEvent.setup().click(screen.getByText(scopedName));

      await waitFor(() => expect(acceptedPosts).toHaveLength(1));
      expect(acceptedPosts[0].header).toBe(CLINIC2);
      expect(acceptedPosts[0].body[refKey]).toBe(refId);
      expect(acceptedPosts[0].body.unit_price).toBe(EXPECTED_PRICE_BY_ID[refId]);
      expect(acceptedPosts[0].body.quantity).toBe(1);
      expect(rejectedPosts).toHaveLength(0);
    },
  );

  it.each(CATEGORIES)(
    "clinic2 応答が遅延している間は $category の clinic1 選択肢を提示・送信しない",
    async ({ category, globalName, scopedName, refKey, refId }) => {
      let releaseGate: () => void = () => {};
      clinic2GateByCategory = {
        [category]: new Promise<void>((resolve) => {
          releaseGate = resolve;
        }),
      };

      const queryClient = newQueryClient();
      const wrapper = makeWrapper(queryClient);
      const hook = category === "consultations" ? useGetAllConsultations : useGetAllProcedures;
      const warm = renderHook(() => hook({ clinicId: CLINIC1 }), { wrapper });
      await waitFor(() => expect(warm.result.current.isSuccess).toBe(true));

      const view = render(
        <TreatmentsTab medicalRecordId={RECORD_ID} recordClinicId={undefined} />,
        { wrapper },
      );
      const user = userEvent.setup();
      await user.click(await screen.findByRole("button", { name: /マスタから追加/ }));
      await screen.findByText(globalName);

      view.rerender(<TreatmentsTab medicalRecordId={RECORD_ID} recordClinicId={CLINIC2} />);

      await waitFor(() => expect(screen.queryByText(globalName)).not.toBeInTheDocument());
      expect(acceptedPosts).toHaveLength(0);
      expect(rejectedPosts).toHaveLength(0);

      clinic2GateByCategory = {};
      releaseGate();
      await user.click(await screen.findByText(scopedName));
      await waitFor(() => expect(acceptedPosts).toHaveLength(1));
      expect(acceptedPosts[0].body[refKey]).toBe(refId);
      expect(acceptedPosts[0].body.unit_price).toBe(EXPECTED_PRICE_BY_ID[refId]);
      expect(acceptedPosts[0].header).toBe(CLINIC2);
      expect(rejectedPosts).toHaveLength(0);
    },
  );

  it.each(CATEGORIES)(
    "scoped 403 ($category) は warm 済み scoped1/global データへフォールバックしない",
    async ({ category, globalName, scopedName }) => {
      const queryClient = newQueryClient();
      const wrapper = makeWrapper(queryClient);
      const hook = category === "consultations" ? useGetAllConsultations : useGetAllProcedures;
      const warmScoped = renderHook(() => hook({ clinicId: CLINIC1 }), { wrapper });
      const warmGlobal = renderHook(() => hook(), { wrapper });
      await waitFor(() => {
        expect(warmScoped.result.current.isSuccess).toBe(true);
        expect(warmGlobal.result.current.isSuccess).toBe(true);
      });

      const endpoint = `*/v1/masters/${category}`;
      server.use(
        http.get(endpoint, ({ request }) =>
          clinicOf(request) === CLINIC2
            ? HttpResponse.json({ message: "forbidden" }, { status: 403 })
            : HttpResponse.json(
                category === "consultations" ? CLINIC1_CONSULTATIONS : CLINIC1_PROCEDURES,
              ),
        ),
      );

      await renderTreatmentsTab(CLINIC2, queryClient);

      const scopedKey = queryKeys.masters.categoryForClinic(category, CLINIC2);
      await waitFor(() => expect(queryClient.getQueryState(scopedKey)?.status).toBe("error"));
      const otherName = CATEGORIES.find((c) => c.category !== category)!.scopedName;
      await screen.findByText(otherName);
      expect(screen.queryByText(globalName)).not.toBeInTheDocument();
      expect(screen.queryByText(scopedName)).not.toBeInTheDocument();
      expect(acceptedPosts).toHaveLength(0);
      expect(rejectedPosts).toHaveLength(0);
    },
  );
});

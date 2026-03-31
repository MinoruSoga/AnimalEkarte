import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useMedicalRecordForm } from "./use-medical-record-form";

// ──────────────────────────────────────────────────────────
// モック定義
// ──────────────────────────────────────────────────────────

const mockNavigate = vi.fn();
let mockSearchParams = new URLSearchParams();

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParams, vi.fn()],
  useLocation: () => ({ pathname: "/medical-records/10", search: "", state: null }),
}));

vi.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}));

vi.mock("sonner", () => ({ toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

// API フック群をすべてスタブ化（デフォルト: データなし・ローディングなし）
const noData = { data: undefined, isLoading: false, isError: false };
const noMutation = { mutateAsync: vi.fn(), isPending: false };

vi.mock("@/hooks/use-pet", () => ({ useGetPet: () => noData }));
vi.mock("@/hooks/use-owner", () => ({ useGetOwner: () => noData }));
vi.mock("../api/get-medical-record", () => ({ useGetMedicalRecord: () => noData }));
vi.mock("../api/create-medical-record", () => ({ useCreateMedicalRecord: () => noMutation }));
vi.mock("../api/update-medical-record", () => ({ useUpdateMedicalRecord: () => noMutation }));
vi.mock("../api/inquiries", () => ({ useUpdateInquiry: () => noMutation }));
vi.mock("../api/treatment-plans", () => ({ useUpdateTreatmentPlan: () => noMutation }));

// ──────────────────────────────────────────────────────────
// テスト
// ──────────────────────────────────────────────────────────

describe("useMedicalRecordForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = new URLSearchParams();
  });

  // ──────────────────────────
  // isNewRecord
  // ──────────────────────────
  describe("isNewRecord", () => {
    it("recordId を渡さないと isNewRecord = true になる（バグ再現確認）", () => {
      const { result } = renderHook(() => useMedicalRecordForm());
      expect(result.current.isNewRecord).toBe(true);
    });

    it("recordId を渡すと isNewRecord = false になる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.isNewRecord).toBe(false);
    });
  });

  // ──────────────────────────
  // shouldRedirectToSelectPet（回帰テスト: BUG-カルテ編集）
  // ──────────────────────────
  describe("shouldRedirectToSelectPet", () => {
    it("【回帰】編集時(recordId あり) は shouldRedirectToSelectPet = false", () => {
      // このテストが失敗したら useMedicalRecordForm() に recordId が渡っていない
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.shouldRedirectToSelectPet).toBe(false);
    });

    it("新規作成かつ petId なし → shouldRedirectToSelectPet = true", () => {
      mockSearchParams = new URLSearchParams(); // petId なし
      const { result } = renderHook(() => useMedicalRecordForm());
      expect(result.current.shouldRedirectToSelectPet).toBe(true);
    });

    it("新規作成かつ petId あり → shouldRedirectToSelectPet = false", () => {
      mockSearchParams = new URLSearchParams({ petId: "1" });
      const { result } = renderHook(() => useMedicalRecordForm());
      expect(result.current.shouldRedirectToSelectPet).toBe(false);
    });
  });
});

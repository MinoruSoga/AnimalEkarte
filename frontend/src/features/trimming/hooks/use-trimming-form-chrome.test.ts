import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { toast } from "sonner";
import { useTrimmingFormChrome } from "./use-trimming-form-chrome";

const mockNavigate = vi.fn();
const mockMarkDirty = vi.fn();
const mockMarkClean = vi.fn();

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock("@/hooks/use-master-items", () => ({
  useGetMasterItems: () => ({ data: [] }),
}));

vi.mock("@/hooks/use-trimming-course-types", () => ({
  useGetTrimmingCourseTypes: () => ({ data: [] }),
}));

vi.mock("./use-trimming-history", () => ({
  useTrimmingHistory: () => ({
    sortedHistory: [],
    isHistoryLoading: false,
    historySearchTerm: "",
    historySortOrder: "desc",
    historyDateRange: { from: "", to: "" },
    setHistorySearchTerm: vi.fn(),
    setHistorySortOrder: vi.fn(),
    handleHistoryClear: vi.fn(),
    setHistoryDateRangeFrom: vi.fn(),
    setHistoryDateRangeTo: vi.fn(),
    handleHistoryClick: (updates: Record<string, unknown>) => updates,
  }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/config/paths", () => ({
  paths: {
    trimming: { getHref: () => "/trimming" },
  },
}));

const emptyFormData = {
  reservationTypeId: "",
  startTime: "",
  endTime: "",
  styleRequest: "",
  styleImage: null,
  bw: "",
  bwUnit: "Kg" as const,
  bt: "",
  usedShampoo: "",
  usedRibbon: "",
  remarks: "",
  completedImage: null,
  courseId: "",
  optionIds: [],
  staffId: "",
  staffName: "",
  initialStatus: "in_consultation" as const,
  nextScheduleType: "4weeks",
  nextDate: "",
};

function renderChrome(overrides: Partial<Parameters<typeof useTrimmingFormChrome>[0]> = {}) {
  const setFormData = vi.fn();
  const handleDelete = vi.fn();
  const result = renderHook(() =>
    useTrimmingFormChrome({
      formData: emptyFormData,
      setFormData,
      formState: { success: false, timestamp: 0 },
      selectedPetId: "10",
      redirectPath: "/trimming",
      fromPath: undefined,
      handleDelete,
      unsaved: { isDirty: false, markDirty: mockMarkDirty, markClean: mockMarkClean },
      ...overrides,
    }),
  );
  return { ...result, setFormData, handleDelete };
}

describe("useTrimmingFormChrome", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("fromPath があれば handleBack はそこへ戻る", () => {
    const { result } = renderChrome({ fromPath: "/reception" });

    act(() => {
      result.current.handleBack();
    });

    expect(mockNavigate).toHaveBeenCalledWith("/reception");
  });

  it("fromPath がなければ handleBack は一覧へ戻る", () => {
    const { result } = renderChrome();

    act(() => {
      result.current.handleBack();
    });

    expect(mockNavigate).toHaveBeenCalledWith("/trimming");
  });

  it("handleFormChange は dirty にして setFormData する", () => {
    const { result, setFormData } = renderChrome();

    act(() => {
      result.current.handleFormChange({ courseId: "4" });
    });

    expect(mockMarkDirty).toHaveBeenCalledTimes(1);
    expect(setFormData).toHaveBeenCalledWith({ courseId: "4" });
  });

  describe("handleHistoryClick (EMR-282)", () => {
    const otherElementScroll = vi.fn();
    const appendedAnchors: HTMLElement[] = [];

    const appendAnchor = (id: string): HTMLElement => {
      const element = document.createElement("div");
      element.id = id;
      document.body.appendChild(element);
      appendedAnchors.push(element);
      return element;
    };

    beforeEach(() => {
      // setup.ts の no-op stub では呼び出しを観測できないため、
      // prototype 側を spy に差し替えて「他の要素への scroll」を検出する。
      Element.prototype.scrollIntoView = otherElementScroll;
    });

    afterEach(() => {
      for (const element of appendedAnchors) element.remove();
      appendedAnchors.length = 0;
    });

    it("コピーしたフィールドの先頭アンカーへ smooth/center でスクロールし通知する", () => {
      const courseAnchor = appendAnchor("courseId");
      const courseScroll = vi.fn();
      courseAnchor.scrollIntoView = courseScroll;
      appendAnchor("staffId");
      appendAnchor("trimming-style-request");
      appendAnchor("trimming-used-shampoo");
      appendAnchor("trimming-used-ribbon");
      appendAnchor("trimming-remarks");
      const { result, setFormData } = renderChrome();
      const updates = {
        courseId: "3",
        optionIds: ["7"],
        styleRequest: "短めカット",
        usedShampoo: "低刺激シャンプー",
        usedRibbon: "赤リボン",
        remarks: "左耳短め",
        staffName: "田中",
      };

      act(() => {
        result.current.handleHistoryClick(updates);
      });

      expect(setFormData).toHaveBeenCalledWith(updates);
      expect(courseScroll).toHaveBeenCalledTimes(1);
      expect(courseScroll).toHaveBeenCalledWith({ behavior: "smooth", block: "center" });
      expect(otherElementScroll).not.toHaveBeenCalled();
      expect(toast.success).toHaveBeenCalledWith("前回の内容を反映しました");
    });

    it("courseId が無いコピーでは次に解決できるアンカーへスクロールする", () => {
      const remarksAnchor = appendAnchor("trimming-remarks");
      const remarksScroll = vi.fn();
      remarksAnchor.scrollIntoView = remarksScroll;
      const { result } = renderChrome();

      act(() => {
        result.current.handleHistoryClick({ remarks: "前回メモ", staffName: "佐藤" });
      });

      expect(remarksScroll).toHaveBeenCalledTimes(1);
      expect(remarksScroll).toHaveBeenCalledWith({ behavior: "smooth", block: "center" });
      expect(otherElementScroll).not.toHaveBeenCalled();
      expect(toast.success).toHaveBeenCalledWith("前回の内容を反映しました");
    });

    it("optionIds のみのコピーでは選択肢チェックボックスのアンカーへスクロールする", () => {
      const optionAnchor = appendAnchor("option-7");
      const optionScroll = vi.fn();
      optionAnchor.scrollIntoView = optionScroll;
      const { result } = renderChrome();

      act(() => {
        result.current.handleHistoryClick({ optionIds: ["7"] });
      });

      expect(optionScroll).toHaveBeenCalledTimes(1);
      expect(optionScroll).toHaveBeenCalledWith({ behavior: "smooth", block: "center" });
    });

    it("解決できるアンカーが無くても通知は出る", () => {
      const { result } = renderChrome();

      act(() => {
        result.current.handleHistoryClick({ courseId: "3", remarks: "x" });
      });

      expect(otherElementScroll).not.toHaveBeenCalled();
      expect(toast.success).toHaveBeenCalledWith("前回の内容を反映しました");
    });
  });
});

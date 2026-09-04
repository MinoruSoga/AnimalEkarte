import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
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

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: mockMarkDirty,
    markClean: mockMarkClean,
  }),
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
});

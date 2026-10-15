import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { TrimmingFormData } from "@/types/trimming";
import { TrimmingFormColumns, type TrimmingPatient } from "./TrimmingFormColumns";

const baseFormData: TrimmingFormData = {
  reservationTypeId: "",
  startTime: "",
  endTime: "",
  styleRequest: "",
  styleImage: null,
  bw: "",
  bwUnit: "Kg",
  bt: "",
  usedShampoo: "",
  usedRibbon: "",
  remarks: "",
  completedImage: null,
  courseId: "",
  optionIds: [],
  staffId: "",
  staffName: "鈴木 一郎",
  initialStatus: "in_consultation",
  nextScheduleType: "4weeks",
  nextDate: "",
};

const basePatient: TrimmingPatient = {
  ownerName: "山田 太郎",
  name: "ポチ",
};

function renderColumns(
  patient: TrimmingPatient = basePatient,
  fieldErrors: Record<string, string> = {},
  extraProps: { showRequiredMarks?: boolean } = {},
) {
  return render(
    <MemoryRouter>
      <TrimmingFormColumns
        selectedPet={patient}
        showRequiredMarks={extraProps.showRequiredMarks}
        formData={baseFormData}
        fieldErrors={fieldErrors}
        courses={[]}
        options={[]}
        styleImagePreview={null}
        completedImagePreview={null}
        sortedHistory={[]}
        isHistoryLoading={false}
        showInitialStatusSelector={false}
        onFormChange={vi.fn()}
        onOpenCourseModal={vi.fn()}
        onOpenStaffModal={vi.fn()}
        onStyleImageChange={vi.fn()}
        onCompletedImageChange={vi.fn()}
        onRemoveStyleImage={vi.fn()}
        onRemoveCompletedImage={vi.fn()}
        onHistoryClick={vi.fn()}
      />
    </MemoryRouter>,
  );
}

describe("TrimmingFormColumns", () => {
  it("EMR-272: 担当スタッフのラベルを「担当医」ではなく「担当」と表示する", () => {
    renderColumns();

    const staffButton = screen.getByRole("button", { name: "担当鈴木 一郎" });
    expect(staffButton).toBeInTheDocument();
    expect(staffButton).not.toHaveAccessibleName(/担当医/);
  });

  it("EMR-412 X14: トリミング新規のみ担当ボタンとコースラベルに必須印を出す", () => {
    const { unmount } = renderColumns(basePatient, {}, { showRequiredMarks: true });

    const staffButton = screen.getByRole("button", { name: /担当鈴木 一郎/ });
    expect(staffButton).toHaveTextContent("*");
    const courseLabel = screen.getByText("コース");
    expect(courseLabel).toHaveTextContent("*");
    unmount();

    // 既存編集では必須印を出さない（既定 false）。
    renderColumns(basePatient, {});
    expect(screen.getByRole("button", { name: /担当鈴木 一郎/ })).not.toHaveTextContent("*");
    expect(screen.getByText("コース")).not.toHaveTextContent("*");
  });
});

// ─────────────────────────────────────────────────────────────
// PR679 F6: コース必須印は aria-hidden の装飾のみではなく、
// トリガーボタンへ（必須）の支援技術説明を aria-describedby で結び付ける。
// ─────────────────────────────────────────────────────────────

describe("TrimmingFormColumns コース必須の支援技術説明 (PR679 F6)", () => {
  it("showRequiredMarks のときコーストリガーが（必須）説明を aria-describedby で参照し Label と対応付く", () => {
    renderColumns(basePatient, {}, { showRequiredMarks: true });

    const trigger = document.getElementById("courseId");
    expect(trigger).not.toBeNull();
    const describedBy = trigger?.getAttribute("aria-describedby");
    expect(describedBy).toBeTruthy();
    const description = describedBy ? document.getElementById(describedBy) : null;
    expect(description).not.toBeNull();
    expect(description).toHaveTextContent("（必須）");

    // ラベルとも対応付く（htmlFor → ボタン id、または aria-labelledby のいずれか）
    const label = screen.getByText("コース");
    const isLabelled =
      label.getAttribute("for") === "courseId" ||
      trigger?.getAttribute("aria-labelledby") === label.id;
    expect(isLabelled).toBe(true);
  });

  it("showRequiredMarks=false なら（必須）説明も aria-describedby も付けない", () => {
    renderColumns(basePatient, {});

    const trigger = document.getElementById("courseId");
    expect(trigger).not.toBeNull();
    expect(screen.queryByText("（必須）")).not.toBeInTheDocument();
    const describedBy = trigger?.getAttribute("aria-describedby");
    if (describedBy) {
      expect(document.getElementById(describedBy)).not.toHaveTextContent("（必須）");
    }
  });
});

describe("TrimmingFormColumns 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月の生存ペットは PatientInfoCard 上に Cake バッジを表示する", () => {
    renderColumns({ ...basePatient, birthDate: "2018-10-05", status: "生存" });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05" as const],
    ["birthDate 未登録", undefined],
  ])("%s なら Cake バッジを表示しない", (_label, birthDate) => {
    renderColumns({ ...basePatient, birthDate, status: "生存" });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=死亡）には誕生月でも Cake バッジを表示しない", () => {
    renderColumns({ ...basePatient, birthDate: "2018-10-05", status: "死亡" });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});

describe("TrimmingFormColumns 必須エラー表示 (EMR-319)", () => {
  it("fieldErrors.staffId で担当ボタンに aria-invalid+赤枠、エラー文が出る", () => {
    renderColumns(basePatient, { staffId: "担当者を選択してください" });

    const staffButton = document.getElementById("staffId");
    expect(staffButton).toHaveAttribute("aria-invalid", "true");
    expect(staffButton?.className).toContain("ring-2");
    expect(screen.getByRole("alert")).toHaveTextContent("担当者を選択してください");
  });

  it("fieldErrors.courseId でコース選択トリガーに aria-invalid+赤枠、エラー文が出る", () => {
    const { container } = renderColumns(basePatient, {
      courseId: "コースを選択してください",
    });

    const courseTrigger = container.querySelector("[aria-invalid='true']");
    expect(courseTrigger).not.toBeNull();
    expect(courseTrigger?.className).toContain("ring-2");
    expect(screen.getByRole("alert")).toHaveTextContent("コースを選択してください");
  });

  it("fieldErrors が空なら aria-invalid 付きコントロールは存在しない", () => {
    const { container } = renderColumns(basePatient, {});

    expect(container.querySelector("[aria-invalid='true']")).toBeNull();
  });
});

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("../../hooks/use-medical-record-addenda", () => ({
  useMedicalRecordAddenda: vi.fn(),
  useCreateMedicalRecordAddendum: vi.fn(),
}));

import { MedicalRecordAddenda } from "./MedicalRecordAddenda";
import {
  useMedicalRecordAddenda,
  useCreateMedicalRecordAddendum,
} from "../../hooks/use-medical-record-addenda";
import type { MedicalRecordAddendum as AddendumType } from "@/types/generated/models";

function makeAddendum(overrides: Partial<AddendumType> = {}): AddendumType {
  return {
    id: 1,
    medical_record_id: 10,
    clinic_id: 1,
    author_user_id: 5,
    before_text: "before",
    after_text: "修正後内容",
    reason: "誤記修正",
    created_at: "2026-01-01T10:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(useMedicalRecordAddenda).mockReturnValue({
    data: [],
    isLoading: false,
  } as ReturnType<typeof useMedicalRecordAddenda>);
  vi.mocked(useCreateMedicalRecordAddendum).mockReturnValue({
    mutateAsync: vi.fn().mockResolvedValue({}),
  } as unknown as ReturnType<typeof useCreateMedicalRecordAddendum>);
});

describe("MedicalRecordAddenda", () => {
  it("作成中 状態では何も表示しない", () => {
    const { container } = render(
      <MedicalRecordAddenda medicalRecordId="1" canEdit={true} recordStatus="作成中" />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("確定済 状態で追記セクションを表示する", () => {
    render(<MedicalRecordAddenda medicalRecordId="1" canEdit={false} recordStatus="確定済" />);
    expect(screen.getByTestId("medical-record-addenda")).toBeInTheDocument();
  });

  it("canEdit=true で「追記する」ボタンを表示する", () => {
    render(<MedicalRecordAddenda medicalRecordId="1" canEdit={true} recordStatus="確定済" />);
    expect(screen.getByTestId("addendum-open-button")).toBeInTheDocument();
    expect(screen.getByText("追記する")).toBeInTheDocument();
  });

  it("canEdit=false で「追記する」ボタンを表示しない・リストは表示する", () => {
    render(<MedicalRecordAddenda medicalRecordId="1" canEdit={false} recordStatus="確定済" />);
    expect(screen.queryByTestId("addendum-open-button")).toBeNull();
    expect(screen.getByTestId("medical-record-addenda")).toBeInTheDocument();
  });

  it("addenda 0 件で empty state を表示する", () => {
    render(<MedicalRecordAddenda medicalRecordId="1" canEdit={false} recordStatus="確定済" />);
    expect(screen.getByTestId("addenda-empty")).toBeInTheDocument();
    expect(screen.getByText("追記はありません")).toBeInTheDocument();
  });

  it("addenda 3 件を時系列昇順で表示する", () => {
    vi.mocked(useMedicalRecordAddenda).mockReturnValue({
      data: [
        makeAddendum({ id: 3, after_text: "3番目", created_at: "2026-01-03T10:00:00Z" }),
        makeAddendum({ id: 1, after_text: "1番目", created_at: "2026-01-01T10:00:00Z" }),
        makeAddendum({ id: 2, after_text: "2番目", created_at: "2026-01-02T10:00:00Z" }),
      ],
      isLoading: false,
    } as ReturnType<typeof useMedicalRecordAddenda>);

    render(<MedicalRecordAddenda medicalRecordId="1" canEdit={false} recordStatus="確定済" />);

    const items = screen.getAllByTestId("addendum-item");
    expect(items).toHaveLength(3);
    expect(items[0]).toHaveTextContent("1番目");
    expect(items[1]).toHaveTextContent("2番目");
    expect(items[2]).toHaveTextContent("3番目");
  });

  it("「追記する」ボタンクリックでモーダルが開く", async () => {
    const user = userEvent.setup();
    render(<MedicalRecordAddenda medicalRecordId="1" canEdit={true} recordStatus="確定済" />);
    await user.click(screen.getByTestId("addendum-open-button"));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });
});

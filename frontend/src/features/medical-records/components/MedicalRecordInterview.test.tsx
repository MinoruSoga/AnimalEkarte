import type { ComponentProps, ReactNode } from "react";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { MedicalRecordInterview } from "./MedicalRecordInterview";
import {
  DEFAULT_CHIEF_COMPLAINT,
  DEFAULT_TREATMENT_POLICY,
} from "../hooks/use-medical-record-form-model";
import type { InterviewHistoryItem } from "../types";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true }),
}));

vi.mock("../api/get-chief-complaint-types", () => ({
  useGetChiefComplaintTypes: () => ({ data: [], isLoading: false }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    warning: vi.fn(),
  },
}));

// EMR-325: 明細複写の価格注意（チケット最新コメントの訂正文）。
// 確認ダイアログ説明文と全コピー経路の toast で同一文が使われる。
const COPY_PRICE_CAUTION =
  "対応するマスタが見つかる明細は取得済みのマスタ価格で追加されます。その他・対応マスタが見つからない明細は元の単価を引き継ぎます。コピー後に単価をご確認ください。";

type InterviewProps = ComponentProps<typeof MedicalRecordInterview>;

// EMR-356: weightPanel スロットは未実装のため、現行 Props と拡張案の合成型で fixture を渡す。
type WeightAwareInterviewProps = InterviewProps & { weightPanel?: ReactNode };

const COPY_HISTORY_ITEMS: InterviewHistoryItem[] = [
  {
    id: "10",
    date: "2026/02/01 (日)",
    author: "田中",
    type: "再診",
    title: "前回カルテ",
    content: "元気がない",
    copySource: {
      recordId: "101",
      chiefComplaint: "前回の主訴詳細",
      treatmentPolicy: "前回の治療方針",
      chiefComplaintTypeId: 7,
    },
  },
];

function renderInterview(overrides: Partial<WeightAwareInterviewProps> = {}) {
  const props: WeightAwareInterviewProps = {
    chiefComplaint: DEFAULT_CHIEF_COMPLAINT,
    setChiefComplaint: vi.fn(),
    chiefComplaintTypeId: null,
    setChiefComplaintTypeId: vi.fn(),
    treatmentPolicy: DEFAULT_TREATMENT_POLICY,
    setTreatmentPolicy: vi.fn(),
    historyItems: COPY_HISTORY_ITEMS,
    ...overrides,
  };
  render(
    <MemoryRouter>
      <MedicalRecordInterview {...props} />
    </MemoryRouter>,
  );
  return props;
}

describe("MedicalRecordInterview — form field semantics", () => {
  it("問診内の各form fieldに明示labelとidを接続する", () => {
    renderInterview({ historyItems: [] });

    expect(screen.getByRole("combobox", { name: "主訴区分" })).toHaveAttribute(
      "id",
      "medical-record-chief-complaint-type",
    );
    expect(screen.getByRole("textbox", { name: "主訴詳細" })).toHaveAttribute(
      "name",
      "chiefComplaint",
    );
    expect(screen.getByRole("textbox", { name: "治療方針" })).toHaveAttribute(
      "name",
      "treatmentPolicy",
    );
    expect(screen.getByRole("textbox", { name: "過去の履歴を検索" })).toHaveAttribute(
      "name",
      "medicalRecordHistorySearch",
    );
  });
});

describe("MedicalRecordInterview — 編集ロック（BUG-035）", () => {
  it("isLocked で編集2列は disabled fieldset に入り、履歴検索・閲覧はロック外で有効のまま", () => {
    renderInterview({ isLocked: true });

    const chiefComplaint = screen.getByRole("textbox", { name: "主訴詳細" });
    const search = screen.getByRole("textbox", { name: "過去の履歴を検索" });

    // 編集列は disabled fieldset の子孫として無効化される
    expect(chiefComplaint).toBeDisabled();
    expect(chiefComplaint.closest("fieldset")).toBeDisabled();
    expect(screen.getByRole("textbox", { name: "治療方針" })).toBeDisabled();
    expect(screen.getByRole("combobox", { name: "主訴区分" })).toBeDisabled();

    // 読み取り専用の履歴検索はロック境界の外なので使える
    expect(search.closest("fieldset")).toBeNull();
    expect(search).not.toBeDisabled();
    expect(screen.getByRole("link", { name: /前回カルテ/ })).toHaveAttribute(
      "href",
      "/medical-records/10",
    );
    // 履歴→現在カルテへの複写は mutation なのでロックする
    expect(screen.getByRole("button", { name: "コピー" })).toBeDisabled();
  });
});

describe("MedicalRecordInterview — 前回複写（コピー）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("未編集の既定値では コピー が確認なしで3項目へ即時適用され、価格注意を通知する", async () => {
    const user = userEvent.setup();
    const props = renderInterview();

    await user.click(screen.getByRole("button", { name: "コピー" }));

    expect(props.setChiefComplaint).toHaveBeenCalledWith("前回の主訴詳細");
    expect(props.setTreatmentPolicy).toHaveBeenCalledWith("前回の治療方針");
    expect(props.setChiefComplaintTypeId).toHaveBeenCalledWith(7);
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    // EMR-325: ダイアログを通らない即時適用経路でも同一の価格注意を出す
    expect(toast.warning).toHaveBeenCalledTimes(1);
    expect(toast.warning).toHaveBeenCalledWith(COPY_PRICE_CAUTION);
  });

  it("主訴詳細が編集済みなら ConfirmDialog の確定で適用され、価格注意を通知する", async () => {
    const user = userEvent.setup();
    const props = renderInterview({ chiefComplaint: "編集済みの主訴" });

    await user.click(screen.getByRole("button", { name: "コピー" }));
    const dialog = await screen.findByRole("alertdialog");
    // EMR-325: ダイアログを開いただけでは通知せず、確定（実際の適用）でのみ通知する
    expect(toast.warning).not.toHaveBeenCalled();
    await user.click(within(dialog).getByRole("button", { name: "コピー" }));

    expect(props.setChiefComplaint).toHaveBeenCalledWith("前回の主訴詳細");
    expect(props.setTreatmentPolicy).toHaveBeenCalledWith("前回の治療方針");
    expect(props.setChiefComplaintTypeId).toHaveBeenCalledWith(7);
    expect(toast.warning).toHaveBeenCalledTimes(1);
    expect(toast.warning).toHaveBeenCalledWith(COPY_PRICE_CAUTION);
  });

  it("治療方針が編集済みでも ConfirmDialog のキャンセルは現在値を保持し価格注意も出さない", async () => {
    const user = userEvent.setup();
    const props = renderInterview({ treatmentPolicy: "独自の治療方針" });

    await user.click(screen.getByRole("button", { name: "コピー" }));
    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));

    expect(props.setChiefComplaint).not.toHaveBeenCalled();
    expect(props.setTreatmentPolicy).not.toHaveBeenCalled();
    expect(props.setChiefComplaintTypeId).not.toHaveBeenCalled();
    expect(toast.warning).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    });
  });

  it("主訴区分が選択済みでも未編集扱いせず ConfirmDialog を開く", async () => {
    const user = userEvent.setup();
    const props = renderInterview({ chiefComplaintTypeId: 3 });

    await user.click(screen.getByRole("button", { name: "コピー" }));

    expect(await screen.findByRole("alertdialog")).toBeInTheDocument();
    expect(props.setChiefComplaint).not.toHaveBeenCalled();
    expect(props.setTreatmentPolicy).not.toHaveBeenCalled();
    expect(props.setChiefComplaintTypeId).not.toHaveBeenCalled();
  });

  it("copySource に存在する項目だけを適用する", async () => {
    const user = userEvent.setup();
    const partialItems: InterviewHistoryItem[] = [
      {
        ...COPY_HISTORY_ITEMS[0],
        id: "11",
        copySource: { chiefComplaint: "部分的な主訴" },
      },
    ];
    const props = renderInterview({ historyItems: partialItems });

    await user.click(screen.getByRole("button", { name: "コピー" }));

    expect(props.setChiefComplaint).toHaveBeenCalledWith("部分的な主訴");
    expect(props.setTreatmentPolicy).not.toHaveBeenCalled();
    expect(props.setChiefComplaintTypeId).not.toHaveBeenCalled();
  });

  it("即時適用パスで onCopyRecordTreatments が copySource.recordId で呼ばれる（EMR-219）", async () => {
    const user = userEvent.setup();
    const onCopyRecordTreatments = vi.fn();
    renderInterview({ onCopyRecordTreatments });

    await user.click(screen.getByRole("button", { name: "コピー" }));

    expect(onCopyRecordTreatments).toHaveBeenCalledTimes(1);
    expect(onCopyRecordTreatments).toHaveBeenCalledWith("101");
  });

  it("ConfirmDialog 確定パスでも onCopyRecordTreatments が copySource.recordId で呼ばれる（EMR-219）", async () => {
    const user = userEvent.setup();
    const onCopyRecordTreatments = vi.fn();
    renderInterview({ chiefComplaint: "編集済みの主訴", onCopyRecordTreatments });

    await user.click(screen.getByRole("button", { name: "コピー" }));
    const dialog = await screen.findByRole("alertdialog");
    // EMR-219: 治療明細も追加される旨をダイアログ説明文で告知する
    expect(dialog).toHaveTextContent("前回の治療明細も追加されます");
    // EMR-325: 価格解決の両分岐（マスタ一致→取得済みマスタ価格 / 不一致・その他→元単価
    // 引継ぎ）とコピー後の単価確認を説明文で正確に告知する（「すべて現行マスタ価格に更新」
    // は言い過ぎのため使わない）
    expect(dialog).toHaveTextContent(
      "対応するマスタが見つかる明細は取得済みのマスタ価格で追加されます",
    );
    expect(dialog).toHaveTextContent("対応マスタが見つからない明細は元の単価を引き継ぎます");
    expect(dialog).toHaveTextContent("コピー後に単価をご確認ください");
    expect(dialog).not.toHaveTextContent("すべて現行マスタ価格");
    await user.click(within(dialog).getByRole("button", { name: "コピー" }));

    expect(onCopyRecordTreatments).toHaveBeenCalledTimes(1);
    expect(onCopyRecordTreatments).toHaveBeenCalledWith("101");
  });

  it("ConfirmDialog キャンセルでは onCopyRecordTreatments は呼ばれない（EMR-219）", async () => {
    const user = userEvent.setup();
    const onCopyRecordTreatments = vi.fn();
    renderInterview({ treatmentPolicy: "独自の治療方針", onCopyRecordTreatments });

    await user.click(screen.getByRole("button", { name: "コピー" }));
    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));

    expect(onCopyRecordTreatments).not.toHaveBeenCalled();
  });

  it("copySource.recordId が無い履歴行では onCopyRecordTreatments は呼ばれない（EMR-219）", async () => {
    const user = userEvent.setup();
    const onCopyRecordTreatments = vi.fn();
    const noRecordItems: InterviewHistoryItem[] = [
      {
        ...COPY_HISTORY_ITEMS[0],
        id: "12",
        copySource: { chiefComplaint: "部分的な主訴" },
      },
    ];
    const props = renderInterview({
      historyItems: noRecordItems,
      onCopyRecordTreatments,
    });

    await user.click(screen.getByRole("button", { name: "コピー" }));

    expect(props.setChiefComplaint).toHaveBeenCalledWith("部分的な主訴");
    expect(onCopyRecordTreatments).not.toHaveBeenCalled();
    // EMR-325: recordId 無しの問診のみ複写でも価格注意は applyCopySource から通知される
    expect(toast.warning).toHaveBeenCalledTimes(1);
    expect(toast.warning).toHaveBeenCalledWith(COPY_PRICE_CAUTION);
  });
});

describe("MedicalRecordInterview — 体重パネルスロット（EMR-356）", () => {
  it("weightPanel を問診画面内に描画する（同じペットの過去体重と今回入力導線）", () => {
    renderInterview({
      historyItems: [],
      weightPanel: <div data-testid="interview-weight-panel">体重パネル</div>,
    });

    expect(screen.getByTestId("interview-weight-panel")).toBeInTheDocument();
    expect(screen.getByText("体重パネル")).toBeInTheDocument();
  });

  it("isLocked 時、体重パネルは disabled fieldset 内で mutation できない", () => {
    renderInterview({
      isLocked: true,
      weightPanel: (
        <button type="button" data-testid="interview-weight-save">
          体重を記録
        </button>
      ),
    });

    const control = screen.getByTestId("interview-weight-save");
    // コピー ボタンと同じく mutation はロック境界内（読み取り専用の履歴検索は外）
    expect(control.closest("fieldset")).toBeDisabled();
  });
});

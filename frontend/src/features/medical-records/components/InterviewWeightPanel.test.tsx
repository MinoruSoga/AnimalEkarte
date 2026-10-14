import type { ComponentProps, FormEvent, ReactNode } from "react";
import { isValidElement, startTransition } from "react";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { InterviewWeightPanel } from "./InterviewWeightPanel";
import type { BodyWeightUnit, Vital } from "../types";

// EMR-356: 問診画面の体重パネル。既存バイタル SoT/input path（api/vitals hooks）を
// 再利用する契約を、hook mock で検証する。現時点で ./InterviewWeightPanel は
// 未実装であり、このファイル全体が RED になることが意図である。

const { createVitalMock, getVitalsArgsMock, createVitalArgsMock, capturedAction, mockState } =
  vi.hoisted(() => {
    return {
      createVitalMock: vi.fn(),
      getVitalsArgsMock: vi.fn(),
      createVitalArgsMock: vi.fn(),
      capturedAction: { current: undefined as ((formData: FormData) => void) | undefined },
      mockState: {
        vitals: {
          data: [] as Vital[] | undefined,
          isLoading: false,
          isError: false,
        },
        permissions: { canView: true, canCreate: true, canEdit: true, canDelete: true },
        createPending: false,
      },
    };
  });

vi.mock("../api/vitals", () => ({
  useGetVitals: (medicalRecordId: string, clinicId?: string, options?: { enabled?: boolean }) => {
    getVitalsArgsMock(medicalRecordId, clinicId, options);
    return mockState.vitals;
  },
  useCreateVital: (medicalRecordId: string, clinicId?: string) => {
    createVitalArgsMock(medicalRecordId, clinicId);
    return { mutateAsync: createVitalMock, isPending: mockState.createPending };
  },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mockState.permissions,
}));

vi.mock("react-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-dom")>();
  return {
    ...actual,
    createPortal: (
      children: ReactNode,
      container: Element | DocumentFragment,
      key?: string | null,
    ) => {
      if (
        isValidElement<ComponentProps<"form">>(children) &&
        typeof children.props.action === "function"
      ) {
        capturedAction.current = children.props.action;
      }
      return actual.createPortal(children, container, key);
    },
  };
});

type PanelProps = ComponentProps<typeof InterviewWeightPanel>;

/** EMR-356 計画型（types/index.ts 未拡張のため test 側で構造的に表現する）。 */
interface PastMeasurement {
  id: string;
  value: number;
  unit: BodyWeightUnit;
  measuredAt: string;
}

const PAST_MEASUREMENTS: PastMeasurement[] = [
  { id: "11", value: 4.2, unit: "Kg", measuredAt: "2026-03-01T09:30:00+09:00" },
];

function renderPanel(overrides: Partial<PanelProps> = {}) {
  const props: PanelProps = {
    medicalRecordId: "10",
    recordClinicId: "7",
    isPetDeceased: false,
    ...overrides,
  };
  const view = render(<InterviewWeightPanel {...props} />);
  return { props, ...view };
}

beforeEach(() => {
  createVitalMock.mockReset();
  createVitalMock.mockResolvedValue({});
  getVitalsArgsMock.mockClear();
  createVitalArgsMock.mockClear();
  mockState.vitals = { data: [], isLoading: false, isError: false };
  mockState.permissions = { canView: true, canCreate: true, canEdit: true, canDelete: true };
  mockState.createPending = false;
  capturedAction.current = undefined;
});

describe("InterviewWeightPanel — 過去体重の表示（EMR-356）", () => {
  it("過去体重を値・単位・測定日時つきで表示する", () => {
    const { container } = renderPanel({ pastMeasurements: PAST_MEASUREMENTS });

    expect(container).toHaveTextContent("4.2");
    expect(container).toHaveTextContent("Kg");
    // formatRecordedAt 契約（YYYY-MM-DD HH:mm JST wall）
    expect(container).toHaveTextContent("2026-03-01 09:30");
  });

  it("体重記録なしと履歴取得エラーを区別して表示する", () => {
    const empty = renderPanel({ pastMeasurements: [] });
    expect(empty.container).toHaveTextContent("体重の記録はありません");
    expect(empty.container).not.toHaveTextContent("取得できませんでした");
    empty.unmount();

    const errored = renderPanel({ pastMeasurements: [], historyError: true });
    expect(errored.container).toHaveTextContent("体重の履歴を取得できませんでした");
    expect(errored.container).not.toHaveTextContent("体重の記録はありません");
  });

  it("このカルテに保存済みの体重は今回の測定として表示し、過去測定と区別する", () => {
    mockState.vitals = {
      data: [
        {
          id: "v1",
          medical_record_id: "10",
          recorded_at: "2026-03-05T10:00:00+09:00",
          weight: 4.4,
          weight_unit: "Kg",
          version: 1,
          created_at: "2026-03-05T10:00:00+09:00",
          updated_at: "2026-03-05T10:00:00+09:00",
        },
      ],
      isLoading: false,
      isError: false,
    };
    const { container } = renderPanel({ pastMeasurements: PAST_MEASUREMENTS });

    // 今回測定と過去測定を区別できるラベル境界を持つ
    expect(container).toHaveTextContent("今回");
    expect(container).toHaveTextContent("過去");
    expect(container).toHaveTextContent("4.4");
    expect(container).toHaveTextContent("4.2");
  });
});

describe("InterviewWeightPanel — 今回体重の入力（EMR-356）", () => {
  it("明示保存で1回だけ createVital を呼び、体重のみを既存契約で送る", async () => {
    renderPanel({ pastMeasurements: PAST_MEASUREMENTS });

    // mount / 過去値表示だけでは mutation しない（過去値の自動コピー禁止）
    expect(createVitalMock).not.toHaveBeenCalled();
    const input = screen.getByLabelText("体重");
    expect(input).toHaveValue(""); // 過去値を入力欄へ prefill しない

    const user = userEvent.setup();
    await user.type(input, "4.5");
    await user.click(screen.getByRole("button", { name: "体重を記録" }));

    expect(createVitalMock).toHaveBeenCalledTimes(1);
    const input_arg = createVitalMock.mock.calls[0][0] as {
      recorded_at?: string;
      weight?: number | null;
      weight_unit?: string;
      temperature?: number | null;
      heart_rate?: number | null;
      respiration_rate?: number | null;
    };
    expect(input_arg).toMatchObject({ weight: 4.5, weight_unit: "Kg" });
    expect(typeof input_arg.recorded_at).toBe("string");
    // 体重のみの測定（他バイタル項目は既存 VitalsTab 経路のまま）
    expect(input_arg.temperature ?? null).toBeNull();
    expect(input_arg.heart_rate ?? null).toBeNull();
    expect(input_arg.respiration_rate ?? null).toBeNull();
    // 既存の既存カルテ・拠点横断契約: レコード自身の id/clinicId で hook を呼ぶ
    expect(createVitalArgsMock).toHaveBeenCalledWith("10", "7");
    expect(getVitalsArgsMock).toHaveBeenCalledWith(
      "10",
      "7",
      expect.objectContaining({ enabled: true }),
    );
  });

  it("Kg↔g トグルは未保存の入力値だけを換算し、過去表示は変えない", async () => {
    const { container } = renderPanel({ pastMeasurements: PAST_MEASUREMENTS });

    const input = screen.getByLabelText("体重");
    const user = userEvent.setup();
    await user.type(input, "4.5");
    await user.click(screen.getByRole("button", { name: "Kg" }));

    expect(input).toHaveValue("4500");
    // 過去表示の値・単位は換算しない（表示専用投影）
    expect(container).toHaveTextContent("4.2");

    await user.click(screen.getByRole("button", { name: "g" }));
    expect(input).toHaveValue("4.5");
  });

  it.each(["0", "-1", "abc"])("有限正数でない入力 %s は保存せずエラーを表示する", async (value) => {
    renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), value);
    await user.click(screen.getByRole("button", { name: "体重を記録" }));

    expect(createVitalMock).not.toHaveBeenCalled();
    expect(screen.getByText(/正の数値/)).toBeInTheDocument();
  });

  it("未入力では createVital を呼ばない（無操作=記録しない）", async () => {
    renderPanel();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "体重を記録" }));

    expect(createVitalMock).not.toHaveBeenCalled();
  });

  it("保存中は記録ボタンを無効化し二重送信を防ぐ", () => {
    mockState.createPending = true;
    renderPanel();

    expect(screen.getByRole("button", { name: "体重を記録" })).toBeDisabled();
  });
});

describe("InterviewWeightPanel — 臨床安全境界（EMR-356）", () => {
  it("死亡ペットでは記録 UI を出さない（過去体重の読取は残る）", () => {
    const { container } = renderPanel({
      isPetDeceased: true,
      pastMeasurements: PAST_MEASUREMENTS,
    });

    expect(screen.queryByRole("button", { name: "体重を記録" })).not.toBeInTheDocument();
    expect(screen.queryByLabelText("体重")).not.toBeInTheDocument();
    expect(container).toHaveTextContent("4.2");
  });

  it.each([false, true])("canEdit 不足では記録 UI を出さない（canCreate=%s）", (canCreate) => {
    mockState.permissions = { ...mockState.permissions, canCreate, canEdit: false };
    renderPanel();

    expect(screen.queryByRole("button", { name: "体重を記録" })).not.toBeInTheDocument();
    expect(screen.queryByLabelText("体重")).not.toBeInTheDocument();
    expect(document.querySelector('form[id^="interview-weight-form-"]')).toBeNull();
    expect(createVitalMock).not.toHaveBeenCalled();
  });

  it("canView 無しでは過去体重の読取も出さず、実際の vitals 取得も抑止する", () => {
    mockState.permissions = { ...mockState.permissions, canView: false };
    const { container } = renderPanel({ pastMeasurements: PAST_MEASUREMENTS });

    expect(container).not.toHaveTextContent("4.2");
    // DOM を隠すだけでなくクエリ自体を disabled にする（実 fetch 抑止は api/vitals 側の責務）
    expect(getVitalsArgsMock).toHaveBeenCalledWith(
      "10",
      "7",
      expect.objectContaining({ enabled: false }),
    );
  });

  it("medicalRecordId 未確定では記録 UI を出さない", () => {
    renderPanel({ medicalRecordId: undefined });

    expect(screen.queryByRole("button", { name: "体重を記録" })).not.toBeInTheDocument();
    expect(screen.queryByLabelText("体重")).not.toBeInTheDocument();
  });
});

describe("InterviewWeightPanel — 親 form との共存（EMR-356）", () => {
  it("親 form 内でネストした form を作らず、親 submit を発火しない", async () => {
    const onSubmit = vi.fn((e: FormEvent) => e.preventDefault());
    render(
      <form onSubmit={onSubmit}>
        <InterviewWeightPanel medicalRecordId="10" recordClinicId="7" />
      </form>,
    );

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");
    await user.click(screen.getByRole("button", { name: "体重を記録" }));

    expect(onSubmit).not.toHaveBeenCalled();
    // 親カルテ form の内側にネストした <form> を作らない。
    // 体重側の action 経路（portal form 等）は親の外に置かれ得るため、
    // document 全体の form 数ではなく「親内ネストなし」で検証する。
    const chartForm = document.querySelector("form");
    expect(chartForm).not.toBeNull();
    expect(chartForm?.querySelector("form")).toBeNull();
  });

  it("体重入力欄の Enter は親 form の submit ではなく体重記録として1回だけ処理する", async () => {
    const onSubmit = vi.fn((e: FormEvent) => e.preventDefault());
    render(
      <form onSubmit={onSubmit}>
        <InterviewWeightPanel medicalRecordId="10" recordClinicId="7" />
      </form>,
    );

    const user = userEvent.setup();
    // Enter はカルテ保存ではなく「体重を記録」と等価の明示操作
    await user.type(screen.getByLabelText("体重"), "4.5{Enter}");

    expect(onSubmit).not.toHaveBeenCalled();
    expect(createVitalMock).toHaveBeenCalledTimes(1);
    expect(createVitalMock.mock.calls[0][0]).toMatchObject({ weight: 4.5, weight_unit: "Kg" });
  });

  it("保存応答の未完了中は再クリックしても二重記録しない", async () => {
    let resolveSave: (value: unknown) => void = () => {};
    const savePromise = new Promise((resolve) => {
      resolveSave = resolve;
    });
    createVitalMock.mockReturnValue(savePromise);
    renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");
    const save = screen.getByRole("button", { name: "体重を記録" });
    await user.click(save);
    // 1回目の mutate が未完了の間、2回目のクリックは新しい記録を作らない
    await user.click(save);

    expect(createVitalMock).toHaveBeenCalledTimes(1);
    await act(async () => {
      resolveSave({});
      await savePromise;
    });
    expect(await screen.findByRole("status")).toHaveTextContent("体重を記録しました");
  });
});

describe("InterviewWeightPanel — 識別コンテキスト安全性（EMR-356）", () => {
  it("保存応答中に record が切り替わっても、古い応答が新しいカルテの入力を汚さない", async () => {
    let resolveSave: (value: unknown) => void = () => {};
    createVitalMock.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveSave = resolve;
        }),
    );
    const { props, rerender } = renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");
    await user.click(screen.getByRole("button", { name: "体重を記録" }));
    expect(createVitalMock).toHaveBeenCalledTimes(1);

    // 旧カルテの応答が戻る前に別カルテへ切替し、新コンテキストで新たな値を入力する
    rerender(<InterviewWeightPanel {...props} medicalRecordId="20" />);
    const newInput = await screen.findByLabelText("体重");
    await user.type(newInput, "7.7");
    await act(async () => {
      resolveSave({});
    });

    // 旧応答の遅延完了は新カルテで入力済みの値をクリア/上書きしない
    await waitFor(() => expect(createVitalMock).toHaveBeenCalledTimes(1));
    expect(screen.getByLabelText("体重")).toHaveValue("7.7");
  });

  it("入力途中で死亡状態へ変わると保存導線が取り下げられる", async () => {
    const { props, rerender } = renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");

    rerender(<InterviewWeightPanel {...props} isPetDeceased />);

    expect(screen.queryByRole("button", { name: "体重を記録" })).not.toBeInTheDocument();
    expect(screen.queryByLabelText("体重")).not.toBeInTheDocument();
  });

  it("入力途中で編集権限が剥奪されると保存導線が取り下げられる", async () => {
    const { props, rerender } = renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");

    mockState.permissions = { ...mockState.permissions, canEdit: false };
    // InterviewWeightPanel は memo 化されているため同一 props の rerender では再評価
    // されない。実運用では権限変化が useAuth context 経由で memo を貫いて再描画するが、
    // context 非依存の mock にはその経路が無い。無害な prop 変化で親の再描画を
    // モデル化する（ゲート要件自体は緩めない）。
    rerender(<InterviewWeightPanel {...props} historyLoading />);

    expect(screen.queryByRole("button", { name: "体重を記録" })).not.toBeInTheDocument();
    expect(screen.queryByLabelText("体重")).not.toBeInTheDocument();
  });
});

describe("InterviewWeightPanel — 現在値 query の状態分離（C1, EMR-356 repair1）", () => {
  it("現在値 GET が loading の間は「体重の記録はありません」を出さない", () => {
    mockState.vitals = { data: undefined, isLoading: true, isError: false };

    const { container } = renderPanel({ pastMeasurements: [] });

    expect(container).not.toHaveTextContent("体重の記録はありません");
    expect(container).toHaveTextContent(/読み込み中/);
  });

  it("現在値 GET が error の間は「体重の記録はありません」を出さずエラーを表示する", () => {
    mockState.vitals = { data: undefined, isLoading: false, isError: true };

    const { container } = renderPanel({ pastMeasurements: [] });

    expect(container).not.toHaveTextContent("体重の記録はありません");
    expect(container).toHaveTextContent(/取得できませんでした/);
  });
});

describe("InterviewWeightPanel — 過去投影の 不在/空/部分 の区別（C2, EMR-356 repair1）", () => {
  it("過去投影が未定義（旧API/未提供）のときは「記録なし」と断定しない", () => {
    const { container } = renderPanel({ pastMeasurements: undefined });

    expect(container).not.toHaveTextContent("体重の記録はありません");
    expect(container).not.toHaveTextContent("体重の履歴を取得できませんでした");
  });

  it("履歴エラーでも取得済みの部分測定は表示したままにする（trimming成功/医療失敗相当）", () => {
    const { container } = renderPanel({
      pastMeasurements: PAST_MEASUREMENTS,
      historyError: true,
    });

    expect(container).toHaveTextContent("4.2");
    expect(container).toHaveTextContent("体重の履歴を取得できませんでした");
    expect(container).not.toHaveTextContent("体重の記録はありません");
  });
});

describe("InterviewWeightPanel — 親ロック/確定の書き込みゲート（R2, EMR-356 repair1）", () => {
  it("isLocked コミット後は記録 control が inert で、捕捉済みの click/Enter 経路も mutation しない", async () => {
    const { props, rerender } = renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");

    // layout commit 済みのロック。fieldset の視覚 disabled は form 制御を止めるだけで
    // React のイベントハンドラ（キャプチャ済みコールバック）までは止めないため、
    // ハンドラ側でも最新ゲートを再検査する必要がある。
    rerender(<InterviewWeightPanel {...props} isLocked />);

    const lockedInput = screen.getByLabelText("体重");
    const lockedButton = screen.getByRole("button", { name: "体重を記録" });
    expect(lockedInput).toBeDisabled();
    expect(lockedButton).toBeDisabled();

    // jsdom の fireEvent は disabled 要素でもハンドラへ到達する —— これが
    // 「コミット前にキャプチャされた操作経路」の近似であり、mutation を発行してはいけない。
    await act(async () => {
      fireEvent.keyDown(lockedInput, { key: "Enter" });
      fireEvent.click(lockedButton);
    });
    expect(createVitalMock).not.toHaveBeenCalled();
  });

  it("isFinalized 確定済みでも記録 control が inert で mutation しない", async () => {
    renderPanel({ isFinalized: true });

    const input = screen.getByLabelText("体重");
    const button = screen.getByRole("button", { name: "体重を記録" });
    expect(input).toBeDisabled();
    expect(button).toBeDisabled();

    await act(async () => {
      fireEvent.keyDown(input, { key: "Enter" });
      fireEvent.click(button);
    });
    expect(createVitalMock).not.toHaveBeenCalled();
  });

  it("ロックコミット前に開始した保存は1度きりで、ロック後の完了応答は新たな mutation を起こさない", async () => {
    let resolveSave: (value: unknown) => void = () => {};
    createVitalMock.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveSave = resolve;
        }),
    );
    const { props, rerender } = renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");
    await user.click(screen.getByRole("button", { name: "体重を記録" }));
    expect(createVitalMock).toHaveBeenCalledTimes(1);

    rerender(<InterviewWeightPanel {...props} isLocked />);
    await act(async () => {
      resolveSave({});
    });

    await waitFor(() => {
      expect(createVitalMock).toHaveBeenCalledTimes(1);
    });
  });
});

describe("InterviewWeightPanel — Action/フォーム関連付け（EMR-356 repair1）", () => {
  it("体重入力と記録ボタンの form owner は親カルテ form ではない（外部関連付け）", () => {
    const onSubmit = vi.fn((e: FormEvent) => e.preventDefault());
    render(
      <form data-testid="chart-form" onSubmit={onSubmit}>
        <InterviewWeightPanel medicalRecordId="10" recordClinicId="7" />
      </form>,
    );

    const chartForm = screen.getByTestId("chart-form");
    const input = screen.getByLabelText("体重") as HTMLInputElement;
    const saveButton = screen.getByRole("button", {
      name: "体重を記録",
    }) as HTMLButtonElement;

    // React19 action + 既存 SubmitButton（type="submit"）を親 form 内で安全に使うため、
    // control の form owner をネストしない別 form へ関連付ける設計が必須。
    // owner が親だと type=submit の click / Enter がカルテ保存を発火してしまう。
    expect(input.form).not.toBeNull();
    expect(input.form).not.toBe(chartForm);
    expect(saveButton.form).not.toBeNull();
    expect(saveButton.form).not.toBe(chartForm);
    expect(chartForm.contains(input.form as Node)).toBe(false);
  });
});

describe("InterviewWeightPanel — ARIA/ライフサイクル（R1/R3, EMR-356 repair1）", () => {
  it("複数パネル同居時に label/input/error の id が一意", () => {
    render(
      <>
        <InterviewWeightPanel medicalRecordId="10" recordClinicId="7" />
        <InterviewWeightPanel medicalRecordId="20" recordClinicId="7" />
      </>,
    );

    const inputs = screen.getAllByLabelText("体重") as HTMLInputElement[];
    expect(inputs).toHaveLength(2);
    expect(inputs[0].id).not.toBe("");
    expect(inputs[0].id).not.toBe(inputs[1].id);
  });

  it("保存完了は aria-live の status で通知される", async () => {
    renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");
    await user.click(screen.getByRole("button", { name: "体重を記録" }));

    const status = await screen.findByRole("status");
    expect(status).toHaveTextContent(/記録|保存/);
  });

  it("保存応答前にアンマウントされても late completion が例外を起こさない", async () => {
    let resolveSave: (value: unknown) => void = () => {};
    createVitalMock.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveSave = resolve;
        }),
    );
    const { unmount } = renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");
    await user.click(screen.getByRole("button", { name: "体重を記録" }));

    unmount();
    await act(async () => {
      resolveSave({});
    });
    await act(async () => {});

    expect(createVitalMock).toHaveBeenCalledTimes(1);
  });
});

describe("InterviewWeightPanel — 作成権限は記録ゲートに含めない（EMR-356 permission repair）", () => {
  it.each(["click", "Enter"] as const)(
    "canCreate 無しでも %s で体重を1回だけ記録し、親 submit を発火しない",
    async (submitVia) => {
      mockState.permissions = { ...mockState.permissions, canCreate: false };
      const onSubmit = vi.fn((e: FormEvent) => e.preventDefault());
      render(
        <form onSubmit={onSubmit}>
          <InterviewWeightPanel medicalRecordId="10" recordClinicId="7" />
        </form>,
      );

      const user = userEvent.setup();
      const input = screen.getByLabelText("体重");
      await user.type(input, "4.5");
      if (submitVia === "click") {
        await user.click(screen.getByRole("button", { name: "体重を記録" }));
      } else {
        await user.type(input, "{Enter}");
      }

      expect(await screen.findByText("体重を記録しました")).toBeInTheDocument();
      expect(createVitalMock).toHaveBeenCalledTimes(1);
      const input_arg = createVitalMock.mock.calls[0][0] as {
        recorded_at?: string;
        weight?: number | null;
        weight_unit?: string;
        temperature?: number | null;
        heart_rate?: number | null;
        respiration_rate?: number | null;
      };
      expect(input_arg).toMatchObject({ weight: 4.5, weight_unit: "Kg" });
      expect(typeof input_arg.recorded_at).toBe("string");
      expect(input_arg.temperature ?? null).toBeNull();
      expect(input_arg.heart_rate ?? null).toBeNull();
      expect(input_arg.respiration_rate ?? null).toBeNull();
      expect(createVitalArgsMock).toHaveBeenCalledWith("10", "7");
      expect(getVitalsArgsMock).toHaveBeenCalledWith(
        "10",
        "7",
        expect.objectContaining({ enabled: true }),
      );
      expect(onSubmit).not.toHaveBeenCalled();
    },
  );

  it("入力途中で作成権限が剥奪されても記録導線は残り、そのまま1回保存できる", async () => {
    const { props, rerender } = renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");

    mockState.permissions = { ...mockState.permissions, canCreate: false };
    rerender(<InterviewWeightPanel {...props} historyLoading />);

    await user.click(screen.getByRole("button", { name: "体重を記録" }));

    expect(await screen.findByText("体重を記録しました")).toBeInTheDocument();
    expect(createVitalMock).toHaveBeenCalledTimes(1);
    expect(createVitalMock.mock.calls[0][0]).toMatchObject({ weight: 4.5, weight_unit: "Kg" });
  });

  it("編集権限剥奪コミット後は、捕捉済みの form Action を直接発火しても mutation しない", async () => {
    const { props, rerender } = renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");

    const captured = capturedAction.current;
    if (captured === undefined) {
      throw new Error("portal form の Action を捕捉できていない");
    }

    mockState.permissions = { ...mockState.permissions, canEdit: false };
    rerender(<InterviewWeightPanel {...props} historyLoading />);

    const formData = new FormData();
    formData.set("recordIdentity", "10|7");
    act(() => {
      startTransition(() => {
        captured(formData);
      });
    });
    await act(async () => {});

    expect(createVitalMock).not.toHaveBeenCalled();
  });

  it("canView 剥奪コミット後は、捕捉済みの form Action を直接発火しても mutation しない", async () => {
    const { props, rerender } = renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");

    const captured = capturedAction.current;
    if (captured === undefined) {
      throw new Error("portal form の Action を捕捉できていない");
    }

    mockState.permissions = { ...mockState.permissions, canView: false };
    rerender(<InterviewWeightPanel {...props} historyLoading />);
    expect(getVitalsArgsMock).toHaveBeenCalledWith(
      "10",
      "7",
      expect.objectContaining({ enabled: false }),
    );

    const formData = new FormData();
    formData.set("recordIdentity", "10|7");
    act(() => {
      startTransition(() => {
        captured(formData);
      });
    });
    await act(async () => {});

    expect(createVitalMock).not.toHaveBeenCalled();
  });

  it("canCreate 無しの記録失敗はエラーを表示して保存済みにせず、再試行で成功する", async () => {
    mockState.permissions = { ...mockState.permissions, canCreate: false };
    createVitalMock.mockRejectedValueOnce(new Error("forced failure"));
    renderPanel();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");
    await user.click(screen.getByRole("button", { name: "体重を記録" }));

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(screen.queryByText("体重を記録しました")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "体重を記録" }));

    expect(await screen.findByText("体重を記録しました")).toBeInTheDocument();
    expect(createVitalMock).toHaveBeenCalledTimes(2);
    for (const call of createVitalMock.mock.calls) {
      expect(call[0]).toMatchObject({ weight: 4.5, weight_unit: "Kg" });
    }
  });
});

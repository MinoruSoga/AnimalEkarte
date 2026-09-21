# SLACK-BACKGROUND: 背景分の画面対応図（候補列挙・宛先未決）

状態: **コード照合 READY／対象欄 PO UNKNOWN／内容追加なし**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-BACKGROUND`（L184–188、索引 L420）。保持する現場条件:

- 出典 929–937、親 timestamp `1789539164.115569`。原文要約は「背景分に皮下点滴を追加」
- 同一親は COMPLAINT / BACKGROUND / VITALS に分離する（[SLACK-INTAKE](../todo-campaign-20260919-ready17/SLACK-INTAKE.md) L105、todo-issue L484）。主訴任意・バイタル表示と混ぜない
- 対象画面・欄・臨床的用途は特定できておらず、PO 確認待ち
- 「背景分」を定型文、既往歴、治療行のいずれとも推定しない
- 特定前は内容追加を停止し、用量・方法・臨床文面を作らない
- 既存項目で目的を満たす場合は重複登録せず案内する（既存項目の判定も PO）

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-BACKGROUND`（`remaining-ops-20260920` revision 1）の owned path および人間が読む対応図である。照合 revision `873685b0bea3692c2f8100b19ded660c8357f2b0`（`feat/rem-slack-background-20260920`）。本票は調査のみ。製品コード・テスト・マスタ行は変更しない。

呼び出し行: **無い。** 既存 `docs/work/todo-campaign-20260919-ready17/` に本 unit の票は無く、`todo-issue.md` L184–188 は出典要約であり本票の代替ではない。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

## 実践ゲート（実装しない理由）

[product-philosophy.md](../../product-philosophy.md) の 5 ステップをこの要望に当てると、①要件を疑う段階で止まる。

1. **要件を疑う:** 「背景分に皮下点滴を追加」は画面要望であり、対象欄・目的・マスタか自由文か・実施/請求との関係が未裁定。責任者個人名が無い。
2. **削除:** 既存の定型文・所見・治療行のどれかで足りるなら新規欄も新規マスタも作らない。足りるかは PO が決める。本票で既存項目を宛先に選ばない。
3. **簡素化 / サイクル短縮 / 自動化:** 欄未決のまま文面や処置を足す最適化は禁止。

裁定前に 皮下点滴 の copy をいずれかの欄へ書くのは①違反。本票は候補の並置まで。

## 医院事実（コード外・UNKNOWN）

数値・院内ルール・臨床文面をコードから捏造しない。未採取なら該当セルは **UNKNOWN**。

| 項目 | 文書/コードで分かること | 医院事実 |
| --- | --- | --- |
| 「背景分」が指す画面ラベル | 現行 UI に「背景分」という見出しは無い（下記ギャップ） | **UNKNOWN**。PO が対象欄を決める |
| 追加したい「皮下点滴」の意味 | 原文要約のみ。処置実施、予定、自由文、定型文、請求対象の区別なし | **UNKNOWN**。用量・方法・文面を作らない |
| マスタか自由文か | 定型文は文字列置換、治療は `treatments` 行、所見は `clinical_plans` | **UNKNOWN** |
| 実施 / 請求との関係 | 治療行は確認後に未請求へ載り得る（[SLACK-PLAN-MANUAL](../todo-campaign-20260919-ready17/SLACK-PLAN-MANUAL.md)） | **UNKNOWN**。本票で請求経路を選ばない |
| 報告医院・端末 | なし | **UNKNOWN** |
| フロント/API revision | 本票作成時 worktree HEAD `873685b0bea3692c2f8100b19ded660c8357f2b0` | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけない読み替え

現場の「背景」「定型」「点滴」は次のどれでも同じ言葉になる。候補を列挙しても、宛先にはしない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| T — 定型文挿入 | [InterviewChiefComplaint.tsx](../../../frontend/src/features/medical-records/components/InterviewChiefComplaint.tsx) L106–124。ボタンは親の `INTERVIEW_TEMPLATES`（[MedicalRecordInterview.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordInterview.tsx) L24–32） | 「背景分」= 定型ボタンと推定しない。定型は **主訴詳細の全文置換**（MedicalRecordInterview L71–76） |
| C — 主訴詳細 | InterviewChiefComplaint L127–139 `id="medical-record-chief-complaint"`。問診 PATCH は `chief_complaint`（[use-medical-record-save-action.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) L236–246） | COMPLAINT（主訴任意）と混ぜない。主訴本文へ点滴文を足さない |
| N — 問診「治療方針」 | [InterviewTreatmentPolicy.tsx](../../../frontend/src/features/medical-records/components/InterviewTreatmentPolicy.tsx) L32–48。保存は inquiry `notes`（save-action L242–245）。hydrate は BUG-034（[use-medical-record-form.test.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-form.test.ts) L377–429） | ラベルが「治療方針」でも `clinical_plans.treatment_policy` ではない |
| P — 診察/治療プランの所見 | [ClinicalPlanSection.tsx](../../../frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx) L72–149。見出し「診察所見・診断・治療方針」。保存は `physical_exam` / `treatment_policy` / `diagnosis_details`（save-action L272–276） | 問診 notes と同一視しない。自由文へ点滴を足す宛先にもしない |
| R — 治療プラン表 | [MedicalRecordDiagnosisPlan.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx) L228–254 見出し「治療プラン」。`useGetTreatments` / `useCreateTreatment`（L13–19, L79–84） | 同名 backend `treatment_plans` をこの表の保存先とみなさない（PLAN-MANUAL） |
| U — 治療タブ | [MedicalRecordClinicalTabs.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordClinicalTabs.tsx) と [use-treatments-tab.ts](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts)。同一 `treatments` 集合の別 UI | プラン表と別テーブルではない。どちらかを「背景分」と決めない |
| M — 治療マスタ検索 | [TreatmentSearchDialog.tsx](../../../frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx) L44, L77–120。カテゴリ 診察/検査/処置/予防/入院/薬剤。選択は `item.name` を `content` に載せる（DiagnosisPlan L149–165） | リポジトリに「皮下点滴」マスタ名は無い（下記ギャップ）。無いことを「作る根拠」にしない |
| H — 問診テンプレマスタ category `history` | [InterviewTemplateSettings.tsx](../../../frontend/src/features/master/routes/InterviewTemplateSettings.tsx) L35–41。`history: "既往歴"`。CRUD は `/settings/inquiry-templates` と `/settings/interview/templates`（[settings-routes.tsx](../../../frontend/src/app/routes/settings-routes.tsx) L314–360） | 設定画面の「既往歴」ラベルをカルテの「背景分」と同一視しない。問診タブの定型ボタンはこのマスタを読まない |
| I — 問診履歴 | [InterviewHistory.tsx](../../../frontend/src/features/medical-records/components/InterviewHistory.tsx) は過去詳細への Link。空なら DEFAULT_HISTORY_ITEMS（MedicalRecordInterview L34–58, L78–79） | 履歴リンク・モック行を背景欄にしない |
| V — バイタル | 同一 Slack 親の VITALS | 背景分の欄特定とバイタル表示位置を混ぜない |

## 候補サーフェス（宛先を選ばない）

列の定義:

- **画面:** ユーザーが見るタブ / 見出し
- **入力:** 操作できるコントロール
- **永続化:** 保存先（コード）
- **本票の扱い:** 候補として列挙するだけ。採用しない

| ID | 画面 | 入力 | 永続化 | 本票の扱い |
| --- | --- | --- | --- | --- |
| S1 | カルテ「問診」→ 主訴情報 → **定型文挿入** | ハードコード 4 ボタン（定期検診 / ワクチン / 下痢・嘔吐 / 皮膚） | クリックは `setChiefComplaint(text)`。問診保存時 `inquiries.chief_complaint` | 候補。背景分でも 皮下点滴 定型でもない |
| S2 | 同左 → **主訴詳細** | textarea | `inquiries.chief_complaint` | 候補。COMPLAINT の対象であり BACKGROUND の宛先ではない |
| S3 | 問診 → **治療方針**（次工程へ連携） | textarea | `inquiries.notes` | 候補。clinical_plan の治療方針と二重 |
| S4 | 設定 → 問診テンプレート（`history` = 表示ラベル「既往歴」） | マスタ CRUD。category / title / content | `inquiry_templates`（[InquiryTemplate](../../../frontend/src/types/generated/models.ts) L1772–1784） | 候補。カルテ問診の定型ボタンは未配線 |
| S5 | カルテ「診察/治療プラン」→ 身体検査所見 / 診断詳細 / 治療方針 | 3 textarea | `clinical_plans.physical_exam` / `diagnosis_details` / `treatment_policy` | 候補。自由文。処置実施でも請求でもない |
| S6 | 同タブ → 見出し **治療プラン** の表 | マスタ検索ダイアログ、またはコード上の空行 `handleAddRow`（クリックは検索優先。PLAN-MANUAL） | `treatments`（`content` / `unit_price` / `item_type`） | 候補。請求候補になり得る。背景文ではない |
| S7 | カルテ「治療」タブ | 同一 `treatments` + 薬剤投与量ゲート | 同じ `treatments` | 候補。S6 と別欄ではない |

**採用した宛先: 無し。** 上表のどの ID も BACKGROUND の入力欄ではない。PO 意味は **UNKNOWN**。

## コード根拠（引用）

### 1. 問診定型文は主訴の置換であり、マスタ `history` ではない

`MedicalRecordInterview` の定型はモジュール定数。ラベルに「背景」「点滴」は無い。

```24:32:frontend/src/features/medical-records/components/MedicalRecordInterview.tsx
const INTERVIEW_TEMPLATES: { label: string; text: string }[] = [
  { label: "定期検診", text: "# 定期検診\n特に異常なし。食欲・元気あり。" },
  { label: "ワクチン", text: "# 混合ワクチン接種\n体調良好。" },
  {
    label: "下痢・嘔吐",
    text: "# 消化器症状\n・嘔吐：あり（回数：　）\n・下痢：あり（性状：　）\n・食欲：なし",
  },
  { label: "皮膚", text: "# 皮膚症状\n・痒み：あり\n・発赤：あり\n・部位：" },
];
```

挿入は追記ではなく全文置換。

```71:76:frontend/src/features/medical-records/components/MedicalRecordInterview.tsx
  const handleInsertTemplate = useCallback(
    (text: string) => {
      setChiefComplaint(text);
    },
    [setChiefComplaint],
  );
```

UI 見出しは「定型文挿入」。マスタ編集リンクは `/settings/interview/templates`（InterviewChiefComplaint L106–124、paths.ts L276–278）。そのルートは `InterviewTemplateSettings` を lazy load する（settings-routes L348–360）が、**問診タブのボタン配列は `useGetInquiryTemplates` を呼ばない**（当該 hook の利用は設定画面とそのテストのみ）。

設定マスタの category 表示に「既往歴」がある。

```35:41:frontend/src/features/master/routes/InterviewTemplateSettings.tsx
const INQUIRY_CATEGORY_LABELS: Record<string, string> = {
  chief_complaint: "主訴",
  history: "既往歴",
  current_medications: "現在の投薬",
  notes: "メモ/備考",
};
```

これは設定 UI のラベル写像である。カルテに「既往歴」欄を出す根拠にも、「背景分」の同義にもしない。

### 2. 診察/治療プランは所見テキストと treatments 表が同居する

タブ登録:

```66:90:frontend/src/features/medical-records/components/MedicalRecordClinicalTabs.tsx
      <MedicalRecordMountedTab
        tab="診察/治療プラン"
        activeTab={activeTab}
        mountedTabs={mountedTabs}
      >
        <MedicalRecordDiagnosisPlan
          isNewRecord={isNewRecord}
          chiefComplaint={chiefComplaint}
          physicalExam={physicalExam}
          setPhysicalExam={onPhysicalExamChange}
          plan={plan}
          setPlan={onPlanChange}
```

治療行は treatments API。空行 create とマスタ選択は別コールバック。マスタ選択は `item.name` を `content` にコピーする。clinic マスタの実データ名は本票のローカル照合対象外（Docker/DB 未実行）。リポジトリ文字列としての「皮下点滴」は `todo-issue.md` 以外に無い。

```135:165:frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx
  const handleAddRow = useCallback(() => {
    if (!canCreate) return;
    createTreatmentFn({
      item_type: "other",
      content: "",
      unit_price: 0,
      quantity: 1,
      is_selected: true,
      is_insurance: false,
      discount_amount: 0,
      sort_order: nextOrder,
    });
  }, [canCreate, nextOrder, createTreatmentFn]);

  const handleSelectTreatment = useCallback(
    (item: TreatmentMasterItem) => {
      if (!canCreate) return;
      createTreatmentFn({
        item_type: resolveItemTypeFromCategory(item.category),
        content: item.name,
        memo: item.category,
        unit_price: item.unitPrice,
        quantity: 1,
        is_selected: true,
        is_insurance: true,
        discount_amount: 0,
        sort_order: nextOrder,
      });
    },
    [canCreate, nextOrder, createTreatmentFn],
  );
```

マスタダイアログのカテゴリ集合（入院チップは CATEGORY_ORDER にあるが、items 組み立ては診察/処置/予防/検査/薬剤）:

```44:44:frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx
const CATEGORY_ORDER = ["診察", "検査", "処置", "予防", "入院", "薬剤"];
```

`resolveItemTypeFromCategory` は 薬剤/処置/診察以外を `"other"` にする（[treatments-tab-model.ts](../../../frontend/src/features/medical-records/lib/treatments-tab-model.ts) L5–10）。カテゴリ推定で「背景分」を作らない。

### 3. 問診の治療方針とプランの治療方針は別レコード

問診タブ保存:

```236:246:frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts
            await updateInquiryMutation.mutateAsync({
              chief_complaint:
                snapshot.chiefComplaint !== snapshot.chiefComplaintDefault
                  ? snapshot.chiefComplaint
                  : undefined,
              chief_complaint_type_id: snapshot.chiefComplaintTypeId,
              notes:
                snapshot.treatmentPolicy !== snapshot.treatmentPolicyDefault
                  ? snapshot.treatmentPolicy
                  : undefined,
            });
```

診察/治療プランタブ保存:

```272:276:frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts
            const treatmentPlanPayload = {
              physical_exam: snapshot.physicalExam,
              treatment_policy: snapshot.plan,
              diagnosis_details: snapshot.assessment,
```

同一ラベル「治療方針」が 2 系統あること自体が、未裁定の「背景分」をどちらかへ割り当ててはならない理由になる。

## ギャップ

| 探したもの | 結果 | 含意 |
| --- | --- | --- |
| UI 文字列「背景分」 | フロントソースに無し（本票作成時 `rg 背景分` は todo-issue / 本票のみ） | 既存ラベルへの 1:1 対応はできない |
| リポジトリ文字列「皮下点滴」 | `todo-issue.md` L186 のみ | 既存マスタ名・定型文面としてコードに無い。clinic DB のマスタ名は **未実行** であり、本票で有無を断定しない |
| 問診定型ボタン ← inquiry_templates | 未配線。ハードコード 4 件 | 設定の「既往歴」テンプレをカルテへ足しても、現状の定型ボタンは変わらない |
| 「背景分」専用カラム | inquiries / clinical_plans / treatments に該当名なし | 新カラムは本票の範囲外。削除候補ですら未裁定 |

## 停止条件

次が揃うまで製品変更・マスタ追加・文面作成をしない。

1. 依頼者/臨床責任者が **対象欄**（上表 S1–S7 のいずれか、別画面、または「既存で足りるので追加しない」）を決める
2. 目的（記録 / 実施 / 請求 / 定型の再利用）と、マスタか自由文かを決める
3. 実施/請求と治療行の関係を決める。未決のまま treatments へ行を足さない
4. 既存項目で足りるなら重複登録しない
5. COMPLAINT・VITALS・PLAN-MANUAL・COPY の票へ BACKGROUND を吸収しない

## 本票がやらないこと

- 皮下点滴の copy、用量、経路、頻度を書かない
- 定型文配列や treatment master へ項目を足さない
- 「一番近そう」な欄へ仮決めしない
- Docker / 共有 DB / 実医院マスタの検索をしない（ローカルファイルのみ）

/**
 * カルテ編集画面の右下に固定される操作ボタン群（バイタル記録・印刷・確定する・保存。
 * MedicalRecordFormActions.tsx と MedicalRecordBillCheck.tsx の 2 か所が描画する）が
 * スクロール末尾の入力・表を覆わないように、スクロール内容の最外側へ付与する下余白。
 *
 * ボタン群は `fixed bottom-6 right-20` の h-10 ボタンで、画面下端から
 * bottomOffsetPx(24px) + buttonHeightPx(40px) = footprintPx(64px) を占有する。
 * 余白は footprintPx に余裕(slackPx)を足した 96px（pb-24）を要求する。
 * （保存は共有 SubmitButton の既定 min-h-11 = 44px だが、実際の占有 68px でも 96px の
 * 下余白は上回る。）
 *
 * PageLayout の既定余白（SUPPORT_WIDGET_LAYOUT.scrollBottomClearance = pb-20）は
 * 他画面と共有する値のため変更せず、カルテ編集ページ側でこの値を使う。
 * 変更時は fixed-action-layout.test.ts の算術検証を更新する。
 */
export const MEDICAL_RECORD_FIXED_ACTION_LAYOUT = {
  /** ボタン群の固定位置クラス（2 か所の描画で共有。見た目は変更しない） */
  position: "fixed bottom-6 right-20 z-50",
  /** ボタン 1 つの高さクラス（h-10 = 40px）。変更時は buttonHeightPx と揃える */
  buttonHeightClass: "h-10",
  /** ボタン 1 つの高さ（h-10 = 40px） */
  buttonHeightPx: 40,
  /** 画面下端からの距離（bottom-6 = 24px） */
  bottomOffsetPx: 24,
  /** ボタン群が画面下端から占有する高さ = bottomOffsetPx + buttonHeightPx */
  footprintPx: 64,
  /** footprintPx に足す余裕 */
  slackPx: 32,
  /** スクロール内容の最外側に付ける下余白（footprintPx + slackPx = 96px） */
  scrollBottomClearance: "pb-24",
  scrollBottomClearancePx: 96,
} as const;

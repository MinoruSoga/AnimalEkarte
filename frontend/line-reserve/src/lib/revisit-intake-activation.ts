// EMR-474: LIFF 再診問診の利用開始 flag（静的。環境変数・サーバー応答・localStorage では切り替えない）。
//
// revisit-intake の 3 route（GET / POST submissions / GET operations）は api.yaml に selected として
// 公開されているが、BE は未実装（technical-spec: 「API文書公開と機能利用開始を分け」）。
// false の間は、予約一覧・完了画面の問診入口、問診ページの mount、問診 API の呼び出しをすべて止める。
// 通常の予約作成・取消は変えない。
//
// true への変更は別 PR で行う。条件: 必要な BE 3 route と認可・receipt 処理が揃い、配信先でも利用可能になった後。
// 有効化後も、pending 予約の入口・server の editable・履歴・結果照会の規則は維持する。
// 公開後の障害時スイッチとしては流用しない（障害時の停止手段は別に設計する）。
export const LIFF_REVISIT_INTAKE_ACTIVE = false;

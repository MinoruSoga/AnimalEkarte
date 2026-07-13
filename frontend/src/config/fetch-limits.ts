// 履歴系フェッチ（トリミング/診療/lstepタグ紐付け飼主など）の取得上限。
// FE3-8: 5 サイトに散在していた同一セマンティクスの直値 100 を集約。値は既存直値のまま。
export const HISTORY_FETCH_LIMIT = 100;

// 予約/受付の日別ビューを1リクエストで取り切るための上限（BUG #82 の経緯は use-get-reservations 参照）
export const DAY_VIEW_FETCH_LIMIT = 100;
// 選択肢マスタ（診断オプション等）の取得上限
export const OPTIONS_FETCH_LIMIT = 100;

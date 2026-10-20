/**
 * EMR-463: 飼主・ペット一覧を「現在許可医院（GET /me `owners_view_clinic_ids`）から開始する経路」へ
 * 接続するかどうかの、owners feature 内で唯一の明示 activation。
 *
 * - false の理由: 463-BE の GetMeResponse（`owners_view_clinic_ids` を返す GET /me 専用 wrapper）が
 *   まだ実装されていない（api.yaml `x-contract-status: selected-not-implemented`）。未提供 field の
 *   取得経路へ本番一覧を切り替えると、全利用者で一覧が「scope 取得失敗」になる。
 *   463-BE 実装後の後続 PR でこの定数を true へ切り替えて接続する。
 * - これはビルド時に決まる静的な activation であり、API 応答の有無や取得失敗を見て
 *   経路を選ぶ fallback ではない。活性時の経路は fail-closed（所属全件・現在医院・旧 scope へ
 *   戻らない）のまま変えない。共通 feature flag 基盤・env・/me 応答の有無では判定しない。
 * - false の間、ownersLoader / OwnersList は PR 前（main 05da02f3）と同じ挙動になる。
 */
export const OWNERS_CLINIC_SCOPE_ACTIVE = false as const;

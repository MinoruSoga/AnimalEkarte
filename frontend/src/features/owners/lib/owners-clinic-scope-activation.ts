/**
 * EMR-463: 飼主・ペット一覧を「現在許可医院（GET /me `owners_view_clinic_ids`）から開始する経路」へ
 * 接続するかどうかの、owners feature 内で唯一の明示 activation。
 *
 * - true の根拠: 463-BE の GetMeResponse（`owners_view_clinic_ids` を返す GET /me 専用 wrapper）が
 *   実装済み（api.yaml `x-contract-status: selected`）。loader/route の既存挙動は PR #754 で
 *   契約 mock により実装済みであり、この定数が本番経路への接続点である。
 * - これはビルド時に決まる静的な activation であり、API 応答の有無や取得失敗を見て
 *   経路を選ぶ fallback ではない。活性時の経路は fail-closed（所属全件・現在医院・旧 scope へ
 *   戻らない）のまま変えない。共通 feature flag 基盤・env・/me 応答の有無では判定しない。
 */
export const OWNERS_CLINIC_SCOPE_ACTIVE = true as const;

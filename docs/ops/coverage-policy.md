# Coverage Policy — AnimalEkarte

> Coverage 設定の実装正本は `.github/workflows/ci.yml`、`backend/.coverage-baseline`、`frontend/.coverage-baseline`、`frontend/vite.config.ts`。

## 計測と除外

### Backend

- 全 shard が `-coverpkg=./internal/...` を使う。`internal/infra` も現在の母集団に含まれる。
- `cmd/*`、`lstep-migrate/`、`seed-old-db/`、`migrations/` はこの母集団の外にある。
- profile の単純連結は禁止する。`scripts/merge_go_coverprofiles.py` が mode/statement 数を検証し、同一 block を統合する。

### Frontend

Vitest の default exclusions に加え、`src/types/generated/**` と `src/testing/**` を除外する。

## Current ratchet baseline

| 対象 | baseline | 根拠 |
|---|---:|---|
| Backend | **87.3%** | `backend/.coverage-baseline`。2026-08-23、CI run 32639276691 の実測で re-arm |
| Frontend statements | **43.78%** | `frontend/.coverage-baseline` |

baseline file のコメントが値、run、変更理由の正本である。Backend の 91.3% は 2026-07-13 の履歴値であり、現行 threshold ではない。2026-08-14 の 88.1% と 2026-08-23 の 87.3% を含む後続履歴も backend baseline file を参照する。推測値で更新しない。

## Phase status

- **Phase 0 — historical/completed**: artifact 可視化のみだった導入段階。
- **Phase 1 — implemented and armed**: backend/frontend とも baseline から tolerance（既定 0.5pp）を超える低下を CI で fail する。対象は path-filter に該当する main push と main/staging/production PR。
- Frontend script の OK/FAIL/baseline 未記録/`--warn-only` の四ケースは一時的な履歴検証として記録されていたが、HEAD に保守される regression test/fixture はない。継続保証とは扱わない。
- **Phase 2 — proposal, not implemented at reviewed baseline `70dc7405`; current checkout must be revalidated**: patch coverage warning。`octocov` / `diff-cover` は未導入。
- **Phase 3 — proposal, not implemented at reviewed baseline `70dc7405`; current checkout must be revalidated**: domain/capability baseline と manifest による gate。manifest は未導入。

## Baseline 更新

1. 対象 CI artifact の summary を取得する。
2. 実測値と run/commit、変更理由を review する。
3. 対応する `.coverage-baseline` へ実測値を転記する。
4. workflow の対象 branch/path と ratchet step が実行されたことを確認する。

外部 Actions run の保持・到達性はリポジトリから保証できない。必要な証跡は人が確認する。

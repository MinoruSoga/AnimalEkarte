# CodeGraph 定番プロンプト集（AnimalEkarte 向け）

このドキュメントは、Claude Code + CodeGraph MCP でこのリポジトリを調査するときの定番プロンプト集です。
そのまま貼って使える形にしています。

---

## 1. ルートから実装までを一気に追跡

```text
CodeGraphで `/api/v1/clinic-holidays` の実装を、route -> handler -> service -> model の順で辿って。
各レイヤーごとに「ファイル名」「関数名」「責務」を1行でまとめて。
```

## 2. 変更影響範囲の洗い出し（バックエンド）

```text
CodeGraphで `SetClinicHoliday` を変更した場合の影響範囲を出して。
callers / callees / 関連model / 関連APIエンドポイントを分けて一覧化して。
```

## 3. 変更影響範囲の洗い出し（フロントエンド）

```text
CodeGraphで `frontend/src/features/shifts/api/clinic-holidays.ts` に依存しているコンポーネントを列挙して。
画面単位で「どのhook/関数経由で呼んでいるか」も出して。
```

## 4. OpenAPI と実装の整合チェック

```text
CodeGraphで `docs/openapi.yaml` の `/api/v1/clinic-holidays` と Go実装の差分を確認して。
method/path/認証/主要request/response の不一致だけを、根拠ファイル+行番号つきで出して。
```

## 5. ERD と migration の整合チェック

```text
CodeGraphで `docs/ERD.md` と `backend/migrations/001_init.sql` のテーブル整合を確認して。
ERDにあるがmigrationにない、migrationにあるがERDにない、を分けて報告して。
```

## 6. 認証・認可フローの可視化

```text
CodeGraphでログイン後の認証フローを追跡して。
`auth_handler.go` と `middleware/auth.go` を起点に、Cookie発行・検証・X-Clinic-ID適用までを時系列で説明して。
```

## 7. バグ調査の初動テンプレ

```text
CodeGraphで「`/api/v1/owners/{id}` 更新時に 403 になる」原因候補を調べて。
権限制御の分岐点を3つまで挙げて、各候補の根拠コードを提示して。
```

## 8. リファクタ前の安全確認

```text
CodeGraphで `backend/internal/service/owner_service.go` の public 関数ごとに callers を出して。
外部契約（handlerが依存している戻り値/エラー型）を壊すと影響する箇所を優先順で並べて。
```

## 9. 未使用コードの候補抽出

```text
CodeGraphで `backend/internal/handler` 配下の関数から、route未登録かつcallerがない候補を抽出して。
誤検知を避けるため、テスト参照のみのケースは別枠に分けて。
```

## 10. レビュー前の差分理解テンプレ

```text
CodeGraphで「このブランチの変更ファイル」が既存構造に与える影響を要約して。
`API互換性`, `DB影響`, `Frontend影響`, `運用ドキュメント影響` の4区分で短く出して。
```

---

## 使い分けの目安

- 実装経路を知りたい: `1`, `6`
- 変更前に事故を防ぎたい: `2`, `3`, `8`, `10`
- 仕様整合を確認したい: `4`, `5`
- 不具合/負債を潰したい: `7`, `9`

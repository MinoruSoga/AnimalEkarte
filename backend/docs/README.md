# Animal Ekarte API ドキュメント

## ファイル構成

```
docs/
├── README.md        # このファイル
└── api.yaml         # OpenAPI 3.0仕様（手動メンテナンス）
```

## api.yaml

手動で管理するOpenAPI 3.0仕様ファイル。エンドポイントの追加・変更時は合わせて更新する。

## 型生成（tygo）

Goモデルからフロントエンド型を自動生成するパイプライン。

```
backend/internal/model/*.go
    ↓ make codegen
frontend/src/types/generated/models.ts  ← 直接編集禁止
```

モデル変更時の手順:

1. `backend/internal/model/*.go` を編集
2. `make codegen` 実行
3. フロントエンドの型エラーを修正

---
description: CLAUDE.md ファイルを実装に合わせて同期（hooks テーブル再生成 + 死亡パス検出）
argument-hint: [--hooks | --refs | --all (default)]
---

# /sync-claude-md

CLAUDE.md ファイルの実装参照セクションをスキャンし、現在のコードベースに合わせて更新する。

**入力**: $ARGUMENTS（省略時は `--all`）

---

## 対象と判断基準

| 対象 | 動的セクション | 変化トリガー |
|------|-------------|------------|
| `frontend/src/hooks/CLAUDE.md` | フックリスト・importer 数テーブル | hooks 追加/削除、参照元変更 |
| `.claude/CLAUDE.md` | refs テーブルのファイルパス参照 | ファイル移動/削除 |
| その他 CLAUDE.md | Go/Gin正本・局所ガイドへの参照 | 正本の移動、directory guidance変更 |

---

## Step 1: 対象を決定

`$ARGUMENTS` に応じて実行範囲を決める:
- `--hooks` → Step 2 のみ
- `--refs` → Step 3 のみ
- `--all` または省略 → Step 2 + Step 3

---

## Step 2: hooks/CLAUDE.md テーブル再生成

### 2-1. 現在の hooks ファイル一覧を取得

```bash
find frontend/src/hooks -maxdepth 1 -name "use-*.ts" ! -name "*.test.ts" | sort
```

### 2-2. 各フックの importer 数をカウント

各ファイル `use-xxx.ts` に対して:

```bash
grep -r "from ['\"]@/hooks/use-xxx['\"]" frontend/src/ --include="*.ts" --include="*.tsx" | wc -l
```

### 2-3. CLAUDE.md テーブルとの差分を特定

現在の `frontend/src/hooks/CLAUDE.md` のテーブルと実態を比較し、3種類の差分を検出する:

**削除**: CLAUDE.md に記載されているがファイルが存在しない → テーブルから行を削除

**カウント不一致**: importer 数が記載値と異なる → 数値のみ更新（分類・description は変更しない）

**追加**: CLAUDE.md に未記載のフックファイルが存在する → 以下の手順でテーブルに追加:
1. フックファイルを Read して内容を確認
2. `useQuery` / `useMutation` を含む → **Cross-feature データ系** テーブルに追加
3. 含まない → **ユーティリティ系** テーブルに追加
4. description 列はフックファイルの JSDoc・コメント・実装から用途を推測して記入する

### 2-4. CLAUDE.md を更新

差分がある場合のみ Edit ツールでテーブルを更新する。差分ゼロなら「✅ hooks/CLAUDE.md は最新」と報告。

---

## Step 3: .claude/CLAUDE.md の refs パス確認

### 3-1. CLAUDE.md からファイルパス参照を抽出

`.claude/CLAUDE.md` のテーブル・コードブロック内のファイルパスを読み取る（`refs/*.md`、`internal/handler/CLAUDE.md` 等）。

### 3-2. 各パスの存在確認

```bash
ls <path>  # ファイルが存在するか確認
```

### 3-3. 死亡パスを報告

存在しないパスを一覧表示する。自動修正はしない（リネームか削除か判断できないため）。

---

## 出力フォーマット

```
## sync-claude-md 結果

### hooks/CLAUDE.md
- 追加: use-xxx.ts（importer: N）
- 削除: use-yyy.ts（ファイルなし）
- 更新: use-zzz.ts  9 → 6

### .claude/CLAUDE.md refs
- ✅ すべてのパスが有効
  または
- ⚠️ 死亡パス: refs/xxx.md（ファイルなし）

### 変更なし
- backend/CLAUDE.md, handler/CLAUDE.md, ... （静的ルールのため対象外）
```

---

## 注意事項

- **自動修正するのは hooks/CLAUDE.md のテーブルのみ**（機械的に再生成可能なため）
- **refs パスの修正は報告のみ**（人間が意図を判断する必要があるため）
- 差分ゼロの場合でも必ず完了報告を出力する

import { useState, type ReactNode } from "react";
import { toast } from "sonner";
import { BADGE, C, PALETTE, STYLE } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import {
  useGetAutoManagedPrefixes,
  useCreateAutoManagedPrefix,
  useDeleteAutoManagedPrefix,
  useGetConditionTagMappings,
  useCreateConditionTagMapping,
  useDeleteConditionTagMapping,
  useGetSendPurposeTagPrefixes,
  useCreateSendPurposeTagPrefix,
  useDeleteSendPurposeTagPrefix,
} from "../hooks/use-lstep-tag-config";
import type {
  AutoManagedPrefix,
  ConditionTagMapping,
  SendPurposeTagPrefix,
} from "../hooks/use-lstep-tag-config";

// ─────────────────────────────────────────────────
// DeleteButton
// ─────────────────────────────────────────────────

function DeleteButton({
  onDelete,
  disabled,
}: {
  onDelete: () => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onDelete}
      disabled={disabled}
      className={`min-h-11 min-w-11 text-xs ${C.text50} hover:opacity-70 shrink-0 disabled:opacity-30`}
    >
      削除
    </button>
  );
}

// ─────────────────────────────────────────────────
// TagConfigListSection (共通プレゼンテーション部分)
// ─────────────────────────────────────────────────
//
// 3セクション（AutoManagedPrefixes / ConditionTagMappings /
// SendPurposeTagPrefixes）で共通する「loading / 空 / 一覧」の描画のみを
// 汎用化したもの。データ取得・追加・削除は各セクションが自前のフック呼び出し
// として保持する（Rules of Hooks を守るためフックを props で渡さない）。

interface TagConfigListSectionProps<TItem> {
  items: TItem[] | undefined;
  isLoading: boolean;
  getId: (item: TItem) => string | number;
  renderRow: (item: TItem) => ReactNode;
  onDelete: (item: TItem) => void;
  deleteDisabled?: boolean;
}

function TagConfigListSection<TItem>({
  items,
  isLoading,
  getId,
  renderRow,
  onDelete,
  deleteDisabled,
}: TagConfigListSectionProps<TItem>) {
  if (isLoading) {
    return <div className={`text-sm ${C.text50}`}>読み込み中...</div>;
  }

  if (!items || items.length === 0) {
    return <div className={`text-sm ${C.text50} mb-3`}>登録なし</div>;
  }

  return (
    <div className="space-y-1 mb-3">
      {items.map((item) => (
        <div
          key={getId(item)}
          className="flex items-center gap-2 px-2 py-1.5 rounded border"
          style={{ borderColor: PALETTE.borderLight }}
        >
          {renderRow(item)}
          <DeleteButton
            onDelete={() => onDelete(item)}
            disabled={deleteDisabled}
          />
        </div>
      ))}
    </div>
  );
}

// ─────────────────────────────────────────────────
// AutoManagedPrefixesSection
// ─────────────────────────────────────────────────

function AutoManagedPrefixesSection() {
  const { data: items, isLoading } = useGetAutoManagedPrefixes();
  const createMutation = useCreateAutoManagedPrefix();
  const deleteMutation = useDeleteAutoManagedPrefix();

  const [prefix, setPrefix] = useState("");
  const [category, setCategory] = useState("");

  function handleAdd() {
    const trimmedPrefix = prefix.trim();
    const trimmedCategory = category.trim();
    if (!trimmedPrefix || !trimmedCategory) {
      toast.error("プレフィックスとカテゴリは必須です");
      return;
    }
    createMutation.mutate(
      { prefix: trimmedPrefix, category: trimmedCategory },
      {
        onSuccess: () => {
          setPrefix("");
          setCategory("");
          toast.success("追加しました");
        },
      },
    );
  }

  function handleDelete(item: AutoManagedPrefix) {
    deleteMutation.mutate(item.id, {
      onSuccess: () => toast.success("削除しました"),
    });
  }

  return (
    <div>
      <h4 className="text-sm font-medium mb-2" style={{ color: PALETTE.primary }}>
        自動管理プレフィックス (B/C系)
      </h4>
      <p className={`text-xs mb-3 ${C.text50}`}>
        このプレフィックスで始まるタグは自動管理対象となり、手動追加・削除が拒否されます。
      </p>

      <TagConfigListSection
        items={items}
        isLoading={isLoading}
        getId={(item) => item.id}
        renderRow={(item) => (
          <>
            <span className="text-sm font-mono flex-1">{item.prefix}</span>
            <span
              className={`text-xs px-1.5 py-0.5 rounded ${BADGE.blueNoBorder}`}
            >
              {item.category}
            </span>
            {item.description ? (
              <span className={`text-xs ${C.text50} flex-1`}>
                {item.description}
              </span>
            ) : null}
          </>
        )}
        onDelete={handleDelete}
        deleteDisabled={deleteMutation.isPending}
      />

      <div className="flex gap-2 items-end">
        <div className="flex flex-col gap-1">
          <label htmlFor="amp-prefix" className={`text-xs ${C.text50}`}>プレフィックス</label>
          <input
            id="amp-prefix"
            type="text"
            value={prefix}
            onChange={(e) => setPrefix(e.target.value)}
            placeholder="例: vaccine_"
            className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none w-36`}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="amp-category" className={`text-xs ${C.text50}`}>カテゴリ</label>
          <input
            id="amp-category"
            type="text"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            placeholder="例: C2"
            className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none w-24`}
          />
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={handleAdd}
          disabled={createMutation.isPending}
        >
          追加
        </Button>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// ConditionTagMappingsSection
// ─────────────────────────────────────────────────

function ConditionTagMappingsSection() {
  const { data: items, isLoading } = useGetConditionTagMappings();
  const createMutation = useCreateConditionTagMapping();
  const deleteMutation = useDeleteConditionTagMapping();

  const [conditionCode, setConditionCode] = useState("");
  const [tagName, setTagName] = useState("");

  function handleAdd() {
    const trimmedCode = conditionCode.trim();
    const trimmedTag = tagName.trim();
    if (!trimmedCode || !trimmedTag) {
      toast.error("疾患コードとタグ名は必須です");
      return;
    }
    createMutation.mutate(
      { condition_code: trimmedCode, tag_name: trimmedTag },
      {
        onSuccess: () => {
          setConditionCode("");
          setTagName("");
          toast.success("追加しました");
        },
      },
    );
  }

  function handleDelete(item: ConditionTagMapping) {
    deleteMutation.mutate(item.id, {
      onSuccess: () => toast.success("削除しました"),
    });
  }

  return (
    <div>
      <h4 className="text-sm font-medium mb-2" style={{ color: PALETTE.primary }}>
        慢性疾患コード → タグマッピング
      </h4>
      <p className={`text-xs mb-3 ${C.text50}`}>
        慢性疾患コードが記録された飼い主に付与するタグを設定します。
      </p>

      <TagConfigListSection
        items={items}
        isLoading={isLoading}
        getId={(item) => item.id}
        renderRow={(item) => (
          <>
            <span className="text-sm font-mono w-20">{item.condition_code}</span>
            <span className={`text-xs ${C.text50}`}>→</span>
            <span className="text-sm font-mono flex-1">{item.tag_name}</span>
          </>
        )}
        onDelete={handleDelete}
        deleteDisabled={deleteMutation.isPending}
      />

      <div className="flex gap-2 items-end">
        <div className="flex flex-col gap-1">
          <label htmlFor="ctm-condition-code" className={`text-xs ${C.text50}`}>疾患コード</label>
          <input
            id="ctm-condition-code"
            type="text"
            value={conditionCode}
            onChange={(e) => setConditionCode(e.target.value)}
            placeholder="例: DM"
            className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none w-24`}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="ctm-tag-name" className={`text-xs ${C.text50}`}>タグ名</label>
          <input
            id="ctm-tag-name"
            type="text"
            value={tagName}
            onChange={(e) => setTagName(e.target.value)}
            placeholder="例: CHRON_DM"
            className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none w-36`}
          />
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={handleAdd}
          disabled={createMutation.isPending}
        >
          追加
        </Button>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// SendPurposeTagPrefixesSection
// ─────────────────────────────────────────────────

function SendPurposeTagPrefixesSection() {
  const { data: items, isLoading } = useGetSendPurposeTagPrefixes();
  const createMutation = useCreateSendPurposeTagPrefix();
  const deleteMutation = useDeleteSendPurposeTagPrefix();

  const [purpose, setPurpose] = useState("");
  const [tagPrefix, setTagPrefix] = useState("");

  function handleAdd() {
    const trimmedPurpose = purpose.trim();
    const trimmedPrefix = tagPrefix.trim();
    if (!trimmedPurpose || !trimmedPrefix) {
      toast.error("送信目的とタグプレフィックスは必須です");
      return;
    }
    createMutation.mutate(
      { purpose: trimmedPurpose, tag_prefix: trimmedPrefix },
      {
        onSuccess: () => {
          setPurpose("");
          setTagPrefix("");
          toast.success("追加しました");
        },
      },
    );
  }

  function handleDelete(item: SendPurposeTagPrefix) {
    deleteMutation.mutate(item.id, {
      onSuccess: () => toast.success("削除しました"),
    });
  }

  return (
    <div>
      <h4 className="text-sm font-medium mb-2" style={{ color: PALETTE.primary }}>
        LINE送信目的 → タグプレフィックス
      </h4>
      <p className={`text-xs mb-3 ${C.text50}`}>
        LINE個別送信時に、送信目的に応じたタグプレフィックスで始まるタグのみ選択可能になります。
      </p>

      <TagConfigListSection
        items={items}
        isLoading={isLoading}
        getId={(item) => item.id}
        renderRow={(item) => (
          <>
            <span className="text-sm w-36">{item.purpose}</span>
            <span className={`text-xs ${C.text50}`}>→</span>
            <span className="text-sm font-mono flex-1">{item.tag_prefix}</span>
          </>
        )}
        onDelete={handleDelete}
        deleteDisabled={deleteMutation.isPending}
      />

      <div className="flex gap-2 items-end">
        <div className="flex flex-col gap-1">
          <label htmlFor="sp-purpose" className={`text-xs ${C.text50}`}>送信目的</label>
          <input
            id="sp-purpose"
            type="text"
            value={purpose}
            onChange={(e) => setPurpose(e.target.value)}
            placeholder="例: vaccine_reminder"
            className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none w-40`}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="sp-tag-prefix" className={`text-xs ${C.text50}`}>タグプレフィックス</label>
          <input
            id="sp-tag-prefix"
            type="text"
            value={tagPrefix}
            onChange={(e) => setTagPrefix(e.target.value)}
            placeholder="例: PREV_"
            className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none w-28`}
          />
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={handleAdd}
          disabled={createMutation.isPending}
        >
          追加
        </Button>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// LstepTagConfigSection (public)
// ─────────────────────────────────────────────────

export function LstepTagConfigSection() {
  return (
    <div className={`${STYLE.formCard} max-w-2xl mt-6`}>
      <h3 className="text-base font-semibold mb-4" style={{ color: PALETTE.primary }}>
        自動管理タグ設定
      </h3>
      <div className="space-y-8">
        <AutoManagedPrefixesSection />
        <ConditionTagMappingsSection />
        <SendPurposeTagPrefixesSection />
      </div>
    </div>
  );
}

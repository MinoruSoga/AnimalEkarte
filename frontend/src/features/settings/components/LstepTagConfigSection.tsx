import { useActionState, type ReactNode } from "react";
import { toast } from "sonner";
import { BADGE, C, PALETTE, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { handleApiError } from "@/lib/handle-api-error";
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
// TagPairSection (FE-RC-009/050: 共通 追加フォーム抽出)
// ─────────────────────────────────────────────────
//
// AutoManagedPrefixes / ConditionTagMappings / SendPurposeTagPrefixes の
// 3セクションは「一覧 + 2フィールドの追加フォーム」という同型構造を持つ。
// 一覧は既存の TagConfigListSection、追加フォームはここで useActionState +
// <form action> + SubmitButton に統一して重複を解消する。

interface TagPairField {
  name: string;
  label: string;
  placeholder: string;
  widthClassName: string;
  monospace?: boolean;
}

type TagPairFormState = { error?: string } | null;

interface TagPairSectionProps<TItem> {
  title: string;
  description: string;
  items: TItem[] | undefined;
  isLoading: boolean;
  getId: (item: TItem) => string | number;
  renderRow: (item: TItem) => ReactNode;
  onDelete: (item: TItem) => void;
  deleteDisabled?: boolean;
  fieldA: TagPairField;
  fieldB: TagPairField;
  requiredMessage: string;
  addPending: boolean;
  onAdd: (valueA: string, valueB: string) => Promise<void>;
}

function TagPairSection<TItem>({
  title,
  description,
  items,
  isLoading,
  getId,
  renderRow,
  onDelete,
  deleteDisabled,
  fieldA,
  fieldB,
  requiredMessage,
  addPending,
  onAdd,
}: TagPairSectionProps<TItem>) {
  const [state, formAction] = useActionState<TagPairFormState, FormData>(
    async (_prev, formData) => {
      const valueA = ((formData.get(fieldA.name) as string) ?? "").trim();
      const valueB = ((formData.get(fieldB.name) as string) ?? "").trim();
      if (!valueA || !valueB) {
        toast.error(requiredMessage);
        return { error: requiredMessage };
      }
      try {
        await onAdd(valueA, valueB);
        toast.success("追加しました");
        return null;
      } catch (error) {
        handleApiError(error, title);
        return { error: "追加に失敗しました" };
      }
    },
    null,
  );

  return (
    <div>
      <h4 className="text-sm font-medium mb-2" style={{ color: PALETTE.primary }}>
        {title}
      </h4>
      <p className={`text-xs mb-3 ${C.text50}`}>{description}</p>

      <TagConfigListSection
        items={items}
        isLoading={isLoading}
        getId={getId}
        renderRow={renderRow}
        onDelete={onDelete}
        deleteDisabled={deleteDisabled}
      />

      <form action={formAction} className="flex gap-2 items-end" noValidate>
        <div className="flex flex-col gap-1">
          <label htmlFor={fieldA.name} className={`text-xs ${C.text50}`}>
            {fieldA.label}
          </label>
          <input
            id={fieldA.name}
            name={fieldA.name}
            type="text"
            required
            placeholder={fieldA.placeholder}
            className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none ${fieldA.widthClassName}`}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor={fieldB.name} className={`text-xs ${C.text50}`}>
            {fieldB.label}
          </label>
          <input
            id={fieldB.name}
            name={fieldB.name}
            type="text"
            required
            placeholder={fieldB.placeholder}
            className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none ${fieldB.widthClassName}`}
          />
        </div>
        <SubmitButton
          colorVariant="primary"
          size="sm"
          className="h-9"
          disabled={addPending}
          loadingText="追加中..."
        >
          追加
        </SubmitButton>
        {state?.error ? (
          <p className={`text-xs ${C.danger}`} role="alert">
            {state.error}
          </p>
        ) : null}
      </form>
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

  function handleDelete(item: AutoManagedPrefix) {
    deleteMutation.mutate(item.id, {
      onSuccess: () => toast.success("削除しました"),
    });
  }

  return (
    <TagPairSection
      title="自動管理プレフィックス (B/C系)"
      description="このプレフィックスで始まるタグは自動管理対象となり、手動追加・削除が拒否されます。"
      items={items}
      isLoading={isLoading}
      getId={(item) => item.id}
      renderRow={(item) => (
        <>
          <span className="text-sm font-mono flex-1">{item.prefix}</span>
          <span className={`text-xs px-1.5 py-0.5 rounded ${BADGE.blueNoBorder}`}>
            {item.category}
          </span>
          {item.description ? (
            <span className={`text-xs ${C.text50} flex-1`}>{item.description}</span>
          ) : null}
        </>
      )}
      onDelete={handleDelete}
      deleteDisabled={deleteMutation.isPending}
      fieldA={{ name: "amp-prefix", label: "プレフィックス", placeholder: "例: vaccine_", widthClassName: "w-36" }}
      fieldB={{ name: "amp-category", label: "カテゴリ", placeholder: "例: C2", widthClassName: "w-24" }}
      requiredMessage="プレフィックスとカテゴリは必須です"
      addPending={createMutation.isPending}
      onAdd={async (prefix, category) => {
        await createMutation.mutateAsync({ prefix, category });
      }}
    />
  );
}

// ─────────────────────────────────────────────────
// ConditionTagMappingsSection
// ─────────────────────────────────────────────────

function ConditionTagMappingsSection() {
  const { data: items, isLoading } = useGetConditionTagMappings();
  const createMutation = useCreateConditionTagMapping();
  const deleteMutation = useDeleteConditionTagMapping();

  function handleDelete(item: ConditionTagMapping) {
    deleteMutation.mutate(item.id, {
      onSuccess: () => toast.success("削除しました"),
    });
  }

  return (
    <TagPairSection
      title="慢性疾患コード → タグマッピング"
      description="慢性疾患コードが記録された飼い主に付与するタグを設定します。"
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
      fieldA={{ name: "ctm-condition-code", label: "疾患コード", placeholder: "例: DM", widthClassName: "w-24" }}
      fieldB={{ name: "ctm-tag-name", label: "タグ名", placeholder: "例: CHRON_DM", widthClassName: "w-36" }}
      requiredMessage="疾患コードとタグ名は必須です"
      addPending={createMutation.isPending}
      onAdd={async (conditionCode, tagName) => {
        await createMutation.mutateAsync({ condition_code: conditionCode, tag_name: tagName });
      }}
    />
  );
}

// ─────────────────────────────────────────────────
// SendPurposeTagPrefixesSection
// ─────────────────────────────────────────────────

function SendPurposeTagPrefixesSection() {
  const { data: items, isLoading } = useGetSendPurposeTagPrefixes();
  const createMutation = useCreateSendPurposeTagPrefix();
  const deleteMutation = useDeleteSendPurposeTagPrefix();

  function handleDelete(item: SendPurposeTagPrefix) {
    deleteMutation.mutate(item.id, {
      onSuccess: () => toast.success("削除しました"),
    });
  }

  return (
    <TagPairSection
      title="LINE送信目的 → タグプレフィックス"
      description="LINE個別送信時に、送信目的に応じたタグプレフィックスで始まるタグのみ選択可能になります。"
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
      fieldA={{ name: "sp-purpose", label: "送信目的", placeholder: "例: vaccine_reminder", widthClassName: "w-40" }}
      fieldB={{ name: "sp-tag-prefix", label: "タグプレフィックス", placeholder: "例: PREV_", widthClassName: "w-28" }}
      requiredMessage="送信目的とタグプレフィックスは必須です"
      addPending={createMutation.isPending}
      onAdd={async (purpose, tagPrefix) => {
        await createMutation.mutateAsync({ purpose, tag_prefix: tagPrefix });
      }}
    />
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

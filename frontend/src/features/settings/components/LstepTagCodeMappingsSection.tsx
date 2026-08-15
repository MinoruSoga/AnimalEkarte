import { useState } from "react";
import { toast } from "sonner";
import { BADGE, C, STYLE } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import {
  useGetTagCodeMappings,
  usePutTagCodeMappingsForTag,
} from "../hooks/use-lstep-tag-code-mappings";
import type { TagCodeMappingItem, PutMappingEntry } from "../hooks/use-lstep-tag-code-mappings";

// 判定コードを設定可能なタグ名一覧（backend service.ConfigurableTagNames と同期）
const CONFIGURABLE_TAG_NAMES = [
  "HLTH_健診あり",
  "PREV_フィラリア未完了",
  "PREV_ノミダニ対象",
  "LTV_フード購入あり",
] as const;

function groupByTagName(
  items: TagCodeMappingItem[],
): Record<string, TagCodeMappingItem[]> {
  const map: Record<string, TagCodeMappingItem[]> = {};
  for (const item of items) {
    if (!map[item.tag_name]) map[item.tag_name] = [];
    map[item.tag_name].push(item);
  }
  return map;
}

// ─────────────────────────────────────────────────
// Badges
// ─────────────────────────────────────────────────

function NotEnteredBadge() {
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${C.bgDanger8} ${C.danger}`}
    >
      未投入
    </span>
  );
}

function ConfiguredBadge() {
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${BADGE.greenNoBorder}`}
    >
      設定済
    </span>
  );
}

// ─────────────────────────────────────────────────
// MappingRow (view mode)
// ─────────────────────────────────────────────────

function MappingRow({ item }: { item: TagCodeMappingItem }) {
  return (
    <div
      className={`flex items-start gap-3 px-3 py-2 border-t ${C.borderLight}`}
    >
      <span
        className={`text-xs font-mono shrink-0 mt-0.5 px-1.5 py-0.5 rounded ${C.bgSubtle} ${C.text60}`}
      >
        {item.code_type}
      </span>
      {item.codes.length === 0 ? (
        <NotEnteredBadge />
      ) : (
        <span className={`text-sm ${C.text60} break-all`}>
          {item.codes.join(", ")}
        </span>
      )}
    </div>
  );
}

// ─────────────────────────────────────────────────
// Edit form
// ─────────────────────────────────────────────────

interface EntryDraft {
  id: number;
  code_type: string;
  codes_text: string;
}

let _entryIdSeq = 0;
function newEntryDraft(partial?: { code_type: string; codes_text: string }): EntryDraft {
  return { id: _entryIdSeq++, code_type: "", codes_text: "", ...partial };
}

function entryDraftsFromMappings(mappings: TagCodeMappingItem[]): EntryDraft[] {
  if (mappings.length === 0) return [newEntryDraft()];
  return mappings.map((m) =>
    newEntryDraft({ code_type: m.code_type, codes_text: m.codes.join(", ") }),
  );
}

function EntryEditRow({
  entry,
  onChange,
  onRemove,
}: {
  entry: EntryDraft;
  onChange: (updated: EntryDraft) => void;
  onRemove: () => void;
}) {
  return (
    <div className="flex gap-2 items-end">
      <div className="flex flex-col gap-1 w-36 shrink-0">
        <label htmlFor={`code-type-${entry.id}`} className={`text-xs ${C.text60}`}>コード種別</label>
        <input
          id={`code-type-${entry.id}`}
          type="text"
          value={entry.code_type}
          onChange={(e) => onChange({ ...entry, code_type: e.target.value })}
          placeholder="例: checkup_type"
          className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none`}
        />
      </div>
      <div className="flex flex-col gap-1 flex-1">
        <label htmlFor={`codes-text-${entry.id}`} className={`text-xs ${C.text60}`}>コード（カンマ区切り）</label>
        <input
          id={`codes-text-${entry.id}`}
          type="text"
          value={entry.codes_text}
          onChange={(e) => onChange({ ...entry, codes_text: e.target.value })}
          placeholder="例: CHK_A, CHK_B"
          className={`${STYLE.formInput} rounded-xs border px-2 py-1 text-sm outline-none`}
        />
      </div>
      <button
        type="button"
        onClick={onRemove}
        className={`mb-0.5 min-h-11 min-w-11 text-sm ${C.text50} hover:opacity-70 shrink-0`}
        aria-label="行を削除"
      >
        ×
      </button>
    </div>
  );
}

// ─────────────────────────────────────────────────
// TagRow
// ─────────────────────────────────────────────────

function TagRow({
  tagName,
  mappings,
}: {
  tagName: string;
  mappings: TagCodeMappingItem[];
}) {
  const [editing, setEditing] = useState(false);
  const [entries, setEntries] = useState<EntryDraft[]>([]);
  const mutation = usePutTagCodeMappingsForTag();

  const hasAnyCode = mappings.some((m) => m.codes.length > 0);

  const handleStartEdit = () => {
    setEntries(entryDraftsFromMappings(mappings));
    setEditing(true);
  };

  const handleSave = async () => {
    const putEntries: PutMappingEntry[] = entries
      .filter((e) => e.code_type.trim())
      .map((e) => ({
        code_type: e.code_type.trim(),
        codes: e.codes_text
          .split(",")
          .map((c) => c.trim())
          .filter(Boolean),
      }));
    try {
      await mutation.mutateAsync({ tagName, req: { entries: putEntries } });
      toast.success(`${tagName} を保存しました`);
      setEditing(false);
    } catch {
      // handleApiError は mutation onError で処理済み
    }
  };

  return (
    <div className={`border ${C.borderLight} rounded-xs overflow-hidden`}>
      {/* Header */}
      <div
        className={`flex items-center justify-between px-3 py-2.5 ${C.bgSubtle}`}
      >
        <span className={`text-sm font-medium ${C.text}`}>{tagName}</span>
        <div className="flex items-center gap-2">
          {hasAnyCode ? <ConfiguredBadge /> : <NotEnteredBadge />}
          {!editing ? (
            <Button
              variant="outline"
              size="sm"
              onClick={handleStartEdit}
              className="h-6 px-2 text-xs"
            >
              編集
            </Button>
          ) : null}
        </div>
      </div>

      {/* View mode */}
      {!editing ? (
        mappings.length === 0 ? (
          <div className={`px-3 py-2 border-t ${C.borderLight}`}>
            <span className={`text-xs ${C.text50}`}>
              コードが未設定です。
            </span>
          </div>
        ) : (
          mappings.map((m) => <MappingRow key={m.id} item={m} />)
        )
      ) : (
        /* Edit mode */
        <div className={`border-t ${C.borderLight} px-3 py-3 flex flex-col gap-2`}>
          {entries.map((entry) => (
            <EntryEditRow
              key={entry.id}
              entry={entry}
              onChange={(updated) =>
                setEntries((prev) =>
                  prev.map((e) => (e.id === entry.id ? updated : e)),
                )
              }
              onRemove={() =>
                setEntries((prev) => prev.filter((e) => e.id !== entry.id))
              }
            />
          ))}
          <Button
            variant="ghost"
            size="sm"
            type="button"
            onClick={() => setEntries((prev) => [...prev, newEntryDraft()])}
            className="self-start h-7 px-2 text-xs"
          >
            + 行を追加
          </Button>
          <div className="flex gap-2">
            <Button
              size="sm"
              onClick={handleSave}
              disabled={mutation.isPending}
              className="h-7 px-3 text-xs"
            >
              {mutation.isPending ? "保存中..." : "保存"}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setEditing(false)}
              disabled={mutation.isPending}
              className="h-7 px-3 text-xs"
            >
              キャンセル
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

// ─────────────────────────────────────────────────
// LstepTagCodeMappingsSection
// ─────────────────────────────────────────────────

export function LstepTagCodeMappingsSection() {
  const { data, isLoading, isError } = useGetTagCodeMappings();

  if (isLoading) {
    return (
      <div className={`${STYLE.formCard} max-w-2xl mt-6`}>
        <div
          className={`flex items-center justify-center py-8 text-sm ${C.text50}`}
        >
          読み込み中...
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={`${STYLE.formCard} max-w-2xl mt-6`}>
        <div
          className={`flex items-center justify-center py-8 text-sm ${C.danger}`}
        >
          読み込みに失敗しました
        </div>
      </div>
    );
  }

  const grouped = groupByTagName(data ?? []);

  return (
    <div className={`${STYLE.formCard} max-w-2xl mt-6`}>
      <h2 className={`text-base font-semibold ${C.text} mb-1`}>
        タグコードマッピング
      </h2>
      <p className={`text-sm ${C.text60} mb-4`}>
        健診・予防タグの判定に使用する診察種別・処方コードの設定。
      </p>

      <div className="flex flex-col gap-3">
        {CONFIGURABLE_TAG_NAMES.map((tagName) => (
          <TagRow
            key={tagName}
            tagName={tagName}
            mappings={grouped[tagName] ?? []}
          />
        ))}
      </div>
    </div>
  );
}

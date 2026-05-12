import { C, STYLE } from "@/lib/design-tokens";
import { useGetTagCodeMappings } from "./hooks/useLstepTagCodeMappings";
import type { TagCodeMappingItem } from "./hooks/useLstepTagCodeMappings";

// SPEC-002 Q5 確定待ちタグ名一覧（backend service.ConfigurableTagNames と同期）
const CONFIGURABLE_TAG_NAMES = [
  "HLTH_健診あり",
  "HLTH_専門検診候補",
  "PREV_フィラリア未完了",
  "PREV_ノミダニ対象",
  "LTV_フード購入あり",
  "LTV_サプリ購入あり",
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
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${C.bgAccentLight8} ${C.accent}`}
    >
      設定済
    </span>
  );
}

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

function TagRow({
  tagName,
  mappings,
}: {
  tagName: string;
  mappings: TagCodeMappingItem[];
}) {
  const hasAnyCode = mappings.some((m) => m.codes.length > 0);

  return (
    <div className={`border ${C.borderLight} rounded-[4px] overflow-hidden`}>
      <div
        className={`flex items-center justify-between px-3 py-2.5 ${C.bgSubtle}`}
      >
        <span className={`text-sm font-medium ${C.text}`}>{tagName}</span>
        {hasAnyCode ? <ConfiguredBadge /> : <NotEnteredBadge />}
      </div>
      {mappings.length === 0 ? (
        <div className={`px-3 py-2 border-t ${C.borderLight}`}>
          <span className={`text-xs ${C.text50}`}>
            {/* TODO(SPEC-002 Q5): コードリストが確定後に追加 */}
            コードが未設定です。SPEC-002 Q5 確定後に追加予定。
          </span>
        </div>
      ) : (
        mappings.map((m) => <MappingRow key={m.id} item={m} />)
      )}
    </div>
  );
}

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

      {/* SPEC-002 Q5 未確定バナー */}
      <div
        className={`mb-4 rounded-[4px] px-3 py-2.5 text-sm ${C.bgDanger8} ${C.danger} border ${C.borderLight}`}
      >
        <strong>⚠ SPEC-002 Q5 確定待ち</strong>
        <span className="ml-2">
          各タグに紐づく実コード一覧は PO 確認後に追加されます。現在は骨組みのみ表示。
        </span>
      </div>

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

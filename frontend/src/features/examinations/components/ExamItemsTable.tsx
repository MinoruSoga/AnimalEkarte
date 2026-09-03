// React/Framework
import { memo, useCallback, useId, useLayoutEffect, useRef } from "react";

// External
import { Plus, Trash2 } from "lucide-react";

// Internal
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { C, ICON } from "@/lib/design-tokens";

// Relative
import { ExamStatusBadge } from "./ExamStatusBadge";

// 検査項目テーブルの 1 行。
// status / isAssessed / isAbnormal は backend が導出した値を表示するだけ（FE で再計算しない）。
export interface ExamItemRow {
  /** クライアント側の React key 用ローカル ID（不変） */
  key: string;
  /** exam_type_fields.id（テンプレ由来。手動追加行では undefined） */
  examTypeFieldId?: number;
  name: string;
  inspectionValue: string;
  unit: string;
  normalValue: string;
  referenceValue: string;
  refMin?: number;
  refMax?: number;
  sortOrder: number;
  /** backend が導出した判定。新規行・未保存値では undefined */
  status?: "normal" | "high" | "low";
  /** backend が基準値の有無から導出した評価状態。新規行・未保存値では undefined */
  isAssessed?: boolean;
  isAbnormal?: boolean;
}

interface ExamItemsTableProps {
  items: ExamItemRow[];
  onChangeInspectionValue: (key: string, value: string) => void;
  onAddItem?: () => void;
  onRemoveItem?: (key: string) => void;
  onChangeName?: (key: string, value: string) => void;
  disabled?: boolean;
}

/**
 * 検査項目テーブル — `ExaminationForm` に組み込む編集可能テーブル。
 *
 * 表示列: 項目名 / 結果値（編集可） / 単位 / 基準値 / 判定（backend 由来）/ 操作
 * - 異常値判定は backend service 層で行われる。FE はサーバが返した status を表示するだけ。
 * - テンプレ行の名前は固定し、手動追加行の名前だけを編集できる。
 * - 確定済み (`disabled=true`) では input が disabled になる。
 */
export const ExamItemsTable = memo(function ExamItemsTable({
  items,
  onChangeInspectionValue,
  onAddItem,
  onRemoveItem,
  onChangeName,
  disabled = false,
}: ExamItemsTableProps) {
  const tableId = useId();
  const addButtonRef = useRef<HTMLButtonElement>(null);
  const nameInputRefs = useRef(new Map<string, HTMLInputElement>());
  const removeButtonRefs = useRef(new Map<string, HTMLButtonElement>());
  const previousItemKeysRef = useRef(items.map((item) => item.key));
  const pendingFocusRef = useRef<{ type: "add" } | { type: "remove"; index: number } | null>(null);

  useLayoutEffect(() => {
    const pendingFocus = pendingFocusRef.current;
    const previousKeys = previousItemKeysRef.current;

    if (pendingFocus?.type === "add") {
      const addedKey = items.find((item) => !previousKeys.includes(item.key))?.key;
      if (addedKey) {
        nameInputRefs.current.get(addedKey)?.focus();
      }
    } else if (pendingFocus?.type === "remove") {
      const focusKey = items[pendingFocus.index]?.key ?? items[pendingFocus.index - 1]?.key;
      const focusTarget = focusKey ? removeButtonRefs.current.get(focusKey) : addButtonRef.current;
      focusTarget?.focus();
    }

    previousItemKeysRef.current = items.map((item) => item.key);
    pendingFocusRef.current = null;
  }, [items]);
  const handleChange = useCallback(
    (key: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
      onChangeInspectionValue(key, e.target.value);
    },
    [onChangeInspectionValue],
  );
  const handleNameChange = useCallback(
    (key: string) => (event: React.ChangeEvent<HTMLInputElement>) => {
      onChangeName?.(key, event.target.value);
    },
    [onChangeName],
  );
  const handleAddItem = useCallback(() => {
    pendingFocusRef.current = { type: "add" };
    onAddItem?.();
  }, [onAddItem]);
  const handleRemoveItem = useCallback(
    (key: string, index: number) => {
      pendingFocusRef.current = { type: "remove", index };
      onRemoveItem?.(key);
    },
    [onRemoveItem],
  );

  return (
    <div className="space-y-2">
      {onAddItem ? (
        <div className="flex justify-end">
          <Button
            ref={addButtonRef}
            type="button"
            variant="outline"
            onClick={handleAddItem}
            disabled={disabled}
            className="h-11 min-w-11 text-sm"
          >
            <Plus aria-hidden="true" className={`mr-1.5 ${ICON.action}`} />
            検査項目を追加
          </Button>
        </div>
      ) : null}
      {items.length === 0 ? (
        <div className={`p-4 rounded-lg border ${C.borderMedium} ${C.bgWhite}`}>
          <p className={`text-sm ${C.text45} text-center`}>
            検査種別を選択すると検査項目が表示されます
          </p>
        </div>
      ) : (
        <div
          className={`border ${C.borderMedium} rounded-lg ${C.bgWhite} overflow-hidden overflow-x-auto`}
        >
          {/* ヘッダー */}
          <div
            className={`min-w-[700px] grid grid-cols-[2fr_1.5fr_1fr_1.8fr_1.2fr_0.8fr] gap-0 border-b ${C.borderMedium} ${C.bgPage} text-sm font-bold ${C.text80} h-11 items-center`}
          >
            <div className={`p-2 border-r ${C.borderMedium} pl-3`}>項目名</div>
            <div className={`p-2 border-r ${C.borderMedium} text-right`}>結果値</div>
            <div className={`p-2 border-r ${C.borderMedium} text-center`}>単位</div>
            <div className={`p-2 border-r ${C.borderMedium} text-center`}>基準値</div>
            <div className={`p-2 border-r ${C.borderMedium} text-center`}>判定</div>
            <div className="p-2 text-center">操作</div>
          </div>
          {items.map((item, idx) => (
            <div
              key={item.key}
              data-testid="exam-item-row"
              data-abnormal={String(!!item.isAbnormal)}
              className={`min-w-[700px] grid grid-cols-[2fr_1.5fr_1fr_1.8fr_1.2fr_0.8fr] gap-0 ${
                idx !== items.length - 1 ? `border-b ${C.borderMedium}` : ""
              } ${
                item.isAbnormal
                  ? item.status === "high"
                    ? C.bgDanger8
                    : C.bgStatusBlueLight
                  : C.bgWhite
              } text-sm ${C.text} items-center min-h-12`}
            >
              <div className={`p-2 border-r ${C.borderMedium} pl-3 font-medium`}>
                {item.examTypeFieldId === undefined && onChangeName ? (
                  <Input
                    ref={(node) => {
                      if (node) nameInputRefs.current.set(item.key, node);
                      else nameInputRefs.current.delete(item.key);
                    }}
                    id={`${tableId}-name-${idx + 1}`}
                    name={`examItems.${idx}.name`}
                    value={item.name}
                    disabled={disabled}
                    onChange={handleNameChange(item.key)}
                    placeholder="項目名"
                    className={`h-11 min-w-11 text-sm ${C.bgWhite} ${C.borderMedium}`}
                    aria-label={`${item.name.trim() || `検査項目${idx + 1}`}の項目名`}
                  />
                ) : (
                  item.name
                )}
              </div>
              <div className={`px-1.5 border-r ${C.borderMedium}`}>
                <Input
                  id={`${tableId}-result-${idx + 1}`}
                  name={`examItems.${idx}.inspectionValue`}
                  type="text"
                  inputMode="decimal"
                  value={item.inspectionValue}
                  disabled={disabled}
                  onChange={handleChange(item.key)}
                  placeholder="-"
                  className={`h-11 min-w-11 text-sm text-right font-mono ${C.bgWhite} ${C.borderMedium}`}
                  aria-label={`${item.name.trim() || `検査項目${idx + 1}`}の結果値`}
                />
              </div>
              <div className={`p-2 border-r ${C.borderMedium} text-center ${C.text60} text-sm`}>
                {item.unit || "-"}
              </div>
              <div className={`p-2 border-r ${C.borderMedium} text-center ${C.text60} text-sm`}>
                {item.referenceValue || item.normalValue || "-"}
              </div>
              <div className={`p-2 border-r ${C.borderMedium} flex justify-center items-center`}>
                <ExamStatusBadge status={item.status} isAssessed={item.isAssessed} />
              </div>
              <div className="p-0.5 flex justify-center items-center">
                {onRemoveItem ? (
                  <Button
                    ref={(node) => {
                      if (node) removeButtonRefs.current.set(item.key, node);
                      else removeButtonRefs.current.delete(item.key);
                    }}
                    type="button"
                    variant="ghost"
                    disabled={disabled}
                    onClick={() => handleRemoveItem(item.key, idx)}
                    aria-label={`${item.name.trim() || `検査項目${idx + 1}`}を削除`}
                    className={`h-11 min-w-11 ${C.danger}`}
                  >
                    <Trash2 aria-hidden="true" className={ICON.action} />
                  </Button>
                ) : null}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
});

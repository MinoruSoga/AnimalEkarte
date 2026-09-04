import { memo, useCallback, useEffect, useState } from "react";
import type { ChangeEvent } from "react";
import { Lock } from "lucide-react";

import { MasterSidePanel, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { C, LAYOUT } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";

import type { PermissionGroup } from "../api/permission-groups";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import { PermissionRuleTable } from "./PermissionRuleTable";
import type { PermissionRuleField } from "../lib/permission-rule-table-model";
import {
  permissionGroupToFormData,
  type PermissionGroupFormData,
  updatePermissionRule,
} from "../lib/permission-group-side-panel-model";

interface PermissionGroupSidePanelProps {
  item: PermissionGroup | null;
  onClose: () => void;
  onSave: (data: PermissionGroupFormData) => Promise<boolean>;
  onSaveSuccess: () => void;
  onDeleteRequest?: (item: PermissionGroup) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const PermissionGroupSidePanel = memo(function PermissionGroupSidePanel({
  item,
  onClose,
  onSave,
  onSaveSuccess,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: PermissionGroupSidePanelProps) {
  const isNew = item === null;

  const [formData, setFormData] = useState<PermissionGroupFormData>(() =>
    permissionGroupToFormData(item),
  );
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleSave = useCallback(async () => {
    if (!formData.name.trim()) {
      setNameError("グループ名を入力してください");
      return;
    }
    setNameError("");
    try {
      const saved = await onSave(formData);
      if (saved) {
        setIsDirty(false);
        onDirtyChange?.(false);
        onSaveSuccess();
      }
    } catch (error) {
      handleApiError(error, "保存");
    }
  }, [formData, onDirtyChange, onSave, onSaveSuccess]);

  const handleNameChange = useCallback(
    (value: string) => {
      if (readOnly) return;
      setFormDataDirty((prev) => ({ ...prev, name: value }));
      if (value.trim()) setNameError("");
    },
    [readOnly, setFormDataDirty],
  );

  const handleDescriptionChange = useCallback(
    (event: ChangeEvent<HTMLTextAreaElement>) => {
      setFormDataDirty((prev) => ({ ...prev, description: event.target.value }));
    },
    [setFormDataDirty],
  );

  const handleColorChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, color: value }));
    },
    [setFormDataDirty],
  );

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleRuleChange = useCallback(
    (resource: string, field: PermissionRuleField, value: boolean) => {
      setFormDataDirty((prev) => ({
        ...prev,
        rules: updatePermissionRule(prev.rules, resource, field, value),
      }));
    },
    [setFormDataDirty],
  );

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={isNew}
      title={formData.name}
      onTitleChange={handleNameChange}
      onClose={handleClose}
      action={readOnly ? undefined : handleSave}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Lock className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <fieldset disabled={readOnly} className="contents">
        <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />

        <PropertyRow label="説明">
          <textarea
            className={`${MASTER_INPUT_CLASS} resize-none`}
            value={formData.description}
            onChange={handleDescriptionChange}
            placeholder="グループの説明"
            rows={3}
          />
        </PropertyRow>

        <PropertyRow label="カラー">
          <div className="flex items-center gap-3">
            <input
              type="color"
              aria-label="カラー"
              className={`${LAYOUT.colorInputMedium} ${C.borderMedium}`}
              value={formData.color}
              onChange={(event) => handleColorChange(event.target.value)}
            />
            <span className={`text-sm ${C.text50}`}>{formData.color}</span>
          </div>
        </PropertyRow>

        <PermissionRuleTable
          rules={formData.rules}
          onRuleChange={handleRuleChange}
          disabled={readOnly}
        />
      </fieldset>
    </MasterSidePanel>
  );
});

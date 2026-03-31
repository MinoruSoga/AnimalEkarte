// React/Framework
import { ICON, C } from "@/lib/design-tokens";
import { useState, useCallback, useMemo } from "react";

// External
import { Pill, Stethoscope, Utensils, ClipboardList, MoreHorizontal, Pencil, Plus, Loader2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { Input } from "@/components/ui/input";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

// Relative
import {
    useGetCarePlanItems,
    useCreateCarePlanItem,
    useUpdateCarePlanItem,
    useDeleteCarePlanItem,
} from "@/features/hospitalization/api/care-plan-items";

// Types
import type { CarePlanItem, CarePlanItemType, CarePlanTiming, UpdateCarePlanItemInput } from "@/features/hospitalization/api/care-plan-items";

// ---- Static constants ----

const TYPE_OPTIONS: { value: CarePlanItemType; label: string }[] = [
    { value: "food", label: "食事" },
    { value: "medicine", label: "投薬" },
    { value: "treatment", label: "処置・検査" },
    { value: "instruction", label: "指示・その他" },
    { value: "item", label: "持ち物" },
];

const TIMING_OPTIONS: { value: CarePlanTiming; label: string }[] = [
    { value: "morning", label: "朝" },
    { value: "noon", label: "昼" },
    { value: "night", label: "夜" },
];

const TYPE_SELECT_ITEMS = (
    <>
        {TYPE_OPTIONS.map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
            </SelectItem>
        ))}
    </>
);

// ---- Helper functions ----

function TypeIcon({ type }: { type: CarePlanItemType }) {
    if (type === "food") return <Utensils className={`${ICON.action} text-orange-500 shrink-0`} />;
    if (type === "medicine") return <Pill className={`${ICON.action} text-blue-500 shrink-0`} />;
    if (type === "treatment") return <Stethoscope className={`${ICON.action} text-purple-500 shrink-0`} />;
    if (type === "instruction") return <ClipboardList className={`${ICON.action} text-green-500 shrink-0`} />;
    return <MoreHorizontal className={`${ICON.action} text-gray-400 shrink-0`} />;
}

function StatusBadge({ status }: { status: CarePlanItem["status"] }) {
    if (status === "active") {
        return (
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700">
                実施中
            </span>
        );
    }
    if (status === "completed") {
        return (
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-700">
                完了
            </span>
        );
    }
    return (
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-500">
            中止
        </span>
    );
}

function TimingBadges({ timing }: { timing: CarePlanTiming[] }) {
    return (
        <div className="flex gap-1">
            {TIMING_OPTIONS.map((opt) => (
                <span
                    key={opt.value}
                    className={`inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium ${
                        timing.includes(opt.value)
                            ? "bg-indigo-100 text-indigo-700"
                            : "bg-gray-50 text-gray-300"
                    }`}
                >
                    {opt.label}
                </span>
            ))}
        </div>
    );
}

// ---- Edit Row ----

interface EditRowProps {
    item: CarePlanItem;
    onSave: (input: UpdateCarePlanItemInput) => void;
    onCancel: () => void;
    isSaving: boolean;
}

function EditRow({ item, onSave, onCancel, isSaving }: EditRowProps) {
    const [name, setName] = useState(item.name);
    const [type, setType] = useState<CarePlanItemType>(item.type);
    const [timing, setTiming] = useState<CarePlanTiming[]>(item.timing);

    const handleTimingToggle = useCallback((t: CarePlanTiming) => {
        setTiming((prev) =>
            prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]
        );
    }, []);

    const handleSave = useCallback(() => {
        if (!name.trim()) return;
        onSave({ name: name.trim(), type, timing });
    }, [name, type, timing, onSave]);

    return (
        <div className="flex flex-col gap-2 p-3 bg-blue-50/50 rounded-lg border border-blue-100">
            <div className="flex gap-2 items-center">
                <Select value={type} onValueChange={(v) => setType(v as CarePlanItemType)}>
                    <SelectTrigger className="w-28 h-8 text-xs">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>{TYPE_SELECT_ITEMS}</SelectContent>
                </Select>
                <Input
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="h-8 text-sm flex-1"
                    placeholder="名称"
                />
            </div>
            <div className="flex items-center gap-3">
                <span className="text-xs text-gray-500 shrink-0">タイミング:</span>
                <div className="flex gap-2">
                    {TIMING_OPTIONS.map((opt) => (
                        <label key={opt.value} className="flex items-center gap-1 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={timing.includes(opt.value)}
                                onChange={() => handleTimingToggle(opt.value)}
                                className="rounded"
                            />
                            <span className="text-xs">{opt.label}</span>
                        </label>
                    ))}
                </div>
            </div>
            <div className="flex gap-2 justify-end">
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={onCancel}
                    disabled={isSaving}
                    className="h-7 text-xs"
                >
                    キャンセル
                </Button>
                <Button
                    size="sm"
                    onClick={handleSave}
                    disabled={isSaving || !name.trim()}
                    className="h-7 text-xs"
                >
                    {isSaving ? <Loader2 className={`${ICON.action} animate-spin mr-1`} /> : null}
                    保存
                </Button>
            </div>
        </div>
    );
}

// ---- Item Row ----

interface ItemRowProps {
    item: CarePlanItem;
    onEdit: (id: string) => void;
    onDelete: (id: string) => void;
    isDeleting: boolean;
}

function ItemRow({ item, onEdit, onDelete, isDeleting }: ItemRowProps) {
    return (
        <div className="flex items-center gap-3 py-2 px-3 rounded-lg hover:bg-gray-50 transition-colors">
            <TypeIcon type={item.type} />
            <span className={`flex-1 text-sm font-medium ${C.text} truncate`}>{item.name}</span>
            <TimingBadges timing={item.timing} />
            <StatusBadge status={item.status} />
            <div className="flex gap-1 shrink-0">
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => onEdit(item.id)}
                    aria-label="編集"
                >
                    <Pencil className={ICON.action} />
                </Button>
                <DeleteIconButton
                    onClick={() => onDelete(item.id)}
                    disabled={isDeleting}
                    className="size-7"
                />
            </div>
        </div>
    );
}

// ---- Add Form ----

interface AddFormProps {
    onSubmit: (type: CarePlanItemType, name: string, timing: CarePlanTiming[]) => void;
    isSubmitting: boolean;
}

function AddForm({ onSubmit, isSubmitting }: AddFormProps) {
    const [type, setType] = useState<CarePlanItemType>("instruction");
    const [name, setName] = useState("");
    const [timing, setTiming] = useState<CarePlanTiming[]>(["morning"]);

    const handleTimingToggle = useCallback((t: CarePlanTiming) => {
        setTiming((prev) =>
            prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]
        );
    }, []);

    const handleSubmit = useCallback(() => {
        if (!name.trim()) return;
        onSubmit(type, name.trim(), timing);
        setName("");
        setType("instruction");
        setTiming(["morning"]);
    }, [name, type, timing, onSubmit]);

    return (
        <div className={`border-t ${C.borderLight} pt-3 mt-2`}>
            <p className={`text-xs font-medium ${C.text60} mb-2`}>新しいケアプラン項目を追加</p>
            <div className="flex flex-col gap-2">
                <div className="flex gap-2 items-center">
                    <Select value={type} onValueChange={(v) => setType(v as CarePlanItemType)}>
                        <SelectTrigger className="w-28 h-8 text-xs">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>{TYPE_SELECT_ITEMS}</SelectContent>
                    </Select>
                    <Input
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className="h-8 text-sm flex-1"
                        placeholder="名称を入力"
                        onKeyDown={(e) => {
                            if (e.key === "Enter") handleSubmit();
                        }}
                    />
                </div>
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <span className="text-xs text-gray-500 shrink-0">タイミング:</span>
                        <div className="flex gap-2">
                            {TIMING_OPTIONS.map((opt) => (
                                <label key={opt.value} className="flex items-center gap-1 cursor-pointer">
                                    <input
                                        type="checkbox"
                                        checked={timing.includes(opt.value)}
                                        onChange={() => handleTimingToggle(opt.value)}
                                        className="rounded"
                                    />
                                    <span className="text-xs">{opt.label}</span>
                                </label>
                            ))}
                        </div>
                    </div>
                    <Button
                        size="sm"
                        onClick={handleSubmit}
                        disabled={isSubmitting || !name.trim()}
                        className="h-8 text-xs gap-1"
                    >
                        {isSubmitting ? (
                            <Loader2 className={`${ICON.action} animate-spin`} />
                        ) : (
                            <Plus className={ICON.action} />
                        )}
                        追加
                    </Button>
                </div>
            </div>
        </div>
    );
}

// ---- Main Component ----

interface CarePlanTabProps {
    hospitalizationId: string;
}

export function CarePlanTab({ hospitalizationId }: CarePlanTabProps) {
    const { data: items, isLoading } = useGetCarePlanItems(hospitalizationId);
    const createItem = useCreateCarePlanItem(hospitalizationId);
    const updateItem = useUpdateCarePlanItem(hospitalizationId);
    const deleteItem = useDeleteCarePlanItem(hospitalizationId);

    const [editingId, setEditingId] = useState<string | null>(null);
    const [deletingId, setDeletingId] = useState<string | null>(null);

    const handleEdit = useCallback((id: string) => {
        setEditingId(id);
    }, []);

    const handleCancelEdit = useCallback(() => {
        setEditingId(null);
    }, []);

    const handleSaveEdit = useCallback(
        (itemId: string, input: UpdateCarePlanItemInput) => {
            updateItem.mutate(
                { itemId, input },
                {
                    onSuccess: () => {
                        setEditingId(null);
                    },
                }
            );
        },
        [updateItem]
    );

    const handleDelete = useCallback(
        (itemId: string) => {
            setDeletingId(itemId);
            deleteItem.mutate(itemId, {
                onSettled: () => {
                    setDeletingId(null);
                },
            });
        },
        [deleteItem]
    );

    const handleAdd = useCallback(
        (type: CarePlanItemType, name: string, timing: CarePlanTiming[]) => {
            createItem.mutate({ type, name, timing });
        },
        [createItem]
    );

    const itemRows = useMemo(() => {
        if (!items) return null;
        return items.map((item) =>
            editingId === item.id ? (
                <EditRow
                    key={item.id}
                    item={item}
                    onSave={(input) => handleSaveEdit(item.id, input)}
                    onCancel={handleCancelEdit}
                    isSaving={updateItem.isPending}
                />
            ) : (
                <ItemRow
                    key={item.id}
                    item={item}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                    isDeleting={deletingId === item.id}
                />
            )
        );
    }, [items, editingId, deletingId, updateItem.isPending, handleEdit, handleDelete, handleSaveEdit, handleCancelEdit]);

    if (isLoading) {
        return (
            <div className={`flex items-center justify-center py-10 ${C.text40}`}>
                <Loader2 className={`${ICON.page} animate-spin mr-2`} />
                <span className="text-sm">読み込み中...</span>
            </div>
        );
    }

    return (
        <div className="flex flex-col">
            {items && items.length === 0 ? (
                <p className={`text-sm ${C.text40} py-4 text-center`}>
                    ケアプラン項目がありません
                </p>
            ) : (
                <div className="flex flex-col gap-1">
                    {itemRows}
                </div>
            )}
            <AddForm onSubmit={handleAdd} isSubmitting={createItem.isPending} />
        </div>
    );
}

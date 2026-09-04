import { useCallback, useMemo, useState } from "react";

interface UseEditableIdSelectionOptions {
  serverIds: string[] | undefined;
  markDirty: () => void;
}

export function useEditableIdSelection({ serverIds, markDirty }: UseEditableIdSelectionOptions) {
  const [userEditedIds, setUserEditedIds] = useState<string[] | null>(null);

  const ids = useMemo(() => userEditedIds ?? serverIds ?? [], [userEditedIds, serverIds]);
  const idSet = useMemo(() => new Set(ids), [ids]);

  const handleToggle = useCallback(
    (id: string, checked: boolean) => {
      setUserEditedIds((prev) => {
        const current = prev ?? serverIds ?? [];
        return checked ? [...current, id] : current.filter((currentId) => currentId !== id);
      });
      markDirty();
    },
    [serverIds, markDirty],
  );

  return { ids, idSet, handleToggle };
}

export const OUTSIDE_FIRST_PAGE_PET = {
  id: '1003298',
  name: 'SPANKY',
} as const;

export interface RuntimePetReference {
  id: string;
  name: string;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function stringifyId(value: unknown): string | null {
  if (typeof value === 'string') return value;
  if (typeof value === 'number' && Number.isSafeInteger(value)) return String(value);
  return null;
}

/**
 * Read only the stable fields needed by the regression from the live list DTO.
 * The response remains an external boundary, so malformed entries are ignored
 * instead of being cast to the generated application model.
 */
export function readRuntimePetReferences(payload: unknown): RuntimePetReference[] {
  if (!isRecord(payload) || !Array.isArray(payload.data)) return [];

  return payload.data.flatMap((item) => {
    if (!isRecord(item) || typeof item.name !== 'string') return [];
    const id = stringifyId(item.id);
    return id === null ? [] : [{ id, name: item.name }];
  });
}

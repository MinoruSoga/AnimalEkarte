/**
 * Standard state for React 19 useActionState hooks.
 * Ensures consistent handling of success, global errors, and field-level validation errors.
 */
export interface ActionState<T = unknown> {
  /** Whether the last action was successful */
  success: boolean;
  /** Global error message (e.g., "Network Error", "Invalid Credentials") */
  error?: string | null;
  /** Map of field names to their specific error messages */
  fieldErrors?: Record<string, string>;
  /** Optional resulting data (e.g., created resource ID) */
  data?: T;
  /** Unique identifier for the last action attempt (useful for useEffect deps) */
  timestamp: number;
}

/** Initial empty state constant */
export const INITIAL_ACTION_STATE: ActionState<unknown> = {
  success: false,
  error: null,
  fieldErrors: {},
  timestamp: 0,
};

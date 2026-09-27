/**
 * Symlink-safe recursive file enumeration for the e2e tree.
 *
 * Shared by the page-consumer contract (verify-e2e-page-consumers.mjs) for
 * spec discovery and page-object discovery. Fail-closed policy: throws on any
 * symlink (root or descendant), unreadable path, or non-directory root unless
 * the caller marks the root optional.
 */

import { lstatSync, readdirSync } from "node:fs";
import path from "node:path";

const SKIP_DIRS = new Set(["node_modules", "dist"]);

function errText(error) {
  return error && error.message ? error.message : error;
}

function walk(root, options) {
  let rootStat;
  try {
    rootStat = lstatSync(root);
  } catch (error) {
    throw new Error(`${options.unreadableLabel}: ${root}: ${errText(error)}`);
  }
  if (rootStat.isSymbolicLink()) {
    throw new Error(`${options.symlinkRootLabel}: ${root}`);
  }
  if (!rootStat.isDirectory()) {
    if (options.requireDirectory) {
      throw new Error(`e2e root is not a directory: ${root}`);
    }
    return [];
  }

  const out = [];
  const stack = [root];
  while (stack.length) {
    const current = stack.pop();
    let entries;
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch (error) {
      throw new Error(`${options.unreadableLabel}: ${current}: ${errText(error)}`);
    }
    for (const entry of entries) {
      const full = path.join(current, entry.name);
      let st;
      try {
        st = lstatSync(full);
      } catch (error) {
        throw new Error(`${options.unreadableLabel}: ${full}: ${errText(error)}`);
      }
      // Fail closed on any symlink before extension filtering; never follow links.
      if (st.isSymbolicLink()) {
        throw new Error(`${options.symlinkEntryLabel}: ${full}`);
      }
      if (st.isDirectory()) {
        if (!SKIP_DIRS.has(entry.name)) {
          stack.push(full);
        }
      } else if (st.isFile() && options.predicate(entry.name)) {
        out.push(full);
      }
    }
  }
  return out.sort();
}

/** Enumerate e2e spec files recursively (never follows symlinks). */
export function listSpecFiles(e2eRoot) {
  return walk(e2eRoot, {
    unreadableLabel: "unreadable e2e scan path",
    symlinkRootLabel: "e2e root is a symlink",
    symlinkEntryLabel: "symlink in e2e tree",
    requireDirectory: true,
    predicate: (name) => name.endsWith(".spec.ts"),
  });
}

/**
 * Enumerate page-object .ts files under e2e/pages recursively (never follows
 * symlinks). Base classes and shared page helpers are consumed transitively
 * via the concrete pages that specs import, so they cannot satisfy the direct
 * spec-consumer contract on their own. A missing or non-directory pages dir
 * yields an empty list.
 */
export function listPageFiles(e2eRoot) {
  return walk(path.join(e2eRoot, "pages"), {
    unreadableLabel: "unreadable e2e pages path",
    symlinkRootLabel: "e2e pages dir is a symlink",
    symlinkEntryLabel: "symlink in e2e pages tree",
    requireDirectory: false,
    predicate: (name) => name.endsWith(".ts") && !name.endsWith(".spec.ts"),
  });
}

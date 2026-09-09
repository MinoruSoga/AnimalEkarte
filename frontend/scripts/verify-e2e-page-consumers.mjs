#!/usr/bin/env node
/**
 * Exact E2E page-object consumer closure via TypeScript AST (not regex).
 *
 * CLI (cwd = frontend root):
 *   node scripts/verify-e2e-page-consumers.mjs --page e2e/pages/foo.ts [--page ...] [--e2e-root e2e]
 */

import { createRequire } from "node:module";
import {
  closeSync,
  constants as fsConstants,
  fstatSync,
  lstatSync,
  openSync,
  readdirSync,
  readFileSync,
  realpathSync,
  statSync,
} from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";

const require = createRequire(import.meta.url);
const ts = require("typescript");

// Only importer-relative ImportDeclaration with clause counts as a consumer.
const SUPPORTED_FORMS = new Set(["static-import"]);
const UNSUPPORTED_EXACT_FORMS = new Set([
  "side-effect-import",
  "dynamic-import",
  "export-from",
  "require",
  "absolute-static-import",
]);

function isImporterRelativeSpecifier(specifier) {
  return (
    typeof specifier === "string" &&
    (specifier.startsWith("./") || specifier.startsWith("../"))
  );
}

function toPosix(value) {
  return String(value).replaceAll("\\", "/");
}

function splitSegments(posixPath) {
  return toPosix(posixPath).split("/").filter((part) => part.length > 0);
}

function identitiesEqual(left, right) {
  if (!left || !right) return false;
  const a = splitSegments(left);
  const b = splitSegments(right);
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i += 1) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}

function withTsExtension(posixPath) {
  const normalized = toPosix(posixPath).replace(/\/+$/, "");
  if (normalized.endsWith(".ts")) return normalized;
  return `${normalized}.ts`;
}

function normalizePageIdentity(pagePath) {
  const raw = toPosix(pagePath).replace(/^\.\/+/, "");
  const stripped = raw.startsWith("frontend/") ? raw.slice("frontend/".length) : raw;
  const identity = withTsExtension(stripped);
  // Reject //, /./, whitespace, and traversal before segment filtering collapses them.
  if (/\s/.test(identity) || identity.includes("\\")) return null;
  if (identity.split("/").some((part) => part === "" || part === "." || part === "..")) {
    return null;
  }
  const parts = splitSegments(identity);
  if (parts.length < 3 || parts[0] !== "e2e" || parts[1] !== "pages") return null;
  if (parts.some((part) => part === "." || part === ".." || part.includes("\\"))) return null;
  return parts.join("/");
}

/**
 * Canonicalize a module specifier against an importer identity (`e2e/...`).
 * Exact segment identity only — no basename/suffix/`includes` authority.
 * @returns {string|null} canonical `e2e/pages/....ts` or null when unrelated/rejected
 */
export function canonicalizeSpecifier(importerIdentity, specifier) {
  if (typeof specifier !== "string") return null;
  if (specifier.includes("\0") || specifier.includes("\\")) return null;
  if (specifier.includes("?") || specifier.includes("#")) return null;
  const trimmed = specifier.trim();
  if (!trimmed) return null;

  let candidate = trimmed;

  // @/* is src/* in frontend/tsconfig.json — never an e2e/pages identity.
  if (candidate.startsWith("@/")) {
    return null;
  } else if (candidate.startsWith("/app/")) {
    candidate = candidate.slice("/app/".length);
  } else if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(candidate)) {
    // Only accept explicit trusted file:///app/e2e/... (three slashes, no authority).
    // Reject http(s), file://evil.com/..., and file://localhost/... (Node clears localhost host).
    if (!candidate.startsWith("file:///app/e2e/")) return null;
    let url;
    try {
      url = new URL(candidate);
    } catch {
      return null;
    }
    if (url.username || url.password || url.search || url.hash) return null;
    if (url.protocol !== "file:") return null;
    if (url.host || url.hostname) return null;
    const pathname = toPosix(url.pathname);
    if (!pathname.startsWith("/app/e2e/")) return null;
    candidate = pathname.slice("/app/".length);
  } else if (candidate.startsWith("/")) {
    return null;
  } else if (candidate.startsWith("./") || candidate.startsWith("../")) {
    const importer = toPosix(importerIdentity);
    if (!importer.startsWith("e2e/")) return null;
    const baseDir = path.posix.dirname(importer);
    const joined = path.posix.normalize(path.posix.join(baseDir, candidate));
    if (!joined.startsWith("e2e/") || joined.split("/").includes("..")) return null;
    candidate = joined;
  } else {
    return null;
  }

  const identity = withTsExtension(candidate);
  const parts = splitSegments(identity);
  if (parts[0] !== "e2e" || parts.includes("..") || parts.includes(".")) return null;
  return parts.join("/");
}

function literalText(node) {
  if (!node) return null;
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) {
    return node.text;
  }
  return null;
}

function isRequireIdentifier(node) {
  return ts.isIdentifier(node) && node.text === "require";
}

/**
 * Collect classified module references from one TS source via AST.
 * Each item: { form, specifier, identity|null, nonliteral?: true, parseError?: true }
 */
export function collectModuleReferences(sourceText, importerIdentity) {
  const sourceFile = ts.createSourceFile(
    importerIdentity || "fixture.ts",
    sourceText,
    ts.ScriptTarget.ES2022,
    true,
    ts.ScriptKind.TS,
  );

  const syntactic = sourceFile.parseDiagnostics || [];
  if (syntactic.length > 0) {
    return [
      {
        form: "parse-error",
        specifier: "",
        identity: null,
        parseError: true,
      },
    ];
  }

  const refs = [];

  function pushRef(form, specifierNode, specifierText) {
    if (specifierText == null) {
      refs.push({
        form: form.startsWith("nonliteral") ? form : `nonliteral-${form}`,
        specifier: "",
        identity: null,
        nonliteral: true,
      });
      return;
    }
    const identity = canonicalizeSpecifier(importerIdentity, specifierText);
    refs.push({
      form,
      specifier: specifierText,
      identity,
    });
  }

  function visit(node) {
    if (ts.isImportDeclaration(node)) {
      const spec = literalText(node.moduleSpecifier);
      if (!node.importClause) {
        pushRef("side-effect-import", node.moduleSpecifier, spec);
      } else if (spec != null && !isImporterRelativeSpecifier(spec)) {
        // Keep identity for exact-target blocking; never a supported consumer.
        pushRef("absolute-static-import", node.moduleSpecifier, spec);
      } else {
        pushRef("static-import", node.moduleSpecifier, spec);
      }
    } else if (ts.isExportDeclaration(node) && node.moduleSpecifier) {
      const spec = literalText(node.moduleSpecifier);
      pushRef("export-from", node.moduleSpecifier, spec);
    } else if (ts.isCallExpression(node)) {
      const expr = node.expression;
      if (expr.kind === ts.SyntaxKind.ImportKeyword) {
        const arg = node.arguments[0];
        const spec = literalText(arg);
        if (spec == null) {
          pushRef("nonliteral-dynamic-import", arg, null);
        } else {
          pushRef("dynamic-import", arg, spec);
        }
      } else if (isRequireIdentifier(expr)) {
        const arg = node.arguments[0];
        const spec = literalText(arg);
        if (spec == null) {
          pushRef("nonliteral-require", arg, null);
        } else {
          pushRef("require", arg, spec);
        }
      }
    }
    ts.forEachChild(node, visit);
  }

  visit(sourceFile);
  return refs;
}

export function listSpecFiles(e2eRoot) {
  let rootStat;
  try {
    rootStat = lstatSync(e2eRoot);
  } catch (error) {
    throw new Error(
      `unreadable e2e scan path: ${e2eRoot}: ${error && error.message ? error.message : error}`,
    );
  }
  if (rootStat.isSymbolicLink()) {
    throw new Error(`e2e root is a symlink: ${e2eRoot}`);
  }
  if (!rootStat.isDirectory()) {
    throw new Error(`e2e root is not a directory: ${e2eRoot}`);
  }

  const out = [];
  const stack = [e2eRoot];
  while (stack.length) {
    const current = stack.pop();
    let entries;
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch (error) {
      throw new Error(
        `unreadable e2e scan path: ${current}: ${error && error.message ? error.message : error}`,
      );
    }
    for (const entry of entries) {
      const full = path.join(current, entry.name);
      let st;
      try {
        st = lstatSync(full);
      } catch (error) {
        throw new Error(
          `unreadable e2e scan path: ${full}: ${error && error.message ? error.message : error}`,
        );
      }
      // Fail closed on any symlink before extension filtering; never follow links.
      if (st.isSymbolicLink()) {
        throw new Error(`symlink in e2e tree: ${full}`);
      }
      if (st.isDirectory()) {
        if (entry.name === "node_modules" || entry.name === "dist") continue;
        stack.push(full);
      } else if (st.isFile() && entry.name.endsWith(".spec.ts")) {
        out.push(full);
      }
    }
  }
  return out.sort();
}

function fileIdentityFromRoot(e2eRoot, absoluteFile) {
  const rel = toPosix(path.relative(e2eRoot, absoluteFile));
  if (rel.startsWith("..") || path.isAbsolute(rel)) return null;
  return `e2e/${rel}`;
}

function assertRawAbsoluteCanonical(absolutePath) {
  if (typeof absolutePath !== "string" || !absolutePath) {
    throw new Error(`enumerated spec path is not a string: ${absolutePath}`);
  }
  if (!path.isAbsolute(absolutePath)) {
    throw new Error(`enumerated spec path is not absolute: ${absolutePath}`);
  }
  // Reject /./ inserts and other spellings that path.relative would collapse.
  if (path.resolve(absolutePath) !== absolutePath) {
    throw new Error(
      `enumerated spec path is noncanonical absolute spelling: ${absolutePath}`,
    );
  }
}

/**
 * Read AST evidence via an O_NOFOLLOW descriptor bound to e2eRoot.
 * Pathname readFile after a prior lstat is not authoritative under TOCTOU.
 */
function readTrustedSpecFile(e2eRoot, absolutePath) {
  assertRawAbsoluteCanonical(absolutePath);
  let rootReal;
  try {
    rootReal = realpathSync(e2eRoot);
  } catch (error) {
    throw new Error(
      `e2e root is unreadable for trusted read: ${e2eRoot}: ${
        error && error.message ? error.message : error
      }`,
    );
  }

  const flags = fsConstants.O_RDONLY | fsConstants.O_NOFOLLOW;
  let fd;
  try {
    fd = openSync(absolutePath, flags);
  } catch (error) {
    const detail = (error && (error.code || error.message)) || error;
    throw new Error(`trusted spec open failed (nofollow): ${absolutePath}: ${detail}`);
  }

  try {
    const st = fstatSync(fd);
    if (!st.isFile()) {
      throw new Error(`trusted spec is not a regular file: ${absolutePath}`);
    }

    // Descriptor identity — fail closed when /proc/self/fd is unavailable.
    let openedReal;
    try {
      openedReal = realpathSync(`/proc/self/fd/${fd}`);
    } catch (error) {
      throw new Error(
        `trusted spec descriptor realpath unavailable (fail closed): ${absolutePath}: ${
          error && error.message ? error.message : error
        }`,
      );
    }
    const rel = toPosix(path.relative(rootReal, openedReal));
    if (!rel || rel.startsWith("..") || path.isAbsolute(rel)) {
      throw new Error(`trusted spec escapes e2e root: ${absolutePath}`);
    }
    return readFileSync(fd, "utf8");
  } finally {
    closeSync(fd);
  }
}

function assertPathInsideRoot(rootAbs, targetAbs, label) {
  let rootReal;
  let targetReal;
  try {
    rootReal = realpathSync(rootAbs);
    targetReal = realpathSync(targetAbs);
  } catch (error) {
    throw new Error(
      `${label} is unreadable: ${targetAbs}: ${error && error.message ? error.message : error}`,
    );
  }
  const rel = toPosix(path.relative(rootReal, targetReal));
  if (!rel || rel.startsWith("..") || path.isAbsolute(rel)) {
    throw new Error(`${label} escapes e2e root: ${targetAbs}`);
  }
  return targetReal;
}

function assertSelectedPageExists(e2eRoot, pageIdentity) {
  if (!pageIdentity.startsWith("e2e/")) {
    throw new Error(`selected page is not canonical: ${pageIdentity}`);
  }
  const absPage = path.join(e2eRoot, pageIdentity.slice("e2e/".length));
  let st;
  try {
    st = lstatSync(absPage);
  } catch {
    throw new Error(`selected page is missing: ${pageIdentity}`);
  }
  if (st.isSymbolicLink()) {
    assertPathInsideRoot(e2eRoot, absPage, `selected page ${pageIdentity}`);
  }
  try {
    st = statSync(absPage);
  } catch {
    throw new Error(`selected page is missing: ${pageIdentity}`);
  }
  if (!st.isFile()) {
    throw new Error(`selected page is not a file: ${pageIdentity}`);
  }
  assertPathInsideRoot(e2eRoot, absPage, `selected page ${pageIdentity}`);
}

/**
 * @param {{
 *   pages: string[],
 *   e2eRoot: string,
 *   listSpecs?: () => string[],
 *   beforeTrustedRead?: (absolutePath: string) => void,
 * }} options
 */
export function verifyPages(options) {
  const e2eRoot = options.e2eRoot;
  const normalized = options.pages.map((page) => normalizePageIdentity(page));
  if (normalized.some((page) => !page) || normalized.length !== options.pages.length) {
    return {
      ok: false,
      error: "one or more --page values are not e2e/pages/*.ts identities",
      pages: [],
    };
  }
  const pages = normalized;

  try {
    for (const page of pages) {
      assertSelectedPageExists(e2eRoot, page);
    }
  } catch (error) {
    return {
      ok: false,
      error: String(error && error.message ? error.message : error),
      pages: [],
    };
  }

  let specAbsPaths;
  try {
    specAbsPaths = options.listSpecs
      ? options.listSpecs()
      : listSpecFiles(e2eRoot);
  } catch (error) {
    return {
      ok: false,
      error: String(error && error.message ? error.message : error),
      pages: [],
    };
  }
  if (!Array.isArray(specAbsPaths)) {
    return {
      ok: false,
      error: "listSpecs must return an array of absolute paths",
      pages: [],
    };
  }

  const pageResults = [];
  let ok = true;

  for (const page of pages) {
    const consumers = [];
    const blocking = [];
    const seenConsumer = new Set();
    const seenBlocking = new Set();

    for (const abs of specAbsPaths) {
      try {
        assertRawAbsoluteCanonical(abs);
      } catch (error) {
        return {
          ok: false,
          error: String(error && error.message ? error.message : error),
          pages: [],
        };
      }
      const identity = fileIdentityFromRoot(e2eRoot, abs);
      if (!identity) {
        return {
          ok: false,
          error: `enumerated spec path escapes e2e root or is noncanonical: ${abs}`,
          pages: [],
        };
      }
      if (typeof options.beforeTrustedRead === "function") {
        options.beforeTrustedRead(abs);
      }
      let text;
      try {
        // Descriptor-bound read — never pathname readFile for trusted AST evidence.
        text = readTrustedSpecFile(e2eRoot, abs);
      } catch (error) {
        const message = String(error && error.message ? error.message : error);
        // Security / containment / symlink failures fail the whole scan closed.
        if (
          /nofollow|symlink|ELOOP|trusted spec|escapes e2e root|descriptor realpath|regular file|noncanonical/i.test(
            message,
          )
        ) {
          return {
            ok: false,
            error: message,
            pages: [],
          };
        }
        const key = `${identity}|parse-error`;
        if (!seenBlocking.has(key)) {
          seenBlocking.add(key);
          blocking.push({ file: identity, form: "parse-error" });
        }
        continue;
      }

      const refs = collectModuleReferences(text, identity);
      for (const ref of refs) {
        if (ref.parseError || ref.form === "parse-error") {
          const key = `${identity}|parse-error`;
          if (!seenBlocking.has(key)) {
            seenBlocking.add(key);
            blocking.push({ file: identity, form: "parse-error" });
          }
          continue;
        }
        if (ref.nonliteral) {
          const key = `${identity}|${ref.form}`;
          if (!seenBlocking.has(key)) {
            seenBlocking.add(key);
            blocking.push({ file: identity, form: ref.form });
          }
          continue;
        }
        if (!ref.identity || !identitiesEqual(ref.identity, page)) {
          continue;
        }
        if (
          SUPPORTED_FORMS.has(ref.form) &&
          isImporterRelativeSpecifier(ref.specifier)
        ) {
          const key = `${identity}|${ref.form}|${ref.specifier}`;
          if (!seenConsumer.has(key)) {
            seenConsumer.add(key);
            consumers.push({
              file: identity,
              form: ref.form,
              specifier: ref.specifier,
            });
          }
        } else {
          // Exact-target require/absolute/unsupported/unknown → blocking.
          const key = `${identity}|${ref.form}`;
          if (!seenBlocking.has(key)) {
            seenBlocking.add(key);
            blocking.push({ file: identity, form: ref.form });
          }
        }
      }
    }

    if (consumers.length < 1 || blocking.length > 0) ok = false;
    pageResults.push({ page, consumers, blocking });
  }

  if (!pages.length) {
    ok = false;
  }

  return { ok, pages: pageResults };
}

export function parseArgs(argv) {
  const pages = [];
  let e2eRoot = "e2e";
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--page") {
      const value = argv[++i];
      if (!value) throw new Error("missing --page value");
      pages.push(value);
    } else if (arg === "--e2e-root") {
      const value = argv[++i];
      if (!value) throw new Error("missing --e2e-root value");
      e2eRoot = value;
    } else if (arg === "--help" || arg === "-h") {
      return { help: true, pages, e2eRoot };
    } else {
      throw new Error(`unknown argument: ${arg}`);
    }
  }
  return { help: false, pages, e2eRoot };
}

export function main(argv = process.argv.slice(2), io = {}) {
  const writeOut = io.writeOut || ((text) => process.stdout.write(text));
  const writeErr = io.writeErr || ((text) => process.stderr.write(text));
  const exit = io.exit || ((code) => process.exit(code));
  const cwd = io.cwd || process.cwd();

  let args;
  try {
    args = parseArgs(argv);
  } catch (error) {
    const payload = { ok: false, error: String(error.message || error) };
    writeErr(JSON.stringify(payload, null, 2) + "\n");
    exit(2);
    return 2;
  }

  if (args.help || args.pages.length === 0) {
    const payload = {
      ok: false,
      error: "usage: --page e2e/pages/<name>.ts [--page ...] [--e2e-root e2e]",
    };
    writeErr(JSON.stringify(payload, null, 2) + "\n");
    exit(2);
    return 2;
  }

  const absoluteRoot = path.isAbsolute(args.e2eRoot)
    ? args.e2eRoot
    : path.resolve(cwd, args.e2eRoot);

  try {
    if (!statSync(absoluteRoot).isDirectory()) {
      throw new Error(`e2e root is not a directory: ${args.e2eRoot}`);
    }
  } catch (error) {
    const payload = { ok: false, error: String(error.message || error) };
    writeErr(JSON.stringify(payload, null, 2) + "\n");
    exit(2);
    return 2;
  }

  const normalizedPages = args.pages.map((page) => normalizePageIdentity(page));
  if (normalizedPages.some((page) => !page)) {
    const payload = {
      ok: false,
      error: "one or more --page values are not e2e/pages/*.ts identities",
    };
    writeErr(JSON.stringify(payload, null, 2) + "\n");
    exit(2);
    return 2;
  }

  const payload = verifyPages({
    pages: normalizedPages,
    e2eRoot: absoluteRoot,
  });
  const text = JSON.stringify(payload, null, 2) + "\n";
  if (payload.ok) writeOut(text);
  else writeErr(text);
  const code = payload.ok ? 0 : 1;
  exit(code);
  return code;
}

const isDirect =
  process.argv[1] &&
  pathToFileURL(path.resolve(process.argv[1])).href === import.meta.url;

if (isDirect) {
  main();
}

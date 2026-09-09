#!/usr/bin/env node
/**
 * Exact E2E page-object consumer closure via TypeScript AST (not regex).
 *
 * CLI (cwd = frontend root):
 *   node scripts/verify-e2e-page-consumers.mjs --page e2e/pages/foo.ts [--page ...] [--e2e-root e2e]
 */

import { createRequire } from "node:module";
import { readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";

const require = createRequire(import.meta.url);
const ts = require("typescript");

const SUPPORTED_FORMS = new Set(["static-import", "require"]);
const UNSUPPORTED_EXACT_FORMS = new Set([
  "side-effect-import",
  "dynamic-import",
  "export-from",
]);

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
  const parts = splitSegments(identity);
  if (parts.length < 3 || parts[0] !== "e2e" || parts[1] !== "pages") return null;
  if (parts.some((part) => part === "." || part === "..")) return null;
  if (parts.some((part) => part.includes("\\"))) return null;
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

  if (candidate.startsWith("@/")) {
    const rest = candidate.slice(2);
    if (!rest.startsWith("e2e/")) return null;
    candidate = rest;
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

function listSpecFiles(e2eRoot) {
  const out = [];
  const stack = [e2eRoot];
  while (stack.length) {
    const current = stack.pop();
    let entries;
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const entry of entries) {
      const full = path.join(current, entry.name);
      if (entry.isDirectory()) {
        if (entry.name === "node_modules" || entry.name === "dist") continue;
        stack.push(full);
      } else if (entry.isFile() && entry.name.endsWith(".spec.ts")) {
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

/**
 * @param {{ pages: string[], e2eRoot: string, readFile?: (p: string) => string, listSpecs?: () => string[] }} options
 */
export function verifyPages(options) {
  const e2eRoot = options.e2eRoot;
  const pages = options.pages.map((page) => normalizePageIdentity(page)).filter(Boolean);
  const specAbsPaths = options.listSpecs
    ? options.listSpecs()
    : listSpecFiles(e2eRoot);
  const read =
    options.readFile ||
    ((absolutePath) => readFileSync(absolutePath, "utf8"));

  const pageResults = [];
  let ok = true;

  for (const page of pages) {
    const consumers = [];
    const blocking = [];
    const seenConsumer = new Set();
    const seenBlocking = new Set();

    for (const abs of specAbsPaths) {
      const identity = fileIdentityFromRoot(e2eRoot, abs);
      if (!identity) continue;
      let text;
      try {
        text = read(abs);
      } catch {
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
        if (SUPPORTED_FORMS.has(ref.form)) {
          const key = `${identity}|${ref.form}|${ref.specifier}`;
          if (!seenConsumer.has(key)) {
            seenConsumer.add(key);
            consumers.push({
              file: identity,
              form: ref.form,
              specifier: ref.specifier,
            });
          }
        } else if (UNSUPPORTED_EXACT_FORMS.has(ref.form)) {
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

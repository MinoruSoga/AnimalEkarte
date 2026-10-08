import fs from "node:fs";
import path from "node:path";

const ALLOWED_KEYS = new Set([
  "APP_ENV",
  "APP_PORT",
  "DB_NAME",
  "DB_PASSWORD",
  "DB_SSL_MODE",
  "DB_USER",
  "JWT_SECRET",
  "STORAGE_TYPE",
  "TRUSTED_PROXY_CIDR",
]);
const REQUIRED_KEYS = ["DB_NAME", "DB_PASSWORD", "DB_USER", "JWT_SECRET"];

export function validateA4EnvFile({ repo, input }) {
  if (!input) throw new Error("A4_ENV_FILE is required");
  const sensitiveRoot = path.join(repo, "sensitive-local");
  const file = path.resolve(repo, input);
  const rootStat = fs.lstatSync(sensitiveRoot);
  if (!rootStat.isDirectory() || rootStat.isSymbolicLink()) {
    throw new Error("sensitive-local must be a real directory");
  }
  if (!file.startsWith(`${sensitiveRoot}${path.sep}`)) {
    throw new Error("A4_ENV_FILE must be under sensitive-local");
  }
  let parent = sensitiveRoot;
  for (const component of path.relative(sensitiveRoot, path.dirname(file)).split(path.sep)) {
    if (component === "") continue;
    parent = path.join(parent, component);
    const stat = fs.lstatSync(parent);
    if (!stat.isDirectory() || stat.isSymbolicLink()) {
      throw new Error("A4_ENV_FILE parent directories may not be symlinks");
    }
  }
  const stat = fs.lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink() || (stat.mode & 0o077) !== 0) {
    throw new Error("A4_ENV_FILE must be an owner-only regular file");
  }
  const values = new Map();
  for (const rawLine of fs.readFileSync(file, "utf8").split(/\r?\n/)) {
    const line = rawLine.trim();
    if (line === "" || line.startsWith("#")) continue;
    const match = /^([A-Z][A-Z0-9_]*)=(.*)$/.exec(line);
    if (!match || !ALLOWED_KEYS.has(match[1])) {
      throw new Error("A4_ENV_FILE contains a malformed or forbidden key");
    }
    values.set(match[1], match[2]);
  }
  for (const key of REQUIRED_KEYS) {
    if (!values.get(key)) throw new Error(`A4_ENV_FILE is missing ${key}`);
  }
  return file;
}

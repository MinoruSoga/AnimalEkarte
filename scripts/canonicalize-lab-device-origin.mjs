#!/usr/bin/env node
import net from "node:net";

// Canonical contract for macOS packaging inputs: an exact HTTPS origin is
// accepted, then emitted exactly as a browser serializes URL.origin.
const raw = (process.argv[2] ?? "").trim();
if (!/^https:\/\/[^/?#\\]+$/.test(raw) || !/^[\x21-\x7e]+$/.test(raw)) process.exit(1);

try {
  const parsed = new URL(raw);
  const authority = raw.slice("https://".length);
  const bracketEnd = authority.indexOf("]");
  const lastColon = authority.lastIndexOf(":");
  const rawHostname = authority.startsWith("[")
    ? authority.slice(1, bracketEnd)
    : authority.slice(0, lastColon >= 0 ? lastColon : authority.length);
  const numericHostname = /^[0-9.]+$/.test(rawHostname);
  const parsedIPv4 = net.isIP(parsed.hostname) === 4;
  const mappedIPv6 = parsed.hostname.toLowerCase().startsWith("[::ffff:");
  if (
    parsed.protocol !== "https:" ||
    parsed.username !== "" ||
    parsed.password !== "" ||
    parsed.hostname === "" ||
    parsed.hostname.includes("*") ||
    parsed.pathname !== "/" ||
    parsed.search !== "" ||
    parsed.hash !== "" ||
    authority.endsWith(":") ||
    parsed.port === "0" ||
    mappedIPv6 ||
    (parsedIPv4 && rawHostname !== parsed.hostname) ||
    (numericHostname && net.isIP(rawHostname) !== 4)
  ) process.exit(1);
  process.stdout.write(`${parsed.origin}\n`);
} catch {
  process.exit(1);
}

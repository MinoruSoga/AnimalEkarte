#!/usr/bin/env node
import net from "node:net";

// Shared host contract for macOS packaging and the Go runtime: canonical dotted
// IPv4, non-mapped IPv6, or strict ASCII DNS. Emit the exact canonical origin.
function isIPv4MappedIPv6(hostname) {
  if (!hostname.startsWith("[") || !hostname.endsWith("]")) return false;
  const address = hostname.slice(1, -1);
  if (net.isIP(address) !== 6) return false;

  const halves = address.split("::");
  if (halves.length > 2) return false;
  const left = halves[0] === "" ? [] : halves[0].split(":");
  const right = halves.length === 1 || halves[1] === "" ? [] : halves[1].split(":");
  const omitted = halves.length === 2 ? 8 - left.length - right.length : 0;
  const groups = [...left, ...Array(omitted).fill("0"), ...right].map((part) => Number.parseInt(part, 16));
  return groups.length === 8 && groups.slice(0, 5).every((part) => part === 0) && groups[5] === 0xffff;
}

function isStrictDNSHostname(hostname) {
  if (hostname.length > 253) return false;
  const labels = hostname.split(".");
  if (labels.some((label) => label.length === 0 || label.length > 63 || !/^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/.test(label))) return false;
  const terminal = labels.at(-1);
  return !/^[0-9]+$/.test(terminal) && !/^0x[0-9a-f]*$/.test(terminal);
}

const raw = (process.argv[2] ?? "").trim();
if (raw.includes("%") || !/^https:\/\/[^/?#\\]+$/.test(raw) || !/^[\x21-\x7e]+$/.test(raw)) process.exit(1);

const authority = raw.slice("https://".length);
if (authority.includes("@")) process.exit(1);

try {
  const parsed = new URL(raw);
  const bracketEnd = authority.indexOf("]");
  const lastColon = authority.lastIndexOf(":");
  const rawHostname = authority.startsWith("[")
    ? authority.slice(1, bracketEnd)
    : authority.slice(0, lastColon >= 0 ? lastColon : authority.length).toLowerCase();
  const parsedIPv4 = net.isIP(parsed.hostname) === 4;
  const parsedIPv6 = parsed.hostname.startsWith("[") && net.isIP(parsed.hostname.slice(1, -1)) === 6;
  const supportedHost = parsedIPv6
    ? !isIPv4MappedIPv6(parsed.hostname.toLowerCase())
    : parsedIPv4
      ? rawHostname === parsed.hostname
      : isStrictDNSHostname(parsed.hostname.toLowerCase());
  if (
    parsed.protocol !== "https:" ||
    parsed.username !== "" ||
    parsed.password !== "" ||
    parsed.hostname === "" ||
    parsed.pathname !== "/" ||
    parsed.search !== "" ||
    parsed.hash !== "" ||
    authority.endsWith(":") ||
    parsed.port === "0" ||
    !supportedHost
  ) process.exit(1);
  process.stdout.write(`${parsed.origin}\n`);
} catch {
  process.exit(1);
}

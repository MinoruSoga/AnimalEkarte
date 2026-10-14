import { readFileSync } from "node:fs";

const REQUIRED_SECRET = "LAB_DEVICE_AGENT_CONSUMER_TOKEN";

try {
  const secrets = JSON.parse(readFileSync(0, "utf8"));
  if (
    !Array.isArray(secrets) ||
    !secrets.every(
      (entry) =>
        entry !== null &&
        typeof entry === "object" &&
        !Array.isArray(entry) &&
        typeof entry.name === "string",
    )
  ) {
    throw new Error();
  }
  if (!secrets.some((entry) => entry.name === REQUIRED_SECRET)) {
    process.stderr.write(
      `Missing required remote secret: ${REQUIRED_SECRET}\n`,
    );
    process.exitCode = 1;
  }
} catch {
  process.stderr.write("Unable to verify required remote secret names\n");
  process.exitCode = 1;
}

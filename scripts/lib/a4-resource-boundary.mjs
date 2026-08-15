const PROJECT_RE = /^animalekarte-a4-[a-z0-9][a-z0-9-]{0,39}$/;
const RUN_ID_RE = /^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/;
const COMMIT_RE = /^[a-f0-9]{40}$/;

function requireBoundary(condition, message) {
  if (!condition) throw new Error(`A4 resource boundary violation: ${message}`);
}

export function validateA4ResourceIdentity({ project, runId, revision }) {
  requireBoundary(PROJECT_RE.test(project ?? ""), "project name is invalid");
  requireBoundary(RUN_ID_RE.test(runId ?? ""), "run ID is invalid");
  requireBoundary(COMMIT_RE.test(revision ?? ""), "release commit is invalid");
  return { project, runId, revision };
}

export function assertFreshA4Resources({ containers, network, volumes }) {
  requireBoundary(containers.length === 0, "project containers already exist");
  requireBoundary(network === undefined, "project network already exists");
  requireBoundary(volumes.every((volume) => volume === undefined), "project volume already exists");
}

export function assertDisposableA4Resources({ containers, network, volumes, project, runId }) {
  requireBoundary(
    containers.length > 0 || network !== undefined || volumes.some((volume) => volume !== undefined),
    "no matching A4 resources exist",
  );
  for (const container of containers) {
    const labels = container.Config?.Labels ?? {};
    requireBoundary(labels["com.docker.compose.project"] === project, "container project label mismatch");
    requireBoundary(labels["com.animalekarte.a4.disposable"] === "true", "container is not disposable");
    requireBoundary(labels["com.animalekarte.a4.run-id"] === runId, "container run ID mismatch");
  }
  for (const [kind, resource] of [
    ["network", network],
    ...volumes.map((volume) => ["volume", volume]),
  ]) {
    if (resource === undefined) continue;
    const labels = resource.Labels ?? {};
    requireBoundary(labels["com.docker.compose.project"] === project, `${kind} project label mismatch`);
    requireBoundary(labels["com.animalekarte.a4.disposable"] === "true", `${kind} is not disposable`);
    requireBoundary(labels["com.animalekarte.a4.run-id"] === runId, `${kind} run ID mismatch`);
  }
}

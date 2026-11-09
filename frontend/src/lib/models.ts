// Ids differ by server for the same model: Ollama uses `name:tag`, `name:cloud`, `name-cloud`,
// `name:latest`; Lemonade exposes Ollama-cloud models as `ollama.name:tag`.

/** isCloudModel reports whether id names a hosted model on Ollama or Lemonade. */
export function isCloudModel(id: string): boolean {
  return id.endsWith(':cloud') || id.endsWith('-cloud') || id.startsWith('ollama.');
}

/** modelKey normalizes an id so the same model matches across servers. */
export function modelKey(id: string): string {
  return id
    .toLowerCase()
    .replace(/^ollama\./, '')
    .replace(/(:cloud|-cloud|:latest)$/, '');
}

/** normalizeModels coerces an /api/models response into a string list. */
export function normalizeModels(raw: unknown): string[] {
  return Array.isArray(raw)
    ? raw.filter((m): m is string => typeof m === 'string' && m !== '')
    : [];
}

/** pickModel chooses the initial model: saved id, its cross-server equivalent, a cloud model, then first. */
export function pickModel(models: string[], saved: string | null): string {
  if (saved) {
    if (models.includes(saved)) return saved;
    const key = modelKey(saved);
    const same = models.find((m) => modelKey(m) === key);
    if (same) return same;
  }
  // Cloud models answer in seconds; local CPU inference times out on the agent prompt.
  return models.find(isCloudModel) ?? models[0] ?? '';
}

async function getJSON(url) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
export const api = {
  actions: () => getJSON("/api/actions"),
  skillAreas: () => getJSON("/api/skill-areas"),
  conditions: () => getJSON("/api/conditions"),
  exhaustionLevels: () => getJSON("/api/exhaustion-levels"),
  monsters: () => getJSON("/api/monsters"),
  monster: (id) => getJSON(`/api/monsters/${id}`),
  monsterStats: () => getJSON("/api/monster-stats"),
  spells: () => getJSON("/api/spells"),
  spell: (id) => getJSON(`/api/spells/${id}`),
  maps: () => getJSON("/api/maps"),
  createMap: async (name, image) => {
    const body = new FormData();
    body.set("name", name);
    body.set("image", image);
    const res = await fetch("/api/maps", { method: "POST", body });
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },
  markers: (mapId) => getJSON(`/api/maps/${encodeURIComponent(mapId)}/markers`),
  createMarker: async (mapId, marker) => {
    const res = await fetch(`/api/maps/${encodeURIComponent(mapId)}/markers`, {
      method: "POST",
      body: JSON.stringify(marker),
    });
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },
  updateMarker: async (mapId, marker) => {
    const res = await fetch(
      `/api/maps/${encodeURIComponent(mapId)}/markers/${marker.i}`,
      { method: "PUT", body: JSON.stringify(marker) },
    );
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },
  deleteMarker: async (mapId, i) => {
    const res = await fetch(
      `/api/maps/${encodeURIComponent(mapId)}/markers/${i}`,
      { method: "DELETE" },
    );
    if (!res.ok) throw new Error(await res.text());
  },
  search: (q) => getJSON(`/api/search?q=${encodeURIComponent(q)}`),
  chat: (q, model) =>
    getJSON(
      `/api/chat?q=${encodeURIComponent(q)}&model=${encodeURIComponent(model)}`,
    ),
  models: () => getJSON("/api/models"),
  notes: () => getJSON("/api/notes"),
  note: (name) => getJSON(`/api/notes/${encodeURIComponent(name)}`),
  saveNote: async (name, content) => {
    const res = await fetch(`/api/notes/${encodeURIComponent(name)}`, {
      method: "PUT",
      body: content,
    });
    if (!res.ok) throw new Error(await res.text());
  },
  initiativePresets: () => getJSON("/api/initiative-presets"),
  saveInitiativePreset: async (name, combatants) => {
    const res = await fetch(
      `/api/initiative-presets/${encodeURIComponent(name)}`,
      {
        method: "PUT",
        body: JSON.stringify(combatants),
      },
    );
    if (!res.ok) throw new Error(await res.text());
  },
  deleteInitiativePreset: async (name) => {
    const res = await fetch(
      `/api/initiative-presets/${encodeURIComponent(name)}`,
      { method: "DELETE" },
    );
    if (!res.ok) throw new Error(await res.text());
  },
};

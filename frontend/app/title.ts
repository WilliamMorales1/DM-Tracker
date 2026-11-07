// Sets document.title from the active tab name and, if focused, the item
// currently open within it (a monster, spell, or reference table).
export function setTitle(...parts) {
  document.title = [...parts.filter(Boolean), "DM Tracker"].join(" - ");
}

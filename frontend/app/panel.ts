const registry = new WeakMap();
export function registerPanel(root, meta) {
  registry.set(root, meta);
  return root;
}
export function getPanelMeta(root) {
  return registry.get(root);
}

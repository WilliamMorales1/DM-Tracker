import '@/styles/index.css';
import { h, mount } from '@/lib/dom';
import { setTitle } from '@/lib/title';
import { getPanelMeta } from '@/lib/panel';
import { monstersPanel } from '@/panels/monsters';
import { spellsPanel } from '@/panels/spells';
import { mapsPanel } from '@/panels/maps';
import { chatPanel } from '@/panels/chat';
import { searchPanel } from '@/panels/search';
import { notesPanel } from '@/panels/notes';
import { initiativePanel } from '@/panels/initiative';
import { refsPanel } from '@/panels/references';
const tabs = [
  { name: 'Notes', path: 'notes', load: notesPanel },
  { name: 'Monsters', path: 'monsters', load: monstersPanel },
  { name: 'Spells', path: 'spells', load: spellsPanel },
  { name: 'Maps', path: 'maps', load: mapsPanel },
  {
    name: 'Initiative',
    path: 'initiative',
    load: async () => initiativePanel(),
  },
  { name: 'Ask Agent', path: 'chat', load: async () => chatPanel() },
  { name: 'Lore Search', path: 'search', load: async () => searchPanel() },
  { name: 'References', path: 'references', load: async () => refsPanel() },
];
const root = document.getElementById('root');
const header = h('header', { className: 'bg-black py-3 mb-4 border-b border-gray-600' }, [
  h('div', { className: 'container' }, [
    h('h1', { className: 'text-5xl font-bold text-yellow-400 mb-0' }, ['DM Tracker']),
  ]),
]);
const navList = h('ul', { className: 'nav-tabs mb-3' }, []);
const panel = h('div', { className: 'bg-gray-600/10 rounded p-3' }, []);
const panelCache = new Map();
function updateURL(path) {
  const url = new URL(window.location.href);
  url.searchParams.set('tab', path);
  window.history.replaceState(null, '', url);
}
async function activate(path, updateHistory = true) {
  if (!tabs.some((tab) => tab.path === path)) path = tabs[0].path;
  if (updateHistory) updateURL(path);
  const currentChild = panel.firstChild;
  if (currentChild) await getPanelMeta(currentChild)?.saveIfDirty?.();
  navList.querySelectorAll('button').forEach((btn) => {
    const isActive = btn.dataset.path === path;
    btn.className = `nav-link ${isActive ? 'active' : 'text-gray-100'}`;
  });
  const tab = tabs.find((t) => t.path === path);
  const cached = panelCache.get(path);
  if (cached) {
    mount(panel, cached);
    setTitle(getPanelTitle(cached), tab.name);
    return;
  }
  panel.innerHTML = '<p>Loading...</p>';
  setTitle(tab.name);
  try {
    const node = await tab.load();
    panelCache.set(path, node);
    mount(panel, node);
    setTitle(getPanelTitle(node), tab.name);
  } catch (err) {
    mount(panel, h('p', { className: 'text-red-500' }, [String(err)]));
  }
}
function getPanelTitle(node) {
  return getPanelMeta(node)?.getTitle?.() ?? null;
}
for (const t of tabs) {
  const btn = h('button', { 'data-path': t.path, onclick: () => activate(t.path) }, [t.name]);
  navList.append(h('li', {}, [btn]));
}
const container = h('div', { className: 'container' }, [navList, panel]);
root.append(header, container);
window.addEventListener('popstate', () => {
  activate(new URLSearchParams(window.location.search).get('tab') || tabs[0].path, false);
});
activate(new URLSearchParams(window.location.search).get('tab') || tabs[0].path);

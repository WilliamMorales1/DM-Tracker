import { api } from '@/api';
import { h, mount } from '@/lib/dom';
import { setTitle } from '@/lib/title';
import { registerPanel } from '@/lib/panel';
import { autocomplete } from '@/lib/autocomplete';
function statRow(label, value) {
  if (!value) return null;
  return h('div', {}, [h('strong', {}, [`${label}. `]), value]);
}
function levelSchoolLine(s) {
  const level = s.level === 0 ? 'Cantrip' : `${s.level}${ordinal(s.level)}-level`;
  const school = s.school.toLowerCase();
  const ritual = s.ritual ? ' (ritual)' : '';
  return s.level === 0 ? `${school} cantrip${ritual}` : `${level} ${school}${ritual}`;
}
function ordinal(n) {
  if (n === 1) return 'st';
  if (n === 2) return 'nd';
  if (n === 3) return 'rd';
  return 'th';
}
function spellStatBlock(s) {
  return h('div', { className: 'card' }, [
    h('div', { className: 'card-body' }, [
      h('h3', { className: 'text-yellow-400 mb-0 text-xl font-bold' }, [s.name]),
      h('div', { className: 'italic text-gray-400 mb-2' }, [levelSchoolLine(s)]),
      statRow('Casting Time', s.castingTime),
      statRow('Range', s.range),
      statRow('Components', s.components),
      statRow('Duration', s.concentration ? `Concentration, ${s.duration}` : s.duration),
      statRow('Classes', s.classes),
      h('hr', { className: 'border-gray-600 my-3' }, []),
      s.description ? h('div', { className: 'whitespace-pre-line' }, [s.description]) : null,
      s.higherLevel
        ? h('p', { className: 'mt-2' }, [
            h('strong', { className: 'italic' }, ['At Higher Levels. ']),
            s.higherLevel,
          ])
        : null,
      h('div', { className: 'text-sm text-gray-400 mt-3' }, [
        s.source ? `Source: ${s.source}` : '',
      ]),
    ]),
  ]);
}
export async function spellsPanel() {
  const list = await api.spells();
  if (list.length === 0) {
    return h('div', { className: 'text-gray-400' }, [
      'No spells found. Run: go run ./cmd/ingest-5etools',
    ]);
  }
  const detail = h('div', { className: 'mt-3' }, []);
  const schools = Array.from(new Set(list.map((s) => s.school).filter(Boolean))).sort(
    (a: string, b: string) => a.localeCompare(b),
  );
  const sources = Array.from(new Set(list.map((s) => s.source).filter(Boolean))).sort(
    (a: string, b: string) => a.localeCompare(b),
  );
  let schoolFilter = '';
  let sourceFilter = '';
  let levelFilter = '';
  let selectSeq = 0;
  let selectedName = null;
  const url = new URL(window.location.href);
  const savedSearch = url.searchParams.get('spellQuery') || '';
  async function selectSpell(s) {
    const seq = ++selectSeq;
    detail.innerHTML = '';
    detail.append(h('div', { className: 'text-gray-400' }, ['Loading...']));
    const full = await api.spell(s.id);
    if (seq !== selectSeq) return;
    mount(detail, spellStatBlock(full));
    selectedName = full.name;
    const url = new URL(window.location.href);
    url.searchParams.set('spell', full.name);
    window.history.replaceState(null, '', url);
    setTitle(selectedName, 'Spells');
  }
  const { input: nameInput, refresh: refreshSuggestions } = autocomplete({
    placeholder: 'Type a spell name...',
    suggestionClass: 'spell-suggestions',
    label: (s) => s.name,
    search: (query) => {
      const q = query.toLowerCase();
      return list
        .filter(
          (s) =>
            (schoolFilter === '' || s.school === schoolFilter) &&
            (sourceFilter === '' || s.source === sourceFilter) &&
            (levelFilter === '' || String(s.level) === levelFilter) &&
            (q === '' || s.name.toLowerCase().includes(q)),
        )
        .slice()
        .sort((a, b) => a.name.localeCompare(b.name));
    },
    renderItem: (s) => [
      h('span', {}, [s.name]),
      h('span', { className: 'text-gray-400 text-sm whitespace-nowrap' }, [
        [s.level === 0 ? 'Cantrip' : `Lvl ${s.level}`, s.school].filter(Boolean).join(' · '),
      ]),
    ],
    onSelect: selectSpell,
  });
  nameInput.value = savedSearch;
  nameInput.addEventListener('input', () => {
    const url = new URL(window.location.href);
    if (nameInput.value) url.searchParams.set('spellQuery', nameInput.value);
    else url.searchParams.delete('spellQuery');
    url.searchParams.delete('spell');
    window.history.replaceState(null, '', url);
  });
  const schoolSelect = h(
    'select',
    {
      className: 'form-select w-auto',
      'aria-label': 'Filter spells by school',
      onchange: (e) => {
        schoolFilter = e.target.value;
        refreshSuggestions();
      },
    },
    [
      h('option', { value: '' }, ['All Schools']),
      ...schools.map((sc) => h('option', { value: sc }, [sc])),
    ],
  );
  const levelSelect = h(
    'select',
    {
      className: 'form-select w-auto',
      'aria-label': 'Filter spells by level',
      onchange: (e) => {
        levelFilter = e.target.value;
        refreshSuggestions();
      },
    },
    [
      h('option', { value: '' }, ['All Levels']),
      ...[0, 1, 2, 3, 4, 5, 6, 7, 8, 9].map((lvl) =>
        h('option', { value: String(lvl) }, [lvl === 0 ? 'Cantrip' : `Level ${lvl}`]),
      ),
    ],
  );
  const sourceSelect = h(
    'select',
    {
      className: 'form-select w-auto',
      'aria-label': 'Filter spells by source',
      onchange: (e) => {
        sourceFilter = e.target.value;
        refreshSuggestions();
      },
    },
    [
      h('option', { value: '' }, ['All Sources']),
      ...sources.map((source) => h('option', { value: source }, [source])),
    ],
  );
  const root = h('div', {}, [
    h('div', { className: 'flex gap-2 mb-2' }, [
      nameInput,
      schoolSelect,
      levelSelect,
      sourceSelect,
    ]),
    detail,
  ]);
  const savedSpell = url.searchParams.get('spell');
  if (savedSpell) {
    const match = list.find((spell) => spell.name === savedSpell);
    if (match) void selectSpell(match);
  }
  return registerPanel(root, { getTitle: () => selectedName });
}

import { api } from '@/api';
import { h } from '@/lib/dom';
import { autocomplete } from '@/lib/autocomplete';
// Monster stats are an immutable lookup fetched once and shared by every
// initiative panel instance, unlike the combat/preset state below which is
// specific to a single panel instance.
let monstersLoaded = false;
const monsterMap = new Map();
async function loadMonsters() {
  if (monstersLoaded) return;
  const list = await api.monsterStats();
  for (const m of list) monsterMap.set(m.name, m);
  monstersLoaded = true;
}
function dexMod(dex) {
  return Math.floor((dex - 10) / 2);
}
// Preset names carry annotations like "(CR 16, scaled)" that are not in the bestiary.
function baseMonsterName(name) {
  return name.replace(/\s*\([^)]*\)\s*$/, '');
}
export async function initiativePanel() {
  await loadMonsters();
  const savedCombat = (() => {
    try {
      return JSON.parse(new URLSearchParams(window.location.search).get('combat') || '{}');
    } catch {
      return {};
    }
  })();
  let combatants = Array.isArray(savedCombat.combatants) ? savedCombat.combatants : [];
  let nextId = Math.max(0, ...combatants.map((combatant) => Number(combatant.id) || 0)) + 1;
  let round = Number(savedCombat.round) || 1;
  let turn = Number(savedCombat.turn) || 0;
  let presets = [];
  let editingId = null;
  async function loadPresets() {
    presets = await api.initiativePresets();
  }
  await loadPresets();
  const root = h('div', {}, []);
  let selectedMonster = null;
  let selectedPreset = '';
  function sorted() {
    return [...combatants].sort((a, b) => b.init - a.init);
  }
  function render() {
    const url = new URL(window.location.href);
    if (combatants.length) {
      url.searchParams.set('combat', JSON.stringify({ combatants, round, turn }));
    } else {
      url.searchParams.delete('combat');
    }
    window.history.replaceState(null, '', url);
    root.innerHTML = '';
    document.querySelectorAll('.initiative-suggestions').forEach((el) => el.remove());
    const order = sorted();
    const roundDisplay = h(
      'button',
      {
        type: 'button',
        className: 'btn btn-warning font-bold pointer-events-none',
        disabled: true,
      },
      [`Round ${round}`],
    );
    const prevBtn = h(
      'button',
      {
        type: 'button',
        className: 'btn btn-outline-warning',
        onclick: () => {
          if (order.length === 0) return;
          turn--;
          if (turn < 0) {
            turn = Math.max(order.length - 1, 0);
            round = Math.max(round - 1, 1);
          }
          render();
        },
      },
      ['Prev Turn'],
    );
    const nextBtn = h(
      'button',
      {
        type: 'button',
        className: 'btn btn-warning',
        onclick: () => {
          if (order.length === 0) return;
          turn++;
          if (turn >= order.length) {
            turn = 0;
            round++;
          }
          render();
        },
      },
      ['Next Turn'],
    );
    const resetBtn = h(
      'button',
      {
        type: 'button',
        className: 'btn btn-outline-danger',
        onclick: () => {
          round = 1;
          turn = 0;
          render();
        },
      },
      ['Reset Rounds'],
    );
    const controls = h('div', { className: 'flex items-stretch gap-2 mb-3' }, [
      roundDisplay,
      prevBtn,
      nextBtn,
      resetBtn,
    ]);
    const presetSelect = h(
      'select',
      {
        className: 'form-select w-[220px]',
        'aria-label': 'Select a preset battle',
        onchange: (e) => {
          selectedPreset = e.target.value;
          render();
        },
      },
      [
        h('option', { value: '' }, ['Preset battle...']),
        ...presets.map((p) =>
          h('option', { value: p.name, selected: p.name === selectedPreset }, [p.name]),
        ),
      ],
    );
    const loadPresetBtn = h(
      'button',
      {
        type: 'button',
        className: 'btn btn-outline-warning',
        disabled: !selectedPreset,
        onclick: () => {
          const preset = presets.find((p) => p.name === selectedPreset);
          if (!preset) return;
          combatants = preset.combatants.map((c) => {
            const monster = monsterMap.get(c.name) ?? monsterMap.get(baseMonsterName(c.name));
            const mod = monster ? dexMod(monster.dex) : 0;
            return {
              id: nextId++,
              name: c.name,
              init: Math.floor(Math.random() * 20) + 1 + mod,
              hp: c.hp,
              maxHp: c.hp,
              ac: c.ac,
            };
          });
          round = 1;
          turn = 0;
          render();
        },
      },
      ['Load'],
    );
    const savePresetBtn = h(
      'button',
      {
        type: 'button',
        className: 'btn btn-outline-success',
        onclick: async () => {
          const name = window.prompt('Save battle as:', selectedPreset || '');
          if (!name) return;
          await api.saveInitiativePreset(
            name,
            order.map((c) => ({ name: c.name, hp: c.hp, ac: c.ac })),
          );
          await loadPresets();
          selectedPreset = name;
          render();
        },
      },
      ['Save'],
    );
    const deletePresetBtn = h(
      'button',
      {
        type: 'button',
        className: 'btn btn-outline-danger',
        disabled: !selectedPreset,
        onclick: async () => {
          if (!selectedPreset) return;
          if (!window.confirm(`Delete preset "${selectedPreset}"?`)) return;
          await api.deleteInitiativePreset(selectedPreset);
          selectedPreset = '';
          await loadPresets();
          render();
        },
      },
      ['Delete'],
    );
    const presetControls = h('div', { className: 'flex items-stretch gap-2 mb-3' }, [
      presetSelect,
      loadPresetBtn,
      savePresetBtn,
      deletePresetBtn,
    ]);
    const rows = order.map((c, idx) => {
      const isActive = idx === turn;
      if (c.id === editingId) {
        const editInit = h('input', {
          type: 'number',
          className: 'form-control w-[70px]',
          value: String(c.init),
          'aria-label': `Edit initiative for ${c.name}`,
        });
        const editName = h('input', {
          type: 'text',
          className: 'form-control w-[140px]',
          value: c.name,
          'aria-label': `Edit name for ${c.name}`,
        });
        const editAc = h('input', {
          type: 'number',
          className: 'form-control w-[70px]',
          value: String(c.ac),
          'aria-label': `Edit armor class for ${c.name}`,
        });
        const editHp = h('input', {
          type: 'number',
          className: 'form-control w-[70px]',
          value: String(c.hp),
          'aria-label': `Edit hit points for ${c.name}`,
        });
        const editMaxHp = h('input', {
          type: 'number',
          className: 'form-control w-[70px]',
          value: String(c.maxHp),
          'aria-label': `Edit max hit points for ${c.name}`,
        });
        function commitEdit() {
          const name = editName.value.trim();
          if (name) c.name = name;
          c.init = Number(editInit.value) || 0;
          c.ac = Number(editAc.value) || 0;
          c.maxHp = Number(editMaxHp.value) || 0;
          c.hp = Math.min(Number(editHp.value) || 0, c.maxHp);
        }
        const editDmgInput = h('input', {
          type: 'number',
          className: 'form-control w-[70px]',
          placeholder: 'Amount',
          'aria-label': `Damage or heal amount for ${c.name}`,
        });
        const editApplyDmg = h(
          'button',
          {
            type: 'button',
            className: 'btn btn-sm btn-outline-danger',
            onclick: () => {
              commitEdit();
              c.hp = Math.max(0, c.hp - (Number(editDmgInput.value) || 0));
              render();
            },
          },
          ['Dmg'],
        );
        const editApplyHeal = h(
          'button',
          {
            type: 'button',
            className: 'btn btn-sm btn-outline-success',
            onclick: () => {
              commitEdit();
              c.hp = Math.min(c.maxHp, c.hp + (Number(editDmgInput.value) || 0));
              render();
            },
          },
          ['Heal'],
        );
        const doneBtn = h(
          'button',
          {
            type: 'button',
            className: 'btn btn-sm btn-outline-secondary',
            onclick: () => {
              commitEdit();
              editingId = null;
              render();
            },
          },
          ['Done'],
        );
        const deleteBtn = h(
          'button',
          {
            type: 'button',
            className: 'btn btn-sm btn-outline-danger',
            onclick: () => {
              combatants = combatants.filter((x) => x.id !== c.id);
              editingId = null;
              if (turn >= combatants.length) turn = 0;
              render();
            },
          },
          ['✕'],
        );
        return h('tr', { className: isActive ? 'table-active' : '' }, [
          h('td', {}, [editInit]),
          h('td', {}, [editName]),
          h('td', {}, [editAc]),
          h('td', {}, [
            h('div', { className: 'flex gap-1 items-center' }, [editHp, ' / ', editMaxHp]),
          ]),
          h('td', {}, [
            h('div', { className: 'flex gap-2 items-center' }, [
              editDmgInput,
              editApplyDmg,
              editApplyHeal,
            ]),
          ]),
          h('td', {}, [h('div', { className: 'flex gap-2 items-center' }, [doneBtn, deleteBtn])]),
        ]);
      }
      const hpDisplay = h('span', { className: c.hp <= 0 ? 'text-red-500 font-bold' : '' }, [
        `${c.hp} / ${c.maxHp}`,
      ]);
      const dmgInput = h('input', {
        type: 'number',
        className: 'form-control w-[70px]',
        placeholder: 'Amount',
        'aria-label': `Damage or heal amount for ${c.name}`,
      });
      const applyDmg = h(
        'button',
        {
          type: 'button',
          className: 'btn btn-sm btn-outline-danger',
          onclick: () => {
            c.hp = Math.max(0, c.hp - (Number(dmgInput.value) || 0));
            dmgInput.value = '';
            render();
          },
        },
        ['Dmg'],
      );
      const applyHeal = h(
        'button',
        {
          type: 'button',
          className: 'btn btn-sm btn-outline-success',
          onclick: () => {
            c.hp = Math.min(c.maxHp, c.hp + (Number(dmgInput.value) || 0));
            dmgInput.value = '';
            render();
          },
        },
        ['Heal'],
      );
      const editBtn = h(
        'button',
        {
          type: 'button',
          className: 'btn btn-sm btn-outline-secondary',
          onclick: () => {
            editingId = c.id;
            render();
          },
        },
        ['Edit'],
      );
      return h('tr', { className: isActive ? 'table-active' : '' }, [
        h('td', {}, [String(c.init)]),
        h('td', {}, [h('strong', {}, [c.name])]),
        h('td', {}, [String(c.ac)]),
        h('td', {}, [hpDisplay]),
        h('td', {}, [
          h('div', { className: 'flex gap-2 items-center' }, [dmgInput, applyDmg, applyHeal]),
        ]),
        h('td', {}, [editBtn]),
      ]);
    });
    function selectMonster(m) {
      selectedMonster = m;
      acInput.value = String(m.ac);
      hpInput.value = String(m.hp);
    }
    const { input: nameInput } = autocomplete({
      placeholder: 'e.g. Goblin',
      suggestionClass: 'initiative-suggestions',
      ariaLabel: 'Combatant name',
      minQueryLength: 2,
      placement: 'above',
      label: (m) => m.name,
      search: (query) => {
        const q = query.toLowerCase();
        return [...monsterMap.values()]
          .filter((m) => m.name.toLowerCase().includes(q))
          .sort((a, b) => a.name.localeCompare(b.name))
          .slice(0, 8);
      },
      renderItem: (m) => [
        h('span', {}, [m.name]),
        h('span', { className: 'text-gray-400 text-sm whitespace-nowrap' }, [
          `AC ${m.ac} · HP ${m.hp}`,
        ]),
      ],
      onSelect: selectMonster,
      onInputChange: (value) => {
        selectedMonster = monsterMap.get(value) ?? null;
      },
    });
    const initInput = h('input', {
      type: 'number',
      className: 'form-control w-[90px]',
      placeholder: '0',
      'aria-label': 'Initiative',
    });
    const acInput = h('input', {
      type: 'number',
      className: 'form-control w-[90px]',
      placeholder: '0',
      'aria-label': 'Armor class',
    });
    const hpInput = h('input', {
      type: 'number',
      className: 'form-control w-[90px]',
      placeholder: '0',
      'aria-label': 'Hit points',
    });
    const rollInitBtn = h(
      'button',
      {
        type: 'button',
        className: 'btn btn-outline-warning flex items-center justify-center px-2',
        title: selectedMonster
          ? `1d20 + ${dexMod(selectedMonster.dex)} (DEX)`
          : '1d20 (select a monster for its DEX mod)',
        onclick: () => {
          const mod = selectedMonster ? dexMod(selectedMonster.dex) : 0;
          initInput.value = String(Math.floor(Math.random() * 20) + 1 + mod);
        },
      },
      [
        h(
          'span',
          {
            className: 'd20-icon',
            style: {
              webkitMaskImage: 'url(/images/d20.png)',
              maskImage: 'url(/images/d20.png)',
            },
          },
          [],
        ),
      ],
    );
    function addCombatant() {
      const name = nameInput.value.trim();
      if (!name) return;
      const hp = Number(hpInput.value) || 0;
      combatants.push({
        id: nextId++,
        name,
        init: Number(initInput.value) || 0,
        hp,
        maxHp: hp,
        ac: Number(acInput.value) || 0,
      });
      selectedMonster = null;
      render();
    }
    const addBtn = h(
      'button',
      { type: 'button', className: 'btn btn-warning', onclick: addCombatant },
      ['Add to Tracker'],
    );
    const addRow = h('tr', {}, [
      h('td', {}, [h('div', { className: 'flex gap-2 items-stretch' }, [initInput, rollInitBtn])]),
      h('td', {}, [nameInput]),
      h('td', {}, [acInput]),
      h('td', {}, [hpInput]),
      h('td', {}, [addBtn]),
      h('td', {}, []),
    ]);
    const table = h('div', { className: 'overflow-x-auto' }, [
      h('table', { className: 'table table-hover w-full' }, [
        h('thead', {}, [
          h('tr', {}, [
            h('th', {}, ['Initiative']),
            h('th', {}, ['Name']),
            h('th', {}, ['AC']),
            h('th', {}, ['HP']),
            h('th', {}, ['Adjust']),
            h('th', {}, ['']),
          ]),
        ]),
        h('tbody', {}, [...rows, addRow]),
      ]),
    ]);
    root.append(presetControls, controls, table);
  }
  render();
  return root;
}

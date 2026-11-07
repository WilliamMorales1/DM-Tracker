import { api } from '../api.js';
import { h } from '../dom.js';

export async function mapsPanel() {
    let maps = await api.maps();
    const select = h('select', { className: 'form-select w-auto', 'aria-label': 'Select a map' }, []);
    const nameInput = h('input', {
        type: 'text',
        className: 'form-control w-[220px]',
        placeholder: 'Map name',
        maxlength: '64',
        pattern: '[A-Za-z0-9_-]+',
        required: true,
        'aria-label': 'Map name',
    });
    const imageInput = h('input', {
        type: 'file',
        className: 'form-control w-auto',
        accept: 'image/png,image/jpeg,image/webp',
        required: true,
        'aria-label': 'Map image',
    });
    const status = h('span', { className: 'text-sm text-gray-400' }, []);
    const image = h('img', {
        className: 'block max-w-full h-auto rounded',
        alt: '',
    });
    const pins = h('div', { className: 'absolute inset-0 pointer-events-none' }, []);
    const stage = h('div', { className: 'relative inline-block max-w-full hidden' }, [image, pins]);
    const addToggle = h('button', { type: 'button', className: 'btn btn-outline-secondary' }, ['Add point']);
    const hint = h('span', { className: 'text-sm text-gray-400' }, []);
    const editor = h('div', { className: 'mb-3 hidden' }, []);
    let markers = [];
    let adding = false;
    let currentId = '';

    function setAdding(on) {
        adding = on;
        addToggle.classList.toggle('active', on);
        stage.style.cursor = on ? 'crosshair' : '';
        hint.textContent = on ? 'Click the map to place a point.' : '';
    }

    function closeEditor() {
        editor.classList.add('hidden');
        editor.innerHTML = '';
    }

    function renderPins() {
        pins.innerHTML = '';
        const w = image.naturalWidth || 1;
        const ht = image.naturalHeight || 1;
        for (const m of markers) {
            pins.append(h('button', {
                type: 'button',
                title: m.title,
                'aria-label': m.title,
                className: 'map-pin',
                style: { left: `${(m.x / w) * 100}%`, top: `${(m.y / ht) * 100}%` },
                onclick: (e) => {
                    e.stopPropagation();
                    openEditor(m);
                },
            }, [String(m.i)]));
        }
    }

    function openEditor(marker) {
        const isNew = marker.i === undefined;
        const title = h('input', {
            type: 'text',
            className: 'form-control',
            maxlength: '100',
            required: true,
            placeholder: 'Title',
            'aria-label': 'Point title',
        });
        title.value = marker.title ?? '';
        const body = h('textarea', {
            className: 'form-control',
            rows: '3',
            maxlength: '500',
            placeholder: 'Description',
            'aria-label': 'Point description',
        });
        body.value = marker.body ?? '';
        const err = h('span', { className: 'map-error' }, []);
        const buttons = [
            h('button', { type: 'submit', className: 'btn btn-outline-warning' }, [isNew ? 'Add' : 'Save']),
            h('button', { type: 'button', className: 'btn btn-outline-secondary', onclick: closeEditor }, ['Cancel']),
        ];
        if (!isNew) {
            buttons.push(h('button', {
                type: 'button',
                className: 'btn btn-outline-danger',
                onclick: async () => {
                    try {
                        await api.deleteMarker(currentId, marker.i);
                        markers = markers.filter((m) => m.i !== marker.i);
                        renderPins();
                        closeEditor();
                    }
                    catch (error) {
                        err.textContent = String(error);
                    }
                },
            }, ['Delete']));
        }
        const form = h('form', {
            className: 'map-editor',
            onsubmit: async (event) => {
                event.preventDefault();
                const next = { ...marker, title: title.value.trim(), body: body.value };
                try {
                    if (isNew) {
                        markers.push(await api.createMarker(currentId, next));
                    }
                    else {
                        const saved = await api.updateMarker(currentId, next);
                        markers = markers.map((m) => (m.i === saved.i ? saved : m));
                    }
                    renderPins();
                    closeEditor();
                }
                catch (error) {
                    err.textContent = String(error);
                }
            },
        }, [title, body, h('div', { className: 'flex gap-2 items-center' }, [...buttons, err])]);
        editor.innerHTML = '';
        editor.append(form);
        editor.classList.remove('hidden');
        title.focus();
    }

    async function loadMarkers(id) {
        markers = [];
        renderPins();
        closeEditor();
        if (!id) return;
        try {
            const loaded = await api.markers(id);
            if (id === currentId) markers = loaded;
        }
        catch (error) {
            status.textContent = String(error);
        }
        renderPins();
    }

    addToggle.addEventListener('click', () => setAdding(!adding));
    image.addEventListener('load', renderPins);
    stage.addEventListener('click', (event) => {
        if (!adding || !currentId) return;
        const rect = image.getBoundingClientRect();
        const x = Math.round(((event.clientX - rect.left) / rect.width) * image.naturalWidth);
        const y = Math.round(((event.clientY - rect.top) / rect.height) * image.naturalHeight);
        setAdding(false);
        openEditor({ x, y });
    });
    const empty = h('p', { className: 'text-gray-400' }, ['No maps yet. Upload an image to add one.']);

    function renderMaps(selected = select.value) {
        select.innerHTML = '';
        select.append(h('option', { value: '' }, ['Select a map...']));
        for (const map of maps)
            select.append(h('option', { value: map.id }, [map.id]));
        select.value = maps.some((map) => map.id === selected) ? selected : maps[0]?.id ?? '';
        const current = maps.find((map) => map.id === select.value);
        stage.classList.toggle('hidden', !current);
        addToggle.classList.toggle('hidden', !current);
        const nextId = current?.id ?? '';
        if (nextId !== currentId) {
            currentId = nextId;
            setAdding(false);
            loadMarkers(nextId);
        }
        empty.classList.toggle('hidden', maps.length > 0);
        if (current) {
            image.src = current.img;
            image.alt = current.id;
        }
    }

    select.addEventListener('change', () => {
        renderMaps(select.value);
        const url = new URL(window.location.href);
        if (select.value) url.searchParams.set('map', select.value);
        else url.searchParams.delete('map');
        window.history.replaceState(null, '', url);
    });
    const form = h('form', {
        className: 'flex flex-wrap gap-2 items-center',
        onsubmit: async (event) => {
            event.preventDefault();
            const submit = form.querySelector('button[type="submit"]');
            submit.disabled = true;
            status.textContent = 'Uploading...';
            try {
                const map = await api.createMap(nameInput.value.trim(), imageInput.files[0]);
                maps.push(map);
                maps.sort((left, right) => left.id.localeCompare(right.id));
                renderMaps(map.id);
                nameInput.value = '';
                imageInput.value = '';
                status.textContent = 'Map added.';
            }
            catch (error) {
                status.textContent = String(error);
            }
            finally {
                submit.disabled = false;
            }
        },
    }, [
        nameInput,
        imageInput,
        h('button', { type: 'submit', className: 'btn btn-outline-warning' }, ['Add Map']),
        status,
    ]);
    const requestedMap = new URLSearchParams(window.location.search).get('map');
    renderMaps(requestedMap || undefined);
    return h('div', {}, [
        h('div', { className: 'flex flex-wrap gap-3 items-center mb-3' }, [select, form]),
        h('div', { className: 'flex gap-3 items-center mb-3' }, [addToggle, hint]),
        editor,
        empty,
        stage,
    ]);
}

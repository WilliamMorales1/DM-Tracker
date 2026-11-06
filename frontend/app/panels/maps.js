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
        className: 'hidden max-w-full h-auto rounded',
        alt: '',
    });
    const empty = h('p', { className: 'text-gray-400' }, ['No maps yet. Upload an image to add one.']);

    function renderMaps(selected = select.value) {
        select.innerHTML = '';
        select.append(h('option', { value: '' }, ['Select a map...']));
        for (const map of maps)
            select.append(h('option', { value: map.id }, [map.id]));
        select.value = maps.some((map) => map.id === selected) ? selected : maps[0]?.id ?? '';
        const current = maps.find((map) => map.id === select.value);
        image.classList.toggle('hidden', !current);
        empty.classList.toggle('hidden', maps.length > 0);
        if (current) {
            image.src = current.img;
            image.alt = current.id;
        }
    }

    select.addEventListener('change', () => renderMaps(select.value));
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
    renderMaps();
    return h('div', {}, [
        h('div', { className: 'flex flex-wrap gap-3 items-center mb-3' }, [select, form]),
        empty,
        image,
    ]);
}

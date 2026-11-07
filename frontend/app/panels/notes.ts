import { api } from '../api.js';
import { h, onSubmitAsync, sanitizeHtml } from '../dom.js';
import { registerPanel } from '../panel.js';
import { marked } from 'marked';
// Minimal editor shim used when local Monaco assets cannot be loaded.
function fallbackMonaco() {
    return {
        editor: {
            create(container, opts) {
                const textarea = document.createElement('textarea');
                textarea.className = 'form-control w-full h-full font-mono resize-none';
                textarea.style.height = '100%';
                textarea.readOnly = !!opts?.readOnly;
                container.appendChild(textarea);
                const listeners = [];
                textarea.addEventListener('input', () => listeners.forEach((cb) => cb()));
                return {
                    getValue: () => textarea.value,
                    setValue: (v) => {
                        textarea.value = v;
                    },
                    updateOptions: (o) => {
                        if (o && 'readOnly' in o)
                            textarea.readOnly = !!o.readOnly;
                    },
                    onDidChangeModelContent: (cb) => {
                        listeners.push(cb);
                    },
                };
            },
        },
    };
}
function loadMonaco() {
    return new Promise((resolve) => {
        const w = window;
        if (w.monaco) {
            resolve(w.monaco);
            return;
        }
        const script = document.createElement('script');
        script.src = '/vendor/monaco/min/vs/loader.js';
        script.onerror = () => resolve(fallbackMonaco());
        script.onload = () => {
            w.require.config({ paths: { vs: '/vendor/monaco/min/vs' } });
            // If local AMD modules fail to load, fall back after a timeout.
            let done = false;
            const timer = setTimeout(() => {
                if (!done)
                    resolve(fallbackMonaco());
            }, 5000);
            w.require(['vs/editor/editor.main'], () => {
                done = true;
                clearTimeout(timer);
                resolve(w.monaco);
            });
        };
        document.head.appendChild(script);
    });
}
export async function notesPanel() {
    const [names, monaco] = await Promise.all([api.notes(), loadMonaco()]);
    const select = h('select', { className: 'form-select w-auto', 'aria-label': 'Select a note' }, [
        h('option', { value: '' }, ['Select a note...']),
        ...names.map((n) => h('option', { value: n }, [n])),
    ]);
    const newNameInput = h('input', {
        type: 'text',
        className: 'form-control w-[160px]',
        placeholder: 'new-note.md',
        'aria-label': 'New note file name',
    });
    const newBtn = h('button', { type: 'submit', className: 'btn btn-outline-warning' }, ['Create']);
    const newStatus = h('span', {}, []);
    const previewBtn = h('button', {
        type: 'button',
        className: 'btn btn-outline-warning',
    }, ['Edit']);
    const newForm = h('form', { className: 'flex gap-2 items-center' }, [
        newNameInput,
        newBtn,
        previewBtn,
        newStatus,
    ]);
    const editorDiv = h('div', {
        className: 'border border-gray-600 rounded h-[500px] hidden',
    }, []);
    const previewDiv = h('div', {
        className: 'border border-gray-600 rounded p-3 overflow-auto markdown-preview h-[500px]',
    }, []);
    const editor = monaco.editor.create(editorDiv, {
        language: 'markdown',
        theme: 'vs-dark',
        automaticLayout: true,
        minimap: { enabled: false },
        wordWrap: 'on',
        readOnly: true,
    });
    let isDirty = false;
    let currentNoteName = '';
    let isPreviewing = true;
    editor.onDidChangeModelContent(() => {
        isDirty = true;
    });
    async function saveIfDirty() {
        if (!isDirty || !currentNoteName)
            return;
        await api.saveNote(currentNoteName, editor.getValue());
        isDirty = false;
    }
    function setMode(preview) {
        isPreviewing = preview;
        if (preview) {
            previewDiv.innerHTML = sanitizeHtml(marked(editor.getValue()));
            editorDiv.classList.add('hidden');
            previewDiv.classList.remove('hidden');
            previewBtn.textContent = 'Edit';
        }
        else {
            editorDiv.classList.remove('hidden');
            previewDiv.classList.add('hidden');
            previewBtn.textContent = 'Preview';
        }
    }
    async function loadNote(name, edit = false) {
        await saveIfDirty();
        currentNoteName = name;
        isDirty = false;
        if (!name) {
            editor.setValue('');
            editor.updateOptions({ readOnly: true });
            setMode(true);
            return;
        }
        const note = await api.note(name);
        editor.setValue(note.content);
        isDirty = false;
        editor.updateOptions({ readOnly: !edit });
        setMode(!edit);
    }
    function togglePreview() {
        if (!currentNoteName)
            return;
        if (!isPreviewing) {
            setMode(true);
            return;
        }
        editor.updateOptions({ readOnly: false });
        setMode(false);
    }
    previewBtn.addEventListener('click', togglePreview);
    select.addEventListener('change', () => loadNote(select.value));
    onSubmitAsync(newForm, newBtn, newStatus, 'Creating...', async () => {
        let name = newNameInput.value.trim();
        if (!name) {
            alert('Name required');
            return;
        }
        if (!name.endsWith('.md'))
            name += '.md';
        if (!/^[A-Za-z0-9_-]+\.md$/.test(name)) {
            alert('Use letters, numbers, _ or - only');
            return;
        }
        await api.saveNote(name, '');
        select.append(h('option', { value: name }, [name]));
        select.value = name;
        newNameInput.value = '';
        await loadNote(name, true);
    });
    const beforeUnloadHandler = () => {
        if (isDirty && currentNoteName) {
            fetch(`/api/notes/${encodeURIComponent(currentNoteName)}`, {
                method: 'PUT',
                body: editor.getValue(),
                keepalive: true,
            });
        }
    };
    window.addEventListener('beforeunload', beforeUnloadHandler);
    const root = h('div', {}, [
        h('div', { className: 'flex gap-2 items-center mb-3' }, [select, newForm]),
        editorDiv,
        previewDiv,
    ]);
    const preferredNote = names.find((name) => name === '00-Run-Sheet.md') ?? names[0];
    if (preferredNote) {
        select.value = preferredNote;
        await loadNote(preferredNote);
    }
    return registerPanel(root, { saveIfDirty });
}

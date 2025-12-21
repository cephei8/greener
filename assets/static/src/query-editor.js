if (typeof Prism !== 'undefined') {
    Prism.languages.greenerQuery = {
        'tag': /#"[^"]*"/,
        'string': /"(?:\\.|[^"\\])*"/,
        'keyword': /\b(?:and|or|offset|limit)\b/i,
        'function': /\b(?:group_by|group)\b/i,
        'identifier': /\b(?:session_id|id|name|status|classname|testsuite|file)\b/i,
        'status': /\b(?:pass|fail|error|skip)\b/i,
        'operator': /[!=]=?/,
        'punctuation': /[(),]/
    };
}

const suggestions = [
    { label: 'session_id', type: 'field', desc: 'Session ID' },
    { label: 'id', type: 'field', desc: 'Testcase ID' },
    { label: 'name', type: 'field', desc: 'Testcase name' },
    { label: 'status', type: 'field', desc: 'Status (pass/fail/error/skip)' },
    { label: 'classname', type: 'field', desc: 'Test class name' },
    { label: 'testsuite', type: 'field', desc: 'Test suite name' },
    { label: 'file', type: 'field', desc: 'File path' },
    { label: 'and', type: 'keyword', desc: 'Logical AND' },
    { label: 'or', type: 'keyword', desc: 'Logical OR' },
    { label: 'offset=', type: 'keyword', desc: 'Skip first N results', insert: 'offset=' },
    { label: 'limit=', type: 'keyword', desc: 'Limit to N results (max 100)', insert: 'limit=' },
    { label: 'group_by()', type: 'function', desc: 'Group results', insert: 'group_by()' },
    { label: 'group = ()', type: 'function', desc: 'Filter by group', insert: 'group = ()' },
    { label: 'pass', type: 'value', desc: 'Passed status' },
    { label: 'fail', type: 'value', desc: 'Failed status' },
    { label: 'error', type: 'value', desc: 'Error status' },
    { label: 'skip', type: 'value', desc: 'Skipped status' },
    { label: '#"label"', type: 'tag', desc: 'Tag/label query', insert: '#""' }
];

function initEditor() {
    const editor = document.querySelector('.query-editor');
    if (!editor) return;

    const container = editor.parentElement;
    const dropdown = document.createElement('div');
    dropdown.className = 'autocomplete-dropdown';
    container.appendChild(dropdown);

    const hiddenInput = document.createElement('input');
    hiddenInput.type = 'hidden';
    hiddenInput.id = 'query-value';
    hiddenInput.name = 'query';
    hiddenInput.value = '';
    container.appendChild(hiddenInput);

    let selectedIndex = -1;

    function highlight() {
        const text = editor.textContent;
        const selection = saveSelection();

        const highlighted = Prism.highlight(text, Prism.languages.greenerQuery, 'greenerQuery');
        editor.innerHTML = highlighted || (text ? text : '');

        restoreSelection(selection);
    }

    function saveSelection() {
        const sel = window.getSelection();
        if (sel.rangeCount === 0) return null;

        const range = sel.getRangeAt(0);
        const preSelectionRange = range.cloneRange();
        preSelectionRange.selectNodeContents(editor);
        preSelectionRange.setEnd(range.startContainer, range.startOffset);

        return {
            start: preSelectionRange.toString().length,
            end: preSelectionRange.toString().length + range.toString().length
        };
    }

    function restoreSelection(saved) {
        if (!saved) return;

        const sel = window.getSelection();
        const range = document.createRange();
        range.selectNodeContents(editor);

        let charIndex = 0;
        let nodeStack = [editor];
        let node, foundStart = false, stop = false;

        while (!stop && (node = nodeStack.pop())) {
            if (node.nodeType === 3) {
                const nextCharIndex = charIndex + node.length;
                if (!foundStart && saved.start >= charIndex && saved.start <= nextCharIndex) {
                    range.setStart(node, saved.start - charIndex);
                    foundStart = true;
                }
                if (foundStart && saved.end >= charIndex && saved.end <= nextCharIndex) {
                    range.setEnd(node, saved.end - charIndex);
                    stop = true;
                }
                charIndex = nextCharIndex;
            } else {
                let i = node.childNodes.length;
                while (i--) {
                    nodeStack.push(node.childNodes[i]);
                }
            }
        }

        sel.removeAllRanges();
        sel.addRange(range);
    }

    function showAutocomplete() {
        const text = editor.textContent;
        const cursorPos = saveSelection()?.start || 0;
        const word = getWordAtCursor(text, cursorPos);

        if (!word || word.length < 1) {
            dropdown.classList.remove('show');
            return;
        }

        const matches = suggestions.filter(s =>
            s.label.toLowerCase().startsWith(word.toLowerCase())
        );

        if (matches.length === 0) {
            dropdown.classList.remove('show');
            return;
        }

        dropdown.innerHTML = matches.map((item, i) => `
            <div class="autocomplete-item" data-index="${i}">
                <span>${item.label}</span>
                <span class="badge">${item.type}</span>
            </div>
        `).join('');

        dropdown.classList.add('show');
        selectedIndex = -1;
    }

    function getWordAtCursor(text, pos) {
        const before = text.substring(0, pos);
        const match = before.match(/[a-z_#"]*$/i);
        return match ? match[0] : '';
    }

    function insertSuggestion(item) {
        const text = editor.textContent;
        const cursorPos = saveSelection()?.start || text.length;
        const word = getWordAtCursor(text, cursorPos);

        const before = text.substring(0, cursorPos - word.length);
        const after = text.substring(cursorPos);
        const insert = item.insert || item.label;

        editor.textContent = before + insert + after;

        const newPos = before.length + insert.length;
        setTimeout(() => {
            editor.focus();
            const range = document.createRange();
            const sel = window.getSelection();
            const textNode = editor.firstChild;
            if (textNode) {
                range.setStart(textNode, Math.min(newPos, textNode.length));
                range.collapse(true);
                sel.removeAllRanges();
                sel.addRange(range);
            }
            highlight();
            dropdown.classList.remove('show');
        }, 0);
    }

    editor.addEventListener('input', () => {
        highlight();
        showAutocomplete();
        hiddenInput.value = editor.textContent || '';
    });

    editor.addEventListener('keydown', (e) => {
        const items = dropdown.querySelectorAll('.autocomplete-item');

        if (dropdown.classList.contains('show')) {
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
                updateSelection(items);
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                selectedIndex = Math.max(selectedIndex - 1, -1);
                updateSelection(items);
            } else if ((e.key === 'Enter' || e.key === 'Tab') && selectedIndex >= 0) {
                e.preventDefault();
                const idx = parseInt(items[selectedIndex].dataset.index);
                insertSuggestion(suggestions.filter(s =>
                    s.label.toLowerCase().startsWith(getWordAtCursor(editor.textContent, saveSelection()?.start || 0).toLowerCase())
                )[idx]);
            } else if (e.key === 'Tab' && items.length > 0) {
                e.preventDefault();
                selectedIndex = 0;
                updateSelection(items);
                const idx = parseInt(items[0].dataset.index);
                insertSuggestion(suggestions.filter(s =>
                    s.label.toLowerCase().startsWith(getWordAtCursor(editor.textContent, saveSelection()?.start || 0).toLowerCase())
                )[idx]);
            } else if (e.key === 'Enter') {
                e.preventDefault();
                dropdown.classList.remove('show');
                document.querySelector('.query-btn')?.click();
            } else if (e.key === 'Escape') {
                dropdown.classList.remove('show');
            }
        } else {
            if (e.key === 'Enter') {
                e.preventDefault();
                document.querySelector('.query-btn')?.click();
            }
        }
    });

    function updateSelection(items) {
        items.forEach((item, i) => {
            item.classList.toggle('selected', i === selectedIndex);
        });
    }

    dropdown.addEventListener('click', (e) => {
        const item = e.target.closest('.autocomplete-item');
        if (item) {
            const idx = parseInt(item.dataset.index);
            const matches = suggestions.filter(s =>
                s.label.toLowerCase().startsWith(getWordAtCursor(editor.textContent, saveSelection()?.start || 0).toLowerCase())
            );
            insertSuggestion(matches[idx]);
        }
    });

    editor.addEventListener('blur', () => {
        setTimeout(() => dropdown.classList.remove('show'), 200);
    });

    const btn = document.querySelector('.query-btn');
    if (btn) {
        btn.addEventListener('click', (e) => {
            hiddenInput.value = editor.textContent || '';
        });
    }

    if (window.initialQuery && window.initialQuery.trim() !== '') {
        editor.textContent = window.initialQuery;
        hiddenInput.value = window.initialQuery;
        highlight();
    } else {
        hiddenInput.value = '';
    }

    console.log('Prism query editor initialized');
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initEditor);
} else {
    initEditor();
}

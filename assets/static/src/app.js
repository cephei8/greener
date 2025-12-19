import htmx from 'htmx.org';
window.htmx = htmx;

import 'htmx-ext-response-targets/dist/response-targets.esm.js';

import Prism from 'prismjs';

window.Prism = Prism;

import './query-editor.js';

console.log('Greener app initialized');
console.log('HTMX version:', htmx.version);
console.log('Prism loaded:', typeof Prism !== 'undefined');

import htmx from "htmx.org";
window.htmx = htmx;

import "htmx-ext-response-targets/dist/response-targets.esm.js";

import Prism from "prismjs";

window.Prism = Prism;

import "./query-editor.js";
import { initIcons } from "./icons.js";

document.addEventListener("DOMContentLoaded", () => {
    initIcons();
});

document.addEventListener("htmx:afterSwap", () => {
    initIcons();
});

document.addEventListener("htmx:afterSettle", () => {
    initIcons();
});

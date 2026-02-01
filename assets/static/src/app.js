import htmx from "htmx.org";
window.htmx = htmx;

import "htmx-ext-response-targets/dist/response-targets.esm.js";

import Prism from "prismjs";

window.Prism = Prism;

import "./query-editor.js";
import { initIcons } from "./icons.js";

document.addEventListener("DOMContentLoaded", () => {
    initIcons();
    initSSE();
});

document.addEventListener("htmx:afterSwap", () => {
    initIcons();
});

document.addEventListener("htmx:afterSettle", () => {
    initIcons();
});

function initSSE() {
    if (window.location.pathname === "/login") {
        return;
    }

    let eventSource = null;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;
    const reconnectDelay = 3000;

    function connect() {
        eventSource = new EventSource("/api/v1/sse/events");

        eventSource.addEventListener("connected", (e) => {
            console.log("SSE connected");
            reconnectAttempts = 0;
        });

        eventSource.addEventListener("mcp-query", (e) => {
            try {
                const data = JSON.parse(e.data);
                handleMCPQuery(data);
            } catch (err) {
                console.error("Failed to parse MCP query event:", err);
            }
        });

        eventSource.onerror = (e) => {
            console.error("SSE connection error");
            eventSource.close();

            if (reconnectAttempts < maxReconnectAttempts) {
                reconnectAttempts++;
                console.log(
                    `Reconnecting SSE (attempt ${reconnectAttempts})...`,
                );
                setTimeout(connect, reconnectDelay);
            }
        };
    }

    connect();

    window.addEventListener("beforeunload", () => {
        if (eventSource) {
            eventSource.close();
        }
    });
}

function handleMCPQuery(data) {
    const { page, query } = data;

    if (window.location.pathname !== page) {
        if (query) {
            window.location.href = page + "?query=" + encodeURIComponent(query);
        } else {
            window.location.href = page;
        }
        return;
    }

    const queryEditor = document.querySelector(".query-editor");
    const queryInput = document.querySelector("#query-value");
    const queryBtn = document.querySelector(".query-btn");

    if (queryEditor && queryInput) {
        queryEditor.textContent = query;
        queryInput.value = query;

        if (window.Prism) {
            const highlighted = Prism.highlight(
                query,
                Prism.languages.greenerQuery,
                "greenerQuery",
            );
            queryEditor.innerHTML = highlighted;
        }

        if (queryBtn) {
            queryBtn.click();
        }
    }
}

function showCopyFeedback(element) {
    const tooltip = document.createElement("div");
    tooltip.className = "copy-tooltip";
    tooltip.textContent = "Copied!";

    const rect = element.getBoundingClientRect();
    tooltip.style.left = rect.left + rect.width / 2 - 30 + "px";
    tooltip.style.top = rect.top - 30 + "px";

    document.body.appendChild(tooltip);

    setTimeout(function () {
        tooltip.classList.add("show");
    }, 10);

    setTimeout(function () {
        tooltip.classList.remove("show");
        setTimeout(function () {
            document.body.removeChild(tooltip);
        }, 200);
    }, 1500);
}

window.copyId = function (id, event) {
    event.preventDefault();
    const textToCopy = '"' + id + '"';
    navigator.clipboard
        .writeText(textToCopy)
        .then(function () {
            showCopyFeedback(event.target);
        })
        .catch(function (err) {
            console.error("Failed to copy: ", err);
        });
};

window.initQueryPage = function (tableId) {
    const queryError = document.getElementById("query-error");
    if (queryError) {
        queryError.textContent = "";
    }

    document.body.addEventListener("htmx:afterSwap", function (event) {
        if (event.detail.target.id === tableId) {
            const queryError = document.getElementById("query-error");
            if (queryError) {
                queryError.textContent = "";
            }

            const queryInput = document.querySelector("#query-value");
            if (queryInput) {
                const queryValue = queryInput.value;
                const url = new URL(window.location);
                if (queryValue && queryValue.trim() !== "") {
                    url.searchParams.set("query", queryValue);
                } else {
                    url.searchParams.delete("query");
                }
                window.history.pushState({}, "", url);
            }
        }
    });
};

import checkCircle from "heroicons/24/solid/check-circle.svg";
import xCircle from "heroicons/24/solid/x-circle.svg";
import questionMarkCircle from "heroicons/24/solid/question-mark-circle.svg";
import exclamationTriangle from "heroicons/24/solid/exclamation-triangle.svg";
import bars3 from "heroicons/24/solid/bars-3.svg";
import arrowLeft from "heroicons/24/solid/arrow-left.svg";

const iconMap = {
    "check-circle": checkCircle,
    "x-circle": xCircle,
    "help-circle": questionMarkCircle,
    "exclamation-triangle": exclamationTriangle,
    "arrow-left": arrowLeft,
    menu: bars3,
};

export function initIcons() {
    const elements = document.querySelectorAll("[data-icon]");

    elements.forEach((element) => {
        const iconName = element.getAttribute("data-icon");
        const iconClass = element.getAttribute("data-icon-class") || "";

        const iconSvg = iconMap[iconName];

        if (!iconSvg) {
            return;
        }

        if (element.querySelector("svg")) {
            return;
        }

        const temp = document.createElement("div");
        temp.innerHTML = iconSvg;
        const svg = temp.querySelector("svg");

        if (!svg) {
            return;
        }

        if (iconClass) {
            svg.setAttribute("class", iconClass);
        }

        element.innerHTML = "";
        element.appendChild(svg);
    });
}

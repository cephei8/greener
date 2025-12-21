module.exports = {
    content: ["./assets/templates/**/*.html"],
    safelist: [
        'alert-warning',
    ],
    theme: {
        extend: {},
    },
    plugins: [require("daisyui")],
    daisyui: {
        themes: ["light", "dark"],
    },
};

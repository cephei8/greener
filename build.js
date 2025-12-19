const esbuild = require("esbuild");
const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");

if (!fs.existsSync("assets/static/dist")) {
    fs.mkdirSync("assets/static/dist", { recursive: true });
}

try {
    execSync(
        "npx tailwindcss -i assets/static/src/app.css -o assets/static/dist/app.css --minify",
        {
            stdio: "inherit",
        },
    );
} catch (error) {
    console.error("Failed to build CSS:", error.message);
    process.exit(1);
}

const buildOptions = {
    entryPoints: ["assets/static/src/app.js"],
    bundle: true,
    minify: true,
    sourcemap: false,
    outfile: "assets/static/dist/app.js",
    platform: "browser",
    target: "es2020",
    define: {
        "process.env.NODE_ENV": '"production"',
    },
};

async function build() {
    try {
        await esbuild.build(buildOptions);
    } catch (error) {
        console.error("Build failed:", error);
        process.exit(1);
    }
}

build();

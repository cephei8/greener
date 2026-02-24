const path = require("path");
const platform = process.platform;
const arch = process.arch;

const addon = require(
    path.join(
        __dirname,
        "prebuilds",
        `${platform}-${arch}`,
        "greener-servermock.node",
    ),
);

class Servermock {
    constructor() {
        this._native = new addon.Servermock();
    }

    fixtureNames() {
        return this._native.fixtureNames();
    }

    fixtureCalls(fixtureName) {
        return this._native.fixtureCalls(fixtureName);
    }

    fixtureResponses(fixtureName) {
        return this._native.fixtureResponses(fixtureName);
    }

    serve(responses) {
        return this._native.serve(responses);
    }

    getPort() {
        return this._native.getPort();
    }

    shutdown() {
        return this._native.shutdown();
    }

    async assert(calls) {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                const result = this._native.assert(calls);
                if (result instanceof Error) {
                    reject(result);
                } else {
                    resolve(result);
                }
            }, 0);
        });
    }
}

module.exports = { Servermock };

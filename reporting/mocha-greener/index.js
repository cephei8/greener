const mocha = require("mocha");
const { Reporter } = require("greener-reporter");
const path = require("path");

function makeClassname(test) {
    let arr = [];
    let parent = test.parent;
    while (parent !== undefined) {
        if (parent.title !== "" && !parent.root) {
            arr.push(parent.title);
        }
        parent = parent.parent;
    }
    arr.reverse();
    return arr.join("::");
}

function GreenerReporter(runner, options = {}) {
    mocha.reporters.Base.call(this, runner, options);

    let ctx = {
        reporter: null,
        sessionPromise: null,
        session: null,
        rootDir: process.cwd(),
    };

    runner.on(mocha.Runner.constants.EVENT_RUN_BEGIN, function () {
        let ingress_endpoint = process.env.GREENER_INGRESS_ENDPOINT;
        let ingress_api_key = process.env.GREENER_INGRESS_API_KEY;
        ctx.reporter = new Reporter(ingress_endpoint, ingress_api_key);

        let session_id = process.env.GREENER_SESSION_ID;
        let description = process.env.GREENER_SESSION_DESCRIPTION;
        let baggage = process.env.GREENER_SESSION_BAGGAGE;
        let labels = process.env.GREENER_SESSION_LABELS;

        ctx.sessionPromise = ctx.reporter
            .createSession(
                session_id === undefined ? null : session_id,
                description === undefined ? null : description,
                baggage === undefined ? null : baggage,
                labels === undefined ? null : labels,
            )
            .then((session) => {
                ctx.session = session;
                return session;
            })
            .catch((error) => {
                const endpoint = ingress_endpoint || "undefined";
                console.error(error);
                throw new Error(
                    `Failed to create Greener session. ` +
                        `Endpoint: ${endpoint}. ` +
                        `Error: ${error.message || error}. ` +
                        `Please verify the Greener server is running and accessible.`,
                );
            });
    });

    runner.on(mocha.Runner.constants.EVENT_RUN_END, async function () {
        try {
            await ctx.sessionPromise;
            await ctx.reporter.shutdown();
        } catch (error) {
            const endpoint =
                process.env.GREENER_INGRESS_ENDPOINT || "undefined";
            console.error(error);
            throw new Error(
                `Failed to report test results to Greener. ` +
                    `Endpoint: ${endpoint}. ` +
                    `Error: ${error.message || error}. ` +
                    `Please verify the Greener server is running and accessible.`,
            );
        }
    });

    runner.on(mocha.Runner.constants.EVENT_TEST_PASS, async function (test) {
        try {
            await ctx.sessionPromise;
            ctx.reporter.createTestcase(
                ctx.session.id,
                test.title,
                makeClassname(test),
                path.relative(ctx.rootDir, test.file),
                null,
                "pass",
                null,
                null,
            );
        } catch (error) {
            const endpoint =
                process.env.GREENER_INGRESS_ENDPOINT || "undefined";
            console.error(error);
            throw new Error(
                `Failed to report test case "${test.title}" to Greener. ` +
                    `Endpoint: ${endpoint}. ` +
                    `Error: ${error.message || error}. ` +
                    `Please verify the Greener server is running and accessible.`,
            );
        }
    });

    runner.on(mocha.Runner.constants.EVENT_TEST_FAIL, async function (
        test,
        err,
    ) {
        try {
            await ctx.sessionPromise;
            ctx.reporter.createTestcase(
                ctx.session.id,
                test.title,
                makeClassname(test),
                path.relative(ctx.rootDir, test.file),
                null,
                "fail",
                null,
                null,
            );
        } catch (error) {
            const endpoint =
                process.env.GREENER_INGRESS_ENDPOINT || "undefined";
            console.error(error);
            throw new Error(
                `Failed to report test case "${test.title}" to Greener. ` +
                    `Endpoint: ${endpoint}. ` +
                    `Error: ${error.message || error}. ` +
                    `Please verify the Greener server is running and accessible.`,
            );
        }
    });

    runner.on(mocha.Runner.constants.EVENT_TEST_PENDING, async function (test) {
        try {
            await ctx.sessionPromise;
            ctx.reporter.createTestcase(
                ctx.session.id,
                test.title,
                makeClassname(test),
                path.relative(ctx.rootDir, test.file),
                null,
                "skip",
                null,
                null,
            );
        } catch (error) {
            const endpoint =
                process.env.GREENER_INGRESS_ENDPOINT || "undefined";
            console.error(error);
            throw new Error(
                `Failed to report test case "${test.title}" to Greener. ` +
                    `Endpoint: ${endpoint}. ` +
                    `Error: ${error.message || error}. ` +
                    `Please verify the Greener server is running and accessible.`,
            );
        }
    });
}

mocha.utils.inherits(GreenerReporter, mocha.reporters.Spec);

module.exports = GreenerReporter;

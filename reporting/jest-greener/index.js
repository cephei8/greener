const { Reporter } = require("greener-reporter");
const path = require("path");

class GreenerReporter {
    constructor(globalConfig, reporterOptions, reporterContext) {
        this._reporter = null;
        this._session = null;
        this._rootDir = globalConfig.rootDir;
    }

    async onRunStart(results, options) {
        let ingress_endpoint = process.env.GREENER_INGRESS_ENDPOINT;
        let ingress_api_key = process.env.GREENER_INGRESS_API_KEY;
        this._reporter = new Reporter(ingress_endpoint, ingress_api_key);

        let session_id = process.env.GREENER_SESSION_ID;
        let description = process.env.GREENER_SESSION_DESCRIPTION;
        let baggage = process.env.GREENER_SESSION_BAGGAGE;
        let labels = process.env.GREENER_SESSION_LABELS;

        try {
            this._session = await this._reporter.createSession(
                session_id === undefined ? null : session_id,
                description === undefined ? null : description,
                baggage === undefined ? null : baggage,
                labels === undefined ? null : labels,
            );
        } catch (error) {
            const endpoint = ingress_endpoint || "undefined";
            throw new Error(
                `Failed to create Greener session. ` +
                    `Endpoint: ${endpoint}. ` +
                    `Error: ${error.message || error}. ` +
                    `Please verify the Greener server is running and accessible.`,
            );
        }
    }

    async onRunComplete(testContexts, results) {
        try {
            for (const fileResults of results.testResults) {
                if (!fileResults.testExecError) {
                    for (const testCaseResult of fileResults.testResults) {
                        if (testCaseResult.status === "pending") {
                            this._reporter.createTestcase(
                                this._session.id,
                                testCaseResult.title,
                                testCaseResult.ancestorTitles.join("::"),
                                path.relative(
                                    this._rootDir,
                                    fileResults.testFilePath,
                                ),
                                null,
                                "skip",
                                null,
                                null,
                            );
                        }
                    }
                } else {
                    this._reporter.createTestcase(
                        this._session.id,
                        "*",
                        null,
                        path.relative(this._rootDir, fileResults.testFilePath),
                        null,
                        "error",
                        null,
                        null,
                    );
                }
            }

            await this._reporter.shutdown();
        } catch (error) {
            const endpoint =
                process.env.GREENER_INGRESS_ENDPOINT || "undefined";
            throw new Error(
                `Failed to report test results to Greener. ` +
                    `Endpoint: ${endpoint}. ` +
                    `Error: ${error.message || error}. ` +
                    `Please verify the Greener server is running and accessible.`,
            );
        }
    }

    onTestCaseResult(test, testCaseResult) {
        let status = { passed: "pass", failed: "fail" }[testCaseResult.status];

        try {
            this._reporter.createTestcase(
                this._session.id,
                testCaseResult.title,
                testCaseResult.ancestorTitles.join("::"),
                path.relative(this._rootDir, test.path),
                null,
                status,
                null,
                null,
            );
        } catch (error) {
            const endpoint =
                process.env.GREENER_INGRESS_ENDPOINT || "undefined";
            throw new Error(
                `Failed to report test case "${testCaseResult.title}" to Greener. ` +
                    `Endpoint: ${endpoint}. ` +
                    `Error: ${error.message || error}. ` +
                    `Please verify the Greener server is running and accessible.`,
            );
        }
    }
}

module.exports = GreenerReporter;

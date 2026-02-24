const http = require("http");
const https = require("https");

class Reporter {
    constructor(endpoint, apiKey) {
        if (typeof endpoint !== "string") {
            throw new Error("server_address must be a string");
        }
        if (typeof apiKey !== "string") {
            throw new Error("apiKey must be a string");
        }

        this.endpoint = endpoint;
        this.apiKey = apiKey;
        this.testcaseBatch = [];
        this.batchTimer = null;
        this.errors = [];
        this.isShutdown = false;
        this.inflightFlush = null;
    }

    async createSession(id, description, baggage, labels) {
        if (this.isShutdown) {
            throw new Error("Reporter has been shut down");
        }

        if (id !== null && typeof id !== "string") {
            throw new Error("session_id must be a nullable string");
        }
        if (description !== null && typeof description !== "string") {
            throw new Error("description must be a nullable string");
        }
        if (baggage !== null && typeof baggage !== "string") {
            throw new Error("baggage must be a nullable string");
        }
        if (labels !== null && typeof labels !== "string") {
            throw new Error("labels must be a nullable string");
        }

        const body = {};
        if (id !== null) {
            body.id = id;
        }
        if (description !== null) {
            body.description = description;
        }
        if (baggage !== null) {
            try {
                body.baggage = JSON.parse(baggage);
            } catch (err) {
                throw new Error(`invalid baggage JSON: ${err.message}`);
            }
        }
        if (labels !== null) {
            body.labels = labels;
        }

        const url = `${this.endpoint}/api/v1/ingress/sessions`;
        const response = await this._makeRequest("POST", url, body);

        return { id: response.id };
    }

    createTestcase(
        sessionId,
        testcaseName,
        testcaseClassname,
        testcaseFile,
        testsuite,
        status,
        output,
        baggage,
    ) {
        if (this.isShutdown) {
            throw new Error("Reporter has been shut down");
        }

        if (typeof sessionId !== "string") {
            throw new Error("session_id must be a string");
        }
        if (typeof testcaseName !== "string") {
            throw new Error("testcase_name must be a string");
        }
        if (typeof status !== "string") {
            throw new Error("status must be a string");
        }

        const validStatuses = ["pass", "fail", "error", "skip"];
        if (!validStatuses.includes(status)) {
            throw new Error("status must be one of pass|fail|err|skip");
        }

        if (
            testcaseClassname !== null &&
            typeof testcaseClassname !== "string"
        ) {
            throw new Error("testcase_classname must be a nullable string");
        }
        if (testcaseFile !== null && typeof testcaseFile !== "string") {
            throw new Error("testcase_file must be a nullable string");
        }
        if (testsuite !== null && typeof testsuite !== "string") {
            throw new Error("testsuite must be a nullable string");
        }
        if (output !== null && typeof output !== "string") {
            throw new Error("stdout_stream must be a nullable string");
        }
        if (baggage !== null && typeof baggage !== "string") {
            throw new Error("baggage must be a nullable string");
        }

        const testcase = {
            sessionId: sessionId,
            testcaseName: testcaseName,
            status: status,
        };

        if (testcaseClassname !== null) {
            testcase.testcaseClassname = testcaseClassname;
        }
        if (testcaseFile !== null) {
            testcase.testcaseFile = testcaseFile;
        }
        if (testsuite !== null) {
            testcase.testsuite = testsuite;
        }
        if (output !== null) {
            testcase.output = output;
        }
        if (baggage !== null) {
            try {
                testcase.baggage = JSON.parse(baggage);
            } catch (err) {
                throw new Error(`invalid baggage JSON: ${err.message}`);
            }
        }

        this.testcaseBatch.push(testcase);

        if (this.testcaseBatch.length >= 100) {
            this._flushTestcases();
        } else if (this.batchTimer === null) {
            this._startBatchTimer();
        }

        return null;
    }

    popError() {
        return this.errors.shift() || null;
    }

    async shutdown() {
        if (this.isShutdown) {
            return;
        }

        if (this.batchTimer !== null) {
            clearTimeout(this.batchTimer);
            this.batchTimer = null;
        }

        if (this.inflightFlush !== null) {
            await this.inflightFlush;
        }

        await this._flushTestcases();

        await new Promise((resolve) => setImmediate(resolve));

        this.isShutdown = true;
    }

    _startBatchTimer() {
        this.batchTimer = setTimeout(() => {
            this._flushTestcases();
        }, 5000);
    }

    async _flushTestcases() {
        if (this.batchTimer !== null) {
            clearTimeout(this.batchTimer);
            this.batchTimer = null;
        }

        if (this.testcaseBatch.length === 0) {
            return;
        }

        const batch = this.testcaseBatch;
        this.testcaseBatch = [];

        const url = `${this.endpoint}/api/v1/ingress/testcases`;
        const body = { testcases: batch };

        const flushPromise = (async () => {
            try {
                await this._makeRequest("POST", url, body);
            } catch (err) {
                this.errors.push(err);
            } finally {
                if (this.inflightFlush === flushPromise) {
                    this.inflightFlush = null;
                }
            }
        })();

        this.inflightFlush = flushPromise;
        return flushPromise;
    }

    _makeRequest(method, url, body) {
        const urlObj = new URL(url);
        const isHttps = urlObj.protocol === "https:";
        const httpModule = isHttps ? https : http;

        const options = {
            method: method,
            hostname: urlObj.hostname,
            port: urlObj.port || (isHttps ? 443 : 80),
            path: urlObj.pathname + urlObj.search,
            headers: {
                "Content-Type": "application/json",
                "X-API-Key": this.apiKey,
            },
        };

        const bodyJson = JSON.stringify(body);

        return new Promise((resolve, reject) => {
            const req = httpModule.request(options, (res) => {
                let data = "";

                res.on("data", (chunk) => {
                    data += chunk;
                });

                res.on("end", () => {
                    if (res.statusCode == 200 || res.statusCode == 201) {
                        try {
                            const response = JSON.parse(data);
                            resolve(response);
                        } catch (err) {
                            reject(
                                new Error(
                                    `failed to parse response: ${err.message}`,
                                ),
                            );
                        }
                    } else {
                        let errorMessage;
                        let errorCode = 1;
                        try {
                            const errorResp = JSON.parse(data);
                            errorMessage =
                                errorResp.detail || errorResp.message;
                            if (errorResp.code !== undefined) {
                                errorCode = errorResp.code;
                            }
                        } catch (err) {
                            errorMessage = data;
                        }

                        const ingressCode = res.statusCode;
                        reject(
                            new Error(
                                `GreenerReporterError ${errorCode}/${ingressCode}: ${errorMessage}`,
                            ),
                        );
                    }
                });
            });

            req.on("error", (err) => {
                const errorCode = 1;
                const ingressCode = 0;
                reject(
                    new Error(
                        `GreenerReporterError ${errorCode}/${ingressCode}: ${err.message}`,
                    ),
                );
            });

            req.write(bodyJson);
            req.end();
        });
    }
}

module.exports = { Reporter };

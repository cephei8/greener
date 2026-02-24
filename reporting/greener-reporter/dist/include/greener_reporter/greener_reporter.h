#ifndef GREENER_REPORTER_GREENER_REPORTER_H
#define GREENER_REPORTER_GREENER_REPORTER_H

#ifdef __cplusplus
extern "C" {
#endif

struct greener_reporter;
struct greener_reporter_session;
struct greener_reporter_error;

struct greener_reporter *
greener_reporter_new(const char *endpoint, const char *api_key,
                     const struct greener_reporter_error **error);

void greener_reporter_delete(struct greener_reporter *reporter,
                             const struct greener_reporter_error **error);

void greener_reporter_report_error_pop(
    struct greener_reporter *reporter,
    const struct greener_reporter_error **error);

const struct greener_reporter_session *
greener_reporter_session_create(struct greener_reporter *reporter,
                                const char *session_id,
                                const char *description, const char *baggage,
                                const char *labels,
                                const struct greener_reporter_error **error);

void greener_reporter_testcase_create(
    struct greener_reporter *reporter, const char *session_id,
    const char *testcase_name, const char *testcase_classname,
    const char *testcase_file, const char *testsuite, const char *status,
    const char *output, const char *baggage,
    const struct greener_reporter_error **error);

void greener_reporter_session_delete(
    const struct greener_reporter_session *session);

void greener_reporter_error_delete(const struct greener_reporter_error *error);

struct greener_reporter_session {
    const char *id;
};

struct greener_reporter_error {
    int code;
    int ingress_code;
    const char *message;
};

enum {
    GREENER_REPORTER_ERROR = 1,
    GREENER_REPORTER_ERROR_INVALID_ARGUMENT = 2,
    GREENER_REPORTER_ERROR_INGRESS = 3,
};

#ifdef __cplusplus
}
#endif

#endif // GREENER_REPORTER_GREENER_REPORTER_H

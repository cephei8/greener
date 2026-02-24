#ifndef GREENER_SERVERMOCK_GREENER_SERVERMOCK_H
#define GREENER_SERVERMOCK_GREENER_SERVERMOCK_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

struct greener_servermock;
struct greener_servermock_error;


struct greener_servermock*
greener_servermock_new();

void
greener_servermock_delete(
    struct greener_servermock* ctx,
    const struct greener_servermock_error** error);

void
greener_servermock_serve(
    struct greener_servermock* ctx,
    const char* responses,
    const struct greener_servermock_error** error);

int
greener_servermock_get_port(
    struct greener_servermock* ctx,
    const struct greener_servermock_error** error);

void
greener_servermock_assert(
    struct greener_servermock* ctx,
    const char* calls,
    const struct greener_servermock_error** error);

void
greener_servermock_fixture_names(
    struct greener_servermock* ctx,
    const char*** names,
    uint32_t* num_names,
    const struct greener_servermock_error** error);

void
greener_servermock_fixture_calls(
    struct greener_servermock* ctx,
    const char* fixture_name,
    const char** calls,
    const struct greener_servermock_error** error);

void
greener_servermock_fixture_responses(
    struct greener_servermock* ctx,
    const char* fixture_name,
    const char** responses,
    const struct greener_servermock_error** error);

void
greener_servermock_error_delete(
    const struct greener_servermock_error* error);

struct greener_servermock_error {
    const char *message;
};

#ifdef __cplusplus
}
#endif

#endif // GREENER_SERVERMOCK_GREENER_SERVERMOCK_H

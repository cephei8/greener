#include "servermock.h"

#include <greener_servermock/greener_servermock.h>

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void
handle_servermock_error(napi_env env,
                        const struct greener_servermock_error *err) {
  char msg[512];
  snprintf(msg, sizeof(msg), "GreenerServermockError: %s", err->message);

  napi_value msg_value;
  napi_create_string_utf8(env, msg, NAPI_AUTO_LENGTH, &msg_value);

  napi_value error;
  napi_create_error(env, NULL, msg_value, &error);

  napi_throw(env, error);
}

static void handle_napi_error(napi_env env) {
  bool pending;
  napi_is_exception_pending(env, &pending);
  if (!pending) {
    const napi_extended_error_info *error_info = NULL;
    napi_get_last_error_info(env, &error_info);
    const char *err_message = error_info->error_message;
    if (err_message == NULL) {
      err_message = "internal error";
    }
    napi_throw_error(env, NULL, err_message);
  }
}

static void servermock_destructor(struct servermock *obj) {
  if (obj->servermock) {
    const struct greener_servermock_error *err;
    greener_servermock_delete(obj->servermock, &err);
  }
  napi_delete_reference(obj->env, obj->wrapper);
  free(obj);
}

void servermock_finalizer(napi_env env, void *obj, void *finalize_hint) {
  (void)finalize_hint;
  servermock_destructor((struct servermock *)obj);
}

static napi_value servermock_ctor(napi_env env, napi_callback_info info);
static napi_value servermock_serve(napi_env env, napi_callback_info info);
static napi_value servermock_get_port(napi_env env, napi_callback_info info);
static napi_value servermock_assert_calls(napi_env env,
                                          napi_callback_info info);
static napi_value servermock_fixture_names(napi_env env,
                                           napi_callback_info info);
static napi_value servermock_fixture_calls(napi_env env,
                                           napi_callback_info info);
static napi_value servermock_fixture_responses(napi_env env,
                                               napi_callback_info info);
static napi_value servermock_shutdown(napi_env env, napi_callback_info info);

static void instance_data_finalizer(napi_env env, void *data, void *hint) {
  (void)hint;
  napi_ref *ctor_ref = (napi_ref *)data;
  napi_status status = napi_delete_reference(env, *ctor_ref);
  free(ctor_ref);

  if (status != napi_ok) {
    handle_napi_error(env);
  }
}

napi_value servermock_register_class(napi_env env, napi_value exports) {
  napi_property_descriptor properties[] = {
      {"serve", NULL, servermock_serve, NULL, NULL, NULL, napi_default, NULL},
      {"getPort", NULL, servermock_get_port, NULL, NULL, NULL, napi_default,
       NULL},
      {"assert", NULL, servermock_assert_calls, NULL, NULL, NULL, napi_default,
       NULL},
      {"fixtureNames", NULL, servermock_fixture_names, NULL, NULL, NULL,
       napi_default, NULL},
      {"fixtureCalls", NULL, servermock_fixture_calls, NULL, NULL, NULL,
       napi_default, NULL},
      {"fixtureResponses", NULL, servermock_fixture_responses, NULL, NULL, NULL,
       napi_default, NULL},
      {"shutdown", NULL, servermock_shutdown, NULL, NULL, NULL, napi_default,
       NULL},
  };

  napi_value ctor_value;
  napi_status status = napi_define_class(
      env, "Servermock", NAPI_AUTO_LENGTH, servermock_ctor, NULL,
      sizeof(properties) / sizeof(properties[0]), properties, &ctor_value);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  napi_ref *ctor_ref = (napi_ref *)malloc(sizeof(napi_ref));
  if (!ctor_ref) {
    napi_throw_error(env, NULL, "out of memory");
    return NULL;
  }

  status = napi_create_reference(env, ctor_value, 1, ctor_ref);
  if (status != napi_ok) {
    free(ctor_ref);
    handle_napi_error(env);
    return NULL;
  }

  status = napi_set_instance_data(env, ctor_ref, instance_data_finalizer, NULL);
  if (status != napi_ok) {
    napi_delete_reference(env, *ctor_ref);
    free(ctor_ref);
    handle_napi_error(env);
    return NULL;
  }

  status = napi_set_named_property(env, exports, "Servermock", ctor_value);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  return exports;
}

static napi_value servermock_ctor(napi_env env, napi_callback_info info) {
  napi_value target;
  napi_status status = napi_get_new_target(env, info, &target);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  size_t argc = 0;
  napi_value jsthis;
  status = napi_get_cb_info(env, info, &argc, NULL, &jsthis, NULL);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (argc != 0) {
    napi_throw_error(env, NULL, "no arguments expected");
    return NULL;
  }

  struct servermock *obj =
      (struct servermock *)malloc(sizeof(struct servermock));
  if (!obj) {
    napi_throw_error(env, NULL, "out of memory");
    return NULL;
  }

  obj->env = env;
  obj->servermock = NULL;
  obj->wrapper = NULL;

  status =
      napi_wrap(env, jsthis, obj, servermock_finalizer, NULL, &obj->wrapper);
  if (status != napi_ok) {
    free(obj);
    handle_napi_error(env);
    return NULL;
  }

  obj->servermock = greener_servermock_new();

  return jsthis;
}

static napi_value servermock_serve(napi_env env, napi_callback_info info) {
  const size_t argc_num = 1;
  size_t argc = argc_num;
  napi_value args[1];
  napi_value jsthis;
  napi_status status = napi_get_cb_info(env, info, &argc, args, &jsthis, NULL);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (argc != 1) {
    napi_throw_error(env, NULL, "invalid number of arguments");
    return NULL;
  }

  char buf[1024 * 1024];
  size_t buf_len;

  napi_valuetype value_type;
  status = napi_typeof(env, args[0], &value_type);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (value_type != napi_string) {
    napi_throw_error(env, NULL, "responses must be a string");
    return NULL;
  }

  status = napi_get_value_string_utf8(env, args[0], buf, sizeof(buf), &buf_len);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  struct servermock *obj;
  status = napi_unwrap(env, jsthis, (void **)&obj);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  const struct greener_servermock_error *err;
  greener_servermock_serve(obj->servermock, buf, &err);

  if (err != NULL) {
    handle_servermock_error(env, err);
    return NULL;
  }

  return NULL;
}

static napi_value servermock_assert_calls(napi_env env,
                                          napi_callback_info info) {
  const size_t argc_num = 1;
  size_t argc = argc_num;
  napi_value args[1];
  napi_value jsthis;
  napi_status status = napi_get_cb_info(env, info, &argc, args, &jsthis, NULL);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (argc != 1) {
    napi_throw_error(env, NULL, "invalid number of arguments");
    return NULL;
  }

  char buf[1024 * 1024];
  size_t buf_len;

  napi_valuetype value_type;
  status = napi_typeof(env, args[0], &value_type);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (value_type != napi_string) {
    napi_throw_error(env, NULL, "calls must be a string");
    return NULL;
  }

  status = napi_get_value_string_utf8(env, args[0], buf, sizeof(buf), &buf_len);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  struct servermock *obj;
  status = napi_unwrap(env, jsthis, (void **)&obj);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  const struct greener_servermock_error *err;
  greener_servermock_assert(obj->servermock, buf, &err);

  if (err != NULL) {
    char msg[512];
    snprintf(msg, sizeof(msg), "GreenerServermockError: %s", err->message);
    greener_servermock_error_delete(err);

    napi_value msg_value;
    napi_create_string_utf8(env, msg, NAPI_AUTO_LENGTH, &msg_value);

    napi_value error;
    napi_create_error(env, NULL, msg_value, &error);

    return error;
  }

  napi_value undefined;
  napi_get_undefined(env, &undefined);
  return undefined;
}

static napi_value servermock_fixture_names(napi_env env,
                                           napi_callback_info info) {
  size_t argc = 0;
  napi_value jsthis;
  napi_status status = napi_get_cb_info(env, info, &argc, NULL, &jsthis, NULL);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (argc != 0) {
    napi_throw_error(env, NULL, "no arguments expected");
    return NULL;
  }

  struct servermock *obj;
  status = napi_unwrap(env, jsthis, (void **)&obj);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  const struct greener_servermock_error *err;
  const char **name_ptrs;
  uint32_t num_names;
  greener_servermock_fixture_names(obj->servermock, &name_ptrs, &num_names,
                                   &err);

  if (err != NULL) {
    handle_servermock_error(env, err);
    return NULL;
  }

  napi_value out_names_value;
  status = napi_create_array_with_length(env, num_names, &out_names_value);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  for (uint32_t i = 0; i < num_names; ++i) {
    napi_value name_value;
    status = napi_create_string_utf8(env, name_ptrs[i], strlen(name_ptrs[i]),
                                     &name_value);
    if (status != napi_ok) {
      handle_napi_error(env);
      return NULL;
    }

    status = napi_set_element(env, out_names_value, i, name_value);
    if (status != napi_ok) {
      handle_napi_error(env);
      return NULL;
    }
  }

  return out_names_value;
}

static napi_value servermock_fixture_calls(napi_env env,
                                           napi_callback_info info) {
  const size_t argc_num = 1;
  size_t argc = argc_num;
  napi_value args[1];
  napi_value jsthis;
  napi_status status = napi_get_cb_info(env, info, &argc, args, &jsthis, NULL);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (argc != 1) {
    napi_throw_error(env, NULL, "invalid number of arguments");
    return NULL;
  }

  char buf[1024 * 1024];
  size_t buf_len;

  napi_valuetype value_type;
  status = napi_typeof(env, args[0], &value_type);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (value_type != napi_string) {
    napi_throw_error(env, NULL, "fixture_name must be a string");
    return NULL;
  }

  status = napi_get_value_string_utf8(env, args[0], buf, sizeof(buf), &buf_len);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  struct servermock *obj;
  status = napi_unwrap(env, jsthis, (void **)&obj);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  const struct greener_servermock_error *err;
  const char *calls;
  greener_servermock_fixture_calls(obj->servermock, buf, &calls, &err);

  if (err != NULL) {
    handle_servermock_error(env, err);
    return NULL;
  }

  napi_value calls_value;
  status = napi_create_string_utf8(env, calls, strlen(calls), &calls_value);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  return calls_value;
}

static napi_value servermock_fixture_responses(napi_env env,
                                               napi_callback_info info) {
  const size_t argc_num = 1;
  size_t argc = argc_num;
  napi_value args[1];
  napi_value jsthis;
  napi_status status = napi_get_cb_info(env, info, &argc, args, &jsthis, NULL);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (argc != 1) {
    napi_throw_error(env, NULL, "invalid number of arguments");
    return NULL;
  }

  char buf[1024 * 1024];
  size_t buf_len;

  napi_valuetype value_type;
  status = napi_typeof(env, args[0], &value_type);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (value_type != napi_string) {
    napi_throw_error(env, NULL, "fixture_name must be a string");
    return NULL;
  }

  status = napi_get_value_string_utf8(env, args[0], buf, sizeof(buf), &buf_len);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  struct servermock *obj;
  status = napi_unwrap(env, jsthis, (void **)&obj);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  const struct greener_servermock_error *err;
  const char *responses;
  greener_servermock_fixture_responses(obj->servermock, buf, &responses, &err);

  if (err != NULL) {
    handle_servermock_error(env, err);
    return NULL;
  }

  napi_value responses_value;
  status = napi_create_string_utf8(env, responses, strlen(responses),
                                   &responses_value);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  return responses_value;
}

static napi_value servermock_shutdown(napi_env env, napi_callback_info info) {
  size_t argc = 0;
  napi_value jsthis;
  napi_status status = napi_get_cb_info(env, info, &argc, NULL, &jsthis, NULL);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (argc != 0) {
    napi_throw_error(env, NULL, "no arguments expected");
    return NULL;
  }

  struct servermock *obj;
  status = napi_unwrap(env, jsthis, (void **)&obj);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  const struct greener_servermock_error *err;
  greener_servermock_delete(obj->servermock, &err);
  obj->servermock = NULL;

  if (err != NULL) {
    handle_servermock_error(env, err);
    return NULL;
  }

  return NULL;
}

static napi_value servermock_get_port(napi_env env, napi_callback_info info) {
  size_t argc = 0;
  napi_value jsthis;
  napi_status status = napi_get_cb_info(env, info, &argc, NULL, &jsthis, NULL);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  if (argc != 0) {
    napi_throw_error(env, NULL, "no arguments expected");
    return NULL;
  }

  struct servermock *obj;
  status = napi_unwrap(env, jsthis, (void **)&obj);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  const struct greener_servermock_error *err;
  int port = greener_servermock_get_port(obj->servermock, &err);

  if (err != NULL) {
    handle_servermock_error(env, err);
    return NULL;
  }

  napi_value port_value;
  status = napi_create_int32(env, port, &port_value);
  if (status != napi_ok) {
    handle_napi_error(env);
    return NULL;
  }

  return port_value;
}

#ifndef GREENER_SERVERMOCK_ADDON_SERVERMOCK_H
#define GREENER_SERVERMOCK_ADDON_SERVERMOCK_H

#include <node_api.h>

struct greener_servermock;

struct servermock {
  struct greener_servermock *servermock;
  napi_env env;
  napi_ref wrapper;
};

napi_value servermock_register_class(napi_env env, napi_value exports);
void servermock_finalizer(napi_env env, void *obj, void *finalize_hint);

#endif /* GREENER_SERVERMOCK_ADDON_SERVERMOCK_H */

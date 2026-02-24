#include "servermock.h"

napi_value register_addon(napi_env env, napi_value exports) {
  return servermock_register_class(env, exports);
}

NAPI_MODULE(NODE_GYP_MODULE_NAME, register_addon)

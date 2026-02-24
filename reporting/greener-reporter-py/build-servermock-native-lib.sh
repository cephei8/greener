#!/usr/bin/env bash
set -e

OS_TYPE=$(uname -s)

cd ../greener-reporter
cargo build --release --package greener-servermock

if [[ "$OS_TYPE" == "Darwin" ]]; then
  install_name_tool -id @rpath/libgreener_servermock.dylib target/release/libgreener_servermock.dylib
elif [[ "$OS_TYPE" == "MINGW"* || "$OS_TYPE" == "MSYS"* || "$OS_TYPE" == "CYGWIN"* ]]; then
  cp target/release/greener_servermock.dll.lib target/release/greener_servermock.lib
fi

cargo test --release --package greener-servermock

cd ../..

if [[ "$OS_TYPE" == "Darwin" ]]; then
  cp ../greener-reporter/target/release/libgreener_servermock.dylib greener_servermock/
elif [[ "$OS_TYPE" == "Linux" ]]; then
  cp ../greener-reporter/target/release/libgreener_servermock.so greener_servermock/
elif [[ "$OS_TYPE" == "MINGW"* || "$OS_TYPE" == "MSYS"* || "$OS_TYPE" == "CYGWIN"* ]]; then
  cp ../greener-reporter/target/release/greener_servermock.dll greener_servermock/
  cp ../greener-reporter/target/release/greener_servermock.lib greener_servermock/
fi

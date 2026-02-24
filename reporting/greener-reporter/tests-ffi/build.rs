use std::env;
use std::path::PathBuf;

fn main() {
    let manifest_dir = env::var("CARGO_MANIFEST_DIR").unwrap();
    let profile = env::var("PROFILE").unwrap();
    let target = std::env::var("TARGET").unwrap();

    let target_dir_with_target = PathBuf::from(&manifest_dir)
        .join("..")
        .join("target")
        .join(&target)
        .join(&profile);

    let target_dir = if target_dir_with_target.exists() {
        target_dir_with_target
    } else {
        PathBuf::from(&manifest_dir)
            .join("..")
            .join("target")
            .join(&profile)
    };

    println!("cargo:rustc-link-search=native={}", target_dir.display());
    println!("cargo:rustc-link-lib=dylib=greener_reporter");
    println!("cargo:rustc-link-lib=dylib=greener_servermock");

    let include_dir = PathBuf::from(&manifest_dir).join("../dist/include");
    let reporter_header = include_dir.join("greener_reporter/greener_reporter.h");
    let servermock_header = include_dir.join("greener_servermock/greener_servermock.h");

    let bindings = bindgen::Builder::default()
        .header(reporter_header.to_string_lossy())
        .header(servermock_header.to_string_lossy())
        .clang_arg(format!("-I{}", include_dir.display()))
        .parse_callbacks(Box::new(bindgen::CargoCallbacks::new()))
        .generate()
        .expect("Unable to generate bindings");

    let out_path = PathBuf::from(env::var("OUT_DIR").unwrap());
    bindings
        .write_to_file(out_path.join("bindings.rs"))
        .expect("Couldn't write bindings!");
}

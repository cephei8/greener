{
    "target_defaults": {
        "include_dirs": ["<(module_root_dir)/dist/include"],
        "conditions": [
            [
                "OS=='mac'",
                {
                    "library_dirs": [
                        "<(module_root_dir)/prebuilds/darwin-<(target_arch)"
                    ],
                    "link_settings": {
                        "libraries": ["-Wl,-rpath,@loader_path"]
                    }
                }
            ],
            [
                "OS=='linux'",
                {
                    "library_dirs": [
                        "<(module_root_dir)/prebuilds/linux-<(target_arch)"
                    ],
                    "ldflags": ["-Wl,-rpath,'$$ORIGIN'"]
                }
            ],
            [
                "OS=='win'",
                {
                    "library_dirs": [
                        "<(module_root_dir)/prebuilds/win32-<(target_arch)"
                    ]
                }
            ]
        ]
    },
    "targets": [
        {
            "target_name": "greener_servermock_addon",
            "sources": ["addon/addon.c", "addon/servermock.c"],
            "libraries": ["-lgreener_servermock"]
        }
    ]
}

{
  description = "Greener";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_25

            gopls
            gotools
            go-tools
            golangci-lint
            go-mockery

            gcc
            pkg-config

            sqlite
            postgresql
            mysql80
          ];

          CGO_ENABLED = "1";
        };
      }
    );
}

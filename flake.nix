{
  description = "Terraform Provider B2";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
  };
  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      perSystem =
        {
          config,
          self',
          inputs',
          pkgs,
          system,
          ...
        }:
        {
          _module.args.pkgs = import inputs.nixpkgs {
            inherit system;
            # beause terraform
            config.allowUnfree = true;
          };
          devShells.default = pkgs.mkShell {
            nativeBuildInputs =
              with pkgs;
              [
                go
                python312Packages.pyinstaller
                python312Packages.black
                python312Packages.flake8
                gnumake
                terraform
              ]
              ++ lib.optional stdenv.isLinux [
                patchelf
                # as soon as staticx become uv, just add it
              ];
          };
        };
    };
}

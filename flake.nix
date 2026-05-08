{
  description = "devshell";
  nixConfig.bash-prompt = "[nix] ";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  };

  outputs =
    { nixpkgs, ... }:
    let
      x86_64-linux = "x86_64-linux";
      aarch64-linux = "aarch64-linux";
      aarch64-darwin = "aarch64-darwin";

      defaultSystems = [
        x86_64-linux
        aarch64-linux
        aarch64-darwin
      ];

      eachSystem =
        systems: fun:
        nixpkgs.lib.genAttrs systems (
          system:
          let
            pkgs = import nixpkgs { inherit system; };
          in
          fun {
            inherit pkgs system;
          }
        );
    in
    {
      devShell = eachSystem defaultSystems (
        { pkgs, ... }:
        pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_26
            gopls
            gotools
            gofumpt
          ];

          shellHook = ''
            for p in $NIX_PROFILES; do
              GOPATH="$p/share/go:$GOPATH"
            done
          '';
        }
      );
    };
}

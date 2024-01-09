{
  description = "Nix flake for tox4go";
  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      packages = flake-utils.lib.flattenTree rec {
        tox4go-state-tool = with pkgs; buildGoModule rec {
          name = "tox4go-state-tool";
          src = ./.;

          subPackages = [ "cmd/state-tool" ];
          vendorHash = "sha256-QXDNPADQqKx7HK3b2IwkalmCdwMX3TaPyEVAzVSwuyo=";

          postInstall = ''
            mv $out/bin/state-tool $out/bin/${name}
          '';
        };
      };
      devShell = with pkgs; mkShell {
        buildInputs = [
          go
        ];
      };
    }
  );
}

{
  description = "Ephemeral IAM";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-22.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
     in {
        packages.default = pkgs.buildGoModule rec {
          pname = "ephemeral-iam";
          version = "0.0.22";

          src = ./.;

          vendorSha256 = "sha256-iJe97gPFTVmiFbHNEqhrA+xqFyXO8kz7K69wm8IJ+AE=";

          buildInputs = [ pkgs.makeWrapper ];
          postInstall = ''
            wrapProgram "$out/bin/ephemeral-iam" \
              --prefix PATH : $out/bin:${pkgs.lib.makeBinPath [ pkgs.google-cloud-sdk ]}
            ln -s $out/bin/ephemeral-iam $out/bin/eiam
          '';

          doCheck = false;
        };

        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [ go gopls ];
        };
      });
}

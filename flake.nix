{
  description = "Fully customizable local REST API for testing";

  inputs = { nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable"; };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};

      version = "0.4.0";
      commitSha = self.shortRev or "dirty";

      src = pkgs.lib.cleanSourceWith {
        src = ./.;
        filter = path: type:
          let baseName = baseNameOf (toString path);
          in (type == "directory") || (pkgs.lib.hasSuffix ".go" baseName)
          || (baseName == "go.mod") || (baseName == "go.sum");
      };

      fakeapi = pkgs.buildGoModule {
        pname = "fakeapi";
        inherit version;
        src = src;

        vendorHash = "sha256-omU6hZWoP9NUFw4yVFBWpjt7Zx9h8/85SJYAHp5BpZM=";

        ldflags = [
          "-X fakeapi/internal/version.Version=${version}"
          "-X fakeapi/internal/version.Commit=${commitSha}"
        ];

        doCheck = false;
      };

    in { packages.${system}.default = fakeapi; };
}


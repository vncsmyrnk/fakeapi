{
  description = "Customizable local REST API for testing";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      serverVersion = "0.4.0";
      cliVersion = "0.1.0";

      src = pkgs.lib.cleanSourceWith {
        src = ./.;
        filter =
          path: type:
          let
            baseName = baseNameOf (toString path);
          in
          (type == "directory")
          || (pkgs.lib.hasSuffix ".go" baseName)
          || (pkgs.lib.hasSuffix ".sql" baseName)
          || (baseName == "go.mod")
          || (baseName == "go.sum");
      };

      server = pkgs.buildGoModule {
        name = "fakeapi";
        src = src;
        version = serverVersion;
        vendorHash = "sha256-AfsFzCWJYBtji9yTHDmmOjxo2vdPGnAasRzwbXLlgCc=";
        doCheck = false;
        subPackages = [
          "cmd/server"
        ];

        nativeBuildInputs = [ pkgs.pkg-config ];
        buildInputs = [ pkgs.sqlite ];

        env = {
          CGO_ENABLED = 1;
        };

        postInstall = ''
          mv $out/bin/server $out/bin/fakeapi
        '';
      };

      cli = pkgs.buildGoModule {
        name = "fakeassert";
        src = src;
        version = cliVersion;
        vendorHash = "sha256-AfsFzCWJYBtji9yTHDmmOjxo2vdPGnAasRzwbXLlgCc=";
        doCheck = false;
        subPackages = [
          "cmd/cli"
        ];

        postInstall = ''
          mv $out/bin/cli $out/bin/fakeassert
        '';

      };

      serverDockerImage = pkgs.dockerTools.buildImage {
        name = "fakeapi";
        tag = if serverVersion != "" then serverVersion else "latest";

        copyToRoot = pkgs.buildEnv {
          name = "image-root";
          paths = with pkgs; [
            server
            coreutils
          ];
          pathsToLink = [
            "/bin"
          ];
        };

        runAsRoot = ''
          #!${pkgs.runtimeShell}
          mkdir -p /data
          chown -R 1000:1000 /data
        '';

        config = {
          Env = [
            "PATH=/bin"
            "FAKEAPI_DB_PATH=/data/db"
          ];

          User = "1000:1000";
          Entrypoint = [ "${server}/bin/fakeapi" ];
          Labels = {
            "org.opencontainers.image.source" = "https://github.com/vncsmyrnk/fakeapi";
          };
        };
      };
    in
    {
      packages.${system} = {
        default = cli;
        cli = cli;
        docker = serverDockerImage;
        server = server;

        all = pkgs.symlinkJoin {
          name = "all-packages";
          paths = [
            cli
            server
          ];
        };
      };
    };
}

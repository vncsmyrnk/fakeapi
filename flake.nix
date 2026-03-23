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

      serverVersion = "0.6.0";
      server = pkgs.buildGoModule {
        name = "fakeapi-server";
        src = src;
        version = serverVersion;
        vendorHash = "sha256-KaOkEQcsbKCO7clVWXO6wgftYzQieujtzuC8xD0ftd8=";
        doCheck = false;

        subPackages = [
          "cmd/server"
        ];

        nativeBuildInputs = [ pkgs.pkg-config ];
        buildInputs = [ pkgs.sqlite ];

        env = {
          CGO_ENABLED = 1;
        };

        ldflags = [
          "-s"
          "-w"
          "-X main.ServerVersion=${serverVersion}"
        ];

        postInstall = ''
          mv $out/bin/server $out/bin/fakeapi-server
        '';
      };

      cliVersion = "0.7.0";
      cli = pkgs.buildGoModule {
        name = "fakeapi-cli";
        src = src;
        version = cliVersion;
        vendorHash = "sha256-KaOkEQcsbKCO7clVWXO6wgftYzQieujtzuC8xD0ftd8=";
        doCheck = false;

        subPackages = [
          "cmd/cli"
        ];

        ldflags = [
          "-s"
          "-w"
          "-X main.CliVersion=${cliVersion}"
        ];

        postInstall = ''
          mv $out/bin/cli $out/bin/fakeapi
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
          WorkingDir = "/data";
          Entrypoint = [ "${server}/bin/fakeapi-server" ];
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
        fakeapi-server = server;

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

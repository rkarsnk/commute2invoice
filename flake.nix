{
  description = "commute2invoice - podman で起動する開発/実行環境";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        appName = "commute2invoice";
        image = "${appName}:latest";
        volume = "${appName}-data";

        # docker-compose.yml をそのまま podman-compose で利用する
        up = pkgs.writeShellApplication {
          name = "podman-up";
          runtimeInputs = [ pkgs.podman-compose pkgs.podman ];
          text = ''
            podman-compose up --build --detach
          '';
        };

        down = pkgs.writeShellApplication {
          name = "podman-down";
          runtimeInputs = [ pkgs.podman-compose pkgs.podman ];
          text = ''
            podman-compose down
          '';
        };

        logs = pkgs.writeShellApplication {
          name = "podman-logs";
          runtimeInputs = [ pkgs.podman-compose pkgs.podman ];
          text = ''
            podman-compose logs --follow app
          '';
        };

        # docker-compose を使わず、docs/macos-container.md 相当の手順を
        # podman ネイティブコマンドで行いたい場合の代替
        runNative = pkgs.writeShellApplication {
          name = "podman-run-native";
          runtimeInputs = [ pkgs.podman ];
          text = ''
            podman build --tag ${image} .
            podman volume create ${volume} >/dev/null 2>&1 || true
            podman run --detach \
              --name ${appName} \
              --publish 8080:8080 \
              --mount type=volume,source=${volume},target=/data \
              ${image}
          '';
        };

        stopNative = pkgs.writeShellApplication {
          name = "podman-stop-native";
          runtimeInputs = [ pkgs.podman ];
          text = ''
            podman stop ${appName}
            podman rm ${appName}
          '';
        };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.podman
            pkgs.podman-compose
            pkgs.go
            pkgs.sqlite
          ];

          shellHook = ''
            echo "commute2invoice 開発シェル (podman)"
            echo "  nix run .#up            -- podman-compose で起動"
            echo "  nix run .#down          -- podman-compose で停止"
            echo "  nix run .#logs          -- ログ追尾"
            echo "  nix run .#run-native    -- podman ネイティブコマンドで起動"
            echo "  nix run .#stop-native   -- podman ネイティブコマンドで停止"
          '';
        };

        apps = {
          up = flake-utils.lib.mkApp { drv = up; };
          down = flake-utils.lib.mkApp { drv = down; };
          logs = flake-utils.lib.mkApp { drv = logs; };
          run-native = flake-utils.lib.mkApp { drv = runNative; };
          stop-native = flake-utils.lib.mkApp { drv = stopNative; };
          default = flake-utils.lib.mkApp { drv = up; };
        };
      });
}

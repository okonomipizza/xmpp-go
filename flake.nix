{
  description = "xmpp-go - Sans-I/O XMPP protocol implementation in Go";

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
            # Go
            go
            gopls
            gotools
            go-tools  # staticcheck
            delve     # debugger

            # Testing
            # rapid is installed via go get

            # Documentation
            graphviz  # for state machine diagrams

            # Utilities
            jq
            curl
          ];

          shellHook = ''
            echo "xmpp-go development environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Commands:"
            echo "  go test ./...           Run tests"
            echo "  go test -v -run=Fuzz    Run fuzz tests"
            echo "  staticcheck ./...       Static analysis"
          '';
        };
      }
    );
}

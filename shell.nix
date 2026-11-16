{
  pkgs ? import <nixpkgs> { },
}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    golangci-lint
    gnumake
    python3
    nodejs
  ];

  # The pinned backlog CLI (tools/bin/backlog).
  shellHook = ''
    export PATH="$PWD/tools/bin:$PATH"
  '';
}

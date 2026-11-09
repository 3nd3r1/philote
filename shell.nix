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
}

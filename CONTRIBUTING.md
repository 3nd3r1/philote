# Contributing

Start with the [thesis](docs/thesis.md) and the [language reference](docs/language.md). Work is tracked on the board: `backlog board`, or `backlog browser` for the web view.

## Develop

```sh
nix-shell
make setup   # once: enables the pre-commit hook
make check   # lint, tests, architecture rules
make smoke   # build the neighbr test spec, run both apps, send real requests (needs network)
```

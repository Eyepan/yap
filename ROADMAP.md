# Roadmap to v1.0.0

-   [x] Move away from .npmrc, write own config that as development progresses, can be then altered to parse .npmrc files. make .yap_config files binary for performance
-   [ ] Project level .yap_config files
<!-- -   [ ] Better npmrc parsing support (currently doesn't handle //<registry>/:\_auth=<token> properly) -->
-   [x] Cleaner CLI Interface
-   [x] Performance improvements (use pointers)
-   [ ] Install script that gets prebuilt binaries from pre-release/release and adds it to path
-   [x] Better error formatting when panicking
-   [x] Better logging support (by implementing proper logging)
-   [x] Add this header for 'Accept: application/vnd.npm.install-v1+json; q=1.0, application/json; q=0.8, _/_'. More info [here](https://github.com/npm/registry/blob/main/docs/responses/package-metadata.md#abbreviated-metadata-format)
-   [x] Connection pooling to reuse http clients for faster metadata fetching
-   [ ] Faster version resolution (if version is directly resolvable, fetch only that version's metadata instead of fetching the entire metadata file) (clashes with current metadata caching implementation)
-   [x] Map for de-duping instead of unique-ing an array
-   [ ] Symlinked install structure (much akin to pnpm's symlinked node_modules structure)
-   [x] Hash package/version for de-duping instead of unique-ing an array for already installed packages
-   [ ] Follow package.json spec
-   [ ] Add ability to maintain package.json (or even package-lock.json)
-   [x] Add `add` command
-   [-] Add `list` command
-   [ ] Add `update` command
-   [ ] Add `uninstall` command
-   [ ] Checksum verification for files

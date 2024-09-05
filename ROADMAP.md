# Roadmap to v1.0.0

-   [ ] Install script that gets prebuilt binaries from pre-release/release and adds it to path
-   [ ] Cleaner CLI Interface
-   [ ] Better logging support (by implementing proper logging)
-   [ ] Better npmrc parsing support (currently doesn't handle //<registry>/:\_auth=<token> properly)
-   [ ] Add this header for 'Accept: application/vnd.npm.install-v1+json; q=1.0, application/json; q=0.8, _/_'. More info [here](https://github.com/npm/registry/blob/main/docs/responses/package-metadata.md#abbreviated-metadata-format)
-   [ ] Connection pooling to reuse http clients for faster metadata fetching
-   [ ] Faster version resolution (if version is directly resolvable, fetch only that version's metadata instead of fetching the entire metadata file) (clashes with current metadata caching implementation)
-   [ ] Hash package/version for de-duping instead of unique-ing an array
-   [ ] Symlinked install structure (much akin to pnpm's symlinked node_modules structure)
-   [ ] Hash package/version for de-duping instead of unique-ing an array for already installed packages
-   [ ] Follow package.json spec
-   [ ] Add ability to maintain package.json (or even package-lock.json)
-   [ ] Add `add` command
-   [ ] Add `update` command
-   [ ] Add `uninstall` command
-   [ ] Checksum verification for files

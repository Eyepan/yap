| ⚠️ Warning: This project is under VERY VERY active development. Things might break all the time, so please don't use this in production. A few suggested alternatives (that are not written in JS): [Cotton](https://github.com/danielhuang/cotton), [Zap](https://github.com/elbywan/zap), [Orogene](https://github.com/orogene/orogene)

# YAP

Yet-Another-Package manager

Built in Go, made to be faster than pnpm but does the exact same, even down to the content addressable storage.

## Features

-   **Fast and Efficient**: Designed to be faster than pnpm while employing (almost) the same caching mechanisms.
-   **Content Addressable Storage**: Ensures data integrity and deduplication.
-   **Concurrent Downloads**: Utilizes multiple workers to download packages concurrently.
-   **Caching**: Caches metadata and packages to speed up subsequent operations.
-   More to come... Check [here](/ROADMAP.md)

## Installation

To install YAP, clone the repository and build the project:

```sh
git clone https://github.com/Eyepan/yap.git
cd yap
go build -o yap main.go
```

## Usage

To use Yap, run the following command

```
Usage:
./yap <command>
Commands:
help
        prints this out!
install
        installs a list of packages
list
        list out packages from lockfile
add     <package-name>@<!version>
        adds this particular package to package.json and install it in the repository
update <package-name>
        updates the selected package to its latest version
update --all
        updates all dependencies to its latest version
uninstall <package-name>
        removes this package from the list of dependencies
```

A better CLI interface is coming soon. Check out the [Roadmap](/ROADMAP.md) to see when it is coming

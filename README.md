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

```sh
./yap install
```

A better CLI interface is coming soon. Check out the [Roadmap](/ROADMAP.md) to see when it is coming

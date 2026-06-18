# Go Static Server

A tiny Go static file server with browser auto-reload. No dependencies.

*Perfect for quickly iterating on front-end projects, prototypes, and experiments.*

Download pre-built Windows executable (.exe): https://archive.org/download/win-static-server/go-server.zip

## What it does

- Serves everything in `public/` on `http://localhost:8080`.
- Watches `public/` for changes and auto-reloads open browser tabs via Server-Sent Events.
- Injects the reload script into `.html` responses automatically — no changes needed to your HTML files.

## Run from source

Requires Go 1.20+.

```
go run main.go
```

Then open http://localhost:8080. Drop any `.html` file into `public/` and visit it by name (e.g. `http://localhost:8080/about.html`).

## Build a Windows .exe

```
go build -o go-static-server.exe
```

For a smaller binary:

```
go build -ldflags="-s -w" -o go-static-server.exe
```

Cross-compiling from Linux or macOS:

```
GOOS=windows GOARCH=amd64 go build -o go-static-server.exe
```

## Distribute

Ship the exe alongside a `public/` folder:

```
go-static-server/
  go-static-server.exe
  public/
    index.html
    ...
```

Double-click the exe (or run it from a terminal). It will look for `public/` next to itself, so it doesn't matter what directory it's launched from. If no sibling `public/` folder is found, it falls back to `./public` relative to the current working directory — which is what makes `go run` work during development.

## Configuration

Everything is hardcoded for simplicity:

- Port: `:8080` (edit `addr` in `main.go`)
- Served directory: `public/` (edit `publicDir` in `main.go`)
- Poll interval: 500ms (edit `watchLoop` in `main.go`)

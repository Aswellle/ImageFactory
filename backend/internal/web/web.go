// Package web provides the embedded frontend assets for ImageForge.
//
// The frontend is built with Vite (pnpm build) and the resulting dist/ folder
// is embedded into the Go binary at compile time via go:embed. This lets a
// single self-contained binary serve both the API and the SPA.
//
// The dist/ directory is populated by the Dockerfile frontend build stage.
// During local development the frontend is served by Vite's dev server instead.
package web

import "embed"

// Dist holds the compiled frontend assets (HTML, JS, CSS, images).
//
//go:embed all:dist
var Dist embed.FS

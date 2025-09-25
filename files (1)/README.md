# Go E-Commerce Example

A basic e-commerce site built in Go with SQLite, user login, and a simple HTML frontend.

## Features

- User signup/login/logout
- Product list and detail page
- Add to cart (basic, per-user)
- SQLite database (auto-created)
- Minimal HTML/CSS theme

## How to Run

1. Install Go: https://go.dev/dl/
2. Get dependencies:
   ```bash
   go mod tidy
   ```
3. Run the server:
   ```bash
   go run main.go
   ```
4. Open [http://localhost:8080](http://localhost:8080) in browser.

## Deploy on GitHub Pages

- To deploy only the static frontend, copy `templates/` and `static/` content into a `docs/` folder.
- For full backend, deploy on a server/VPS.

## Customize

- Add more products in `initDB` function.
- Improve styles in `static/style.css`.
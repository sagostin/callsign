# CallSign Landing Page

A modern landing page for the CallSign PBX platform.

## Quick Start

```bash
# Start the landing page
docker compose up -d

# Access at http://localhost:8088
```

## Files

- `index.html` - Main landing page
- `style.css` - Styling (dark theme matching panel)
- `script.js` - Interactive features
- `assets/` - Favicon and images
- `Caddyfile` - Web server config
- `docker-compose.yml` - Container setup

## Development

For local development without Docker:
```bash
# Using Python
python -m http.server 8088

# Using Node
npx serve -l 8088
```

## Production

Update the `Caddyfile` with your domain:
```
yourdomain.com {
    root * /srv
    file_server
    # ... rest of config
}
```

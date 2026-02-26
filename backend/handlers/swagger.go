package handlers

import (
	"net/http"
	"os"
	"path/filepath"
)

// ServeSwaggerYAML serves the swagger.yaml file
func ServeSwaggerYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	swaggerPath := filepath.Join(".", "swagger.yaml")

	data, err := os.ReadFile(swaggerPath)
	if err != nil {
		http.Error(w, "Swagger file not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ServeSwaggerUI serves a simple Swagger UI HTML page
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="CodeCollab API Documentation — Real-time code collaboration with SFU voice calling">
    <title>CodeCollab API Documentation</title>
    <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.31.2/favicon-32x32.png" sizes="32x32">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.31.2/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; padding: 0; background: #fafafa; }
        .swagger-ui .topbar { display: none; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.31.2/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.31.2/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: window.location.origin + "/swagger.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                syntaxHighlight: { activated: true, theme: "monokai" },
                tryItOutEnabled: false,
                displayRequestDuration: true,
                filter: true,
                tagsSorter: "alpha",
                operationsSorter: "alpha",
                docExpansion: "list",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 2,
                persistAuthorization: true
            });
            window.ui = ui;
        }
    </script>
</body>
</html>`

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

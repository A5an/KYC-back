package api

import (
	"embed"
)

//go:embed swagger-ui
var SwaggerUI embed.FS

//go:embed swagger.yml
var Swagger string

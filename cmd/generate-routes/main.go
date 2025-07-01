package main

import (
	"log"
)

func main() {
	log.Fatal(`
The route extractor is now a library function. In your project, create:

cmd/extract-routes/main.go:
----------------------------------------
package main

import (
    "log"
    "github.com/karurosux/saas-go-kit/routeextractor"
    "github.com/karurosux/saas-go-kit/auth-go"
    // ... your imports
)

func main() {
    // Create your modules
    modules := map[string]core.Module{
        "auth": auth.NewModule(authConfig),
        "role": role.NewModule(roleConfig),
        // ... your custom modules
    }
    
    // Extract routes
    if err := routeextractor.ExtractRoutes(modules, "./generated/routes.json"); err != nil {
        log.Fatal(err)
    }
}
----------------------------------------

Then run: go run ./cmd/extract-routes
`)
}
package schemas

import (
	"embed"
	"log"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed *.json
var schemaFS embed.FS

var TokenSchema *jsonschema.Schema
var MobileEventSchema *jsonschema.Schema

func Initialize() {
    schemas := [...]string{"token.json", "mobile_event.json"}

    for _, schemaName := range schemas {
        if err := compileSchema(schemaName); err != nil {
            log.Fatalf("Failed to compile JSON schema %s: %v", schemaName, err)
        }
    }
}

func compileSchema(schemaName string) error {
	schemaFile, err := schemaFS.Open(schemaName)
	if err != nil {
        return err
	}

	tokenSchema, err := jsonschema.UnmarshalJSON(schemaFile)
	if err != nil {
        return err
	}

	// Compile the schema
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaName, tokenSchema); err != nil {
        return err
	}

	schema, err := compiler.Compile(schemaName)
	if err != nil {
        return err
	}

	switch schemaName {
	case "token.json":
		TokenSchema = schema
	case "mobile_event.json":
		MobileEventSchema = schema
	}

	return nil
}

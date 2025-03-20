package schemas

import (
	"io/fs"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func CompileSchema(file fs.File) (*jsonschema.Schema, error) {
	parsedSchema, err := jsonschema.UnmarshalJSON(file)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	name := stat.Name()

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(name, parsedSchema); err != nil {
		return nil, err
	}

	schema, err := compiler.Compile(name)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

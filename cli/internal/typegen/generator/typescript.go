package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/barisgit/goflux/cli/internal/typegen/config"
	"github.com/barisgit/goflux/cli/internal/typegen/processor"
	"github.com/barisgit/goflux/cli/internal/typegen/types"
)

// GenerateTypeScriptTypes generates TypeScript type definitions
func GenerateTypeScriptTypes(typeDefs []types.TypeDefinition) error {
	typesDir := filepath.Join("frontend", "src", "types")
	if err := os.MkdirAll(typesDir, 0755); err != nil {
		return fmt.Errorf("failed to create types directory: %w", err)
	}

	// Create processor for case conversion
	proc := processor.NewTypeProcessor(config.DefaultCasingConfig())

	var content strings.Builder
	content.WriteString("// Auto-generated TypeScript types from Go structs\n")
	content.WriteString("// Generated by GoFlux type generation system\n")
	content.WriteString("// Do not edit manually\n\n")

	// Sort types for consistent output
	sort.Slice(typeDefs, func(i, j int) bool {
		return typeDefs[i].Name < typeDefs[j].Name
	})

	for _, t := range typeDefs {
		if t.IsEnum {
			// Generate enum type
			content.WriteString(fmt.Sprintf("export type %s = %s\n\n",
				t.Name, strings.Join(t.EnumValues, " | ")))
		} else {
			// Generate interface
			content.WriteString(fmt.Sprintf("export interface %s {\n", t.Name))

			for _, field := range t.Fields {
				fieldName := field.JSONTag
				if fieldName == "" {
					fieldName = proc.ProcessFieldName(field.Name)
				} else {
					// Process the JSON tag through the field name converter for consistency
					fieldName = proc.ProcessFieldName(fieldName)
				}

				optional := ""
				if field.Optional {
					optional = "?"
				}

				content.WriteString(fmt.Sprintf("  %s%s: %s\n",
					fieldName, optional, field.TypeName))
			}

			content.WriteString("}\n\n")
		}
	}

	typesFile := filepath.Join(typesDir, "generated.d.ts")
	if err := os.WriteFile(typesFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write types file: %w", err)
	}

	return nil
}

// ShouldGenerateTypeScriptTypes returns true if TypeScript types should be generated for the given generator
func ShouldGenerateTypeScriptTypes(generatorType string) bool {
	switch generatorType {
	case "basic":
		return false // JavaScript doesn't need TypeScript types
	case "basic-ts", "axios", "trpc-like":
		return true
	default:
		return true // Default to generating types for unknown generators
	}
}

// collectUsedTypes extracts all types used in the API routes
func collectUsedTypes(routes []types.APIRoute, typeDefs []types.TypeDefinition) []string {
	usedTypes := make(map[string]bool)

	for _, route := range routes {
		if route.RequestType != "" {
			// Extract base type from Omit<Type, 'id'> or Partial<Type>
			requestType := route.RequestType
			if strings.Contains(requestType, "Omit<") {
				start := strings.Index(requestType, "Omit<") + 5
				end := strings.Index(requestType[start:], ",")
				if end > 0 {
					usedTypes[requestType[start:start+end]] = true
				}
			} else if strings.Contains(requestType, "Partial<") {
				start := strings.Index(requestType, "Partial<") + 8
				end := strings.Index(requestType[start:], ">")
				if end > 0 {
					usedTypes[requestType[start:start+end]] = true
				}
			} else {
				usedTypes[requestType] = true
			}
		}
		if route.ResponseType != "" {
			responseType := route.ResponseType
			// Handle array types like "User[]"
			if strings.TrimSuffix(responseType, "[]") != responseType {
				responseType = strings.TrimSuffix(responseType, "[]")
			}
			usedTypes[responseType] = true
		}
	}

	// Filter to only include types that exist in our generated types
	var typeNames []string
	for typeName := range usedTypes {
		for _, t := range typeDefs {
			if t.Name == typeName {
				typeNames = append(typeNames, typeName)
				break
			}
		}
	}

	sort.Strings(typeNames)
	return typeNames
}

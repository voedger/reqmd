package internal

import (
	"fmt"
	"strings"
)

type RequirementType struct {
	Identifier string // The prefix that identifies this type (e.g., "it", "cmp", "utest")
	OrderNo    int    // For ordering in reports and coverage analysis
}

type TypeRegistry struct {
	Types map[string]RequirementType // Map of type identifiers to their definitions
	Order []string                   // Ordered list of type identifiers
}

func NewTypeRegistry(typeDefs []RequirementType) *TypeRegistry {
	registry := &TypeRegistry{
		Types: make(map[string]RequirementType),
		Order: make([]string, len(typeDefs)),
	}

	// Add types to registry and build order list
	for i, typeDef := range typeDefs {
		// Check for duplicate identifiers
		if _, exists := registry.Types[typeDef.Identifier]; exists {
			// Handle duplicate (could return error instead)
			panic("Duplicate requirement type identifier: " + typeDef.Identifier)
		}

		registry.Types[typeDef.Identifier] = typeDef
		registry.Order[i] = typeDef.Identifier
	}

	return registry
}

func (r *TypeRegistry) Type(identifier string) (RequirementType, bool) {
	typeDef, exists := r.Types[identifier]
	return typeDef, exists
}

func ExtractTypeFromRequirement(requirementName string) string {
	// Extract the first segment before the dot
	return strings.Split(requirementName, ".")[0]
}

func ParseTypeList(typeList string) ([]string, error) {
	if typeList == "" {
		return nil, nil
	}

	types := strings.Split(typeList, ",")
	uniqueTypes := make(map[string]bool)
	var result []string

	for _, t := range types {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}

		// Validate identifier matches Name rule
		if !identifierRegex.MatchString(t) {
			return nil, fmt.Errorf("invalid type identifier: %s", t)
		}

		if uniqueTypes[t] {
			return nil, fmt.Errorf("duplicate type identifier: %s", t)
		}

		uniqueTypes[t] = true
		result = append(result, t)
	}

	return result, nil
}

package goconfig

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// ValidatorGenerator generates type-specific validation code
type ValidatorGenerator struct {
	packageName string
}

// NewValidatorGenerator creates a new validator generator
func NewValidatorGenerator(packageName string) *ValidatorGenerator {
	return &ValidatorGenerator{packageName: packageName}
}

// GenerateValidatorCode generates optimized validation code for a struct
func (vg *ValidatorGenerator) GenerateValidatorCode(structCode string) (string, error) {
	// Parse the struct definition
	fset := token.NewFileSet()
	src := fmt.Sprintf("package %s\n%s", vg.packageName, structCode)

	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse struct: %v", err)
	}

	// Find the struct definition
	var structType *ast.StructType
	var structName string

	ast.Inspect(file, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			if st, ok := ts.Type.(*ast.StructType); ok {
				structType = st
				structName = ts.Name.Name
			}
		}
		return true
	})

	if structType == nil {
		return "", fmt.Errorf("no struct found in provided code")
	}

	// Generate the validation function
	return vg.generateValidationFunction(structName, structType), nil
}

func (vg *ValidatorGenerator) generateValidationFunction(structName string, structType *ast.StructType) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("// Generated validator for %s - no reflection!\n", structName))
	builder.WriteString(fmt.Sprintf("func (c *%s) Validate() ValidationErrors {\n", structName))
	builder.WriteString("\tvar errors ValidationErrors\n\n")

	// Generate validation code for each field
	for _, field := range structType.Fields.List {
		if field.Tag == nil {
			continue
		}

		tagValue := strings.Trim(field.Tag.Value, "`")
		validateTag := extractValidateTag(tagValue)
		if validateTag == "" {
			continue
		}

		fieldName := field.Names[0].Name
		fieldType := getFieldType(field.Type)

		builder.WriteString(vg.generateFieldValidation(fieldName, fieldType, validateTag))
		builder.WriteString("\n")
	}

	builder.WriteString("\treturn errors\n")
	builder.WriteString("}\n")

	return builder.String()
}

func (vg *ValidatorGenerator) generateFieldValidation(fieldName, fieldType, validateTag string) string {
	var builder strings.Builder
	rules := parseValidationRulesFromTag(validateTag)

	builder.WriteString(fmt.Sprintf("\t// Validate %s\n", fieldName))

	for ruleName, ruleValue := range rules {
		switch fieldType {
		case "string":
			builder.WriteString(vg.generateStringValidation(fieldName, ruleName, ruleValue))
		case "int", "int32", "int64":
			builder.WriteString(vg.generateIntValidation(fieldName, ruleName, ruleValue))
		case "float32", "float64":
			builder.WriteString(vg.generateFloatValidation(fieldName, ruleName, ruleValue))
		case "bool":
			builder.WriteString(vg.generateBoolValidation(fieldName, ruleName, ruleValue))
		}
	}

	return builder.String()
}

func (vg *ValidatorGenerator) generateStringValidation(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif c.%s == \"\" {\n\t\terrors = append(errors, \"field '%s' is required\")\n\t}\n", fieldName, fieldName)
	case "minlen":
		return fmt.Sprintf("\tif len(c.%s) < %s {\n\t\terrors = append(errors, \"field '%s' must be at least %s characters long\")\n\t}\n", fieldName, ruleValue, fieldName, ruleValue)
	case "maxlen":
		return fmt.Sprintf("\tif len(c.%s) > %s {\n\t\terrors = append(errors, \"field '%s' must be at most %s characters long\")\n\t}\n", fieldName, ruleValue, fieldName, ruleValue)
	case "len":
		return fmt.Sprintf("\tif len(c.%s) != %s {\n\t\terrors = append(errors, \"field '%s' must be exactly %s characters long\")\n\t}\n", fieldName, ruleValue, fieldName, ruleValue)
	case "email":
		return fmt.Sprintf("\tif !emailRegex.MatchString(c.%s) {\n\t\terrors = append(errors, \"field '%s' must be a valid email address\")\n\t}\n", fieldName, fieldName)
	default:
		return fmt.Sprintf("\t// TODO: implement %s validation for %s\n", ruleName, fieldName)
	}
}

func (vg *ValidatorGenerator) generateIntValidation(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif c.%s == 0 {\n\t\terrors = append(errors, \"field '%s' is required\")\n\t}\n", fieldName, fieldName)
	case "min":
		return fmt.Sprintf("\tif c.%s < %s {\n\t\terrors = append(errors, \"field '%s' must be at least %s\")\n\t}\n", fieldName, ruleValue, fieldName, ruleValue)
	case "max":
		return fmt.Sprintf("\tif c.%s > %s {\n\t\terrors = append(errors, \"field '%s' must be at most %s\")\n\t}\n", fieldName, ruleValue, fieldName, ruleValue)
	default:
		return fmt.Sprintf("\t// TODO: implement %s validation for %s\n", ruleName, fieldName)
	}
}

func (vg *ValidatorGenerator) generateFloatValidation(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif c.%s == 0.0 {\n\t\terrors = append(errors, \"field '%s' is required\")\n\t}\n", fieldName, fieldName)
	case "min":
		return fmt.Sprintf("\tif c.%s < %s {\n\t\terrors = append(errors, \"field '%s' must be at least %s\")\n\t}\n", fieldName, ruleValue, fieldName, ruleValue)
	case "max":
		return fmt.Sprintf("\tif c.%s > %s {\n\t\terrors = append(errors, \"field '%s' must be at most %s\")\n\t}\n", fieldName, ruleValue, fieldName, ruleValue)
	default:
		return fmt.Sprintf("\t// TODO: implement %s validation for %s\n", ruleName, fieldName)
	}
}

func (vg *ValidatorGenerator) generateBoolValidation(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !c.%s {\n\t\terrors = append(errors, \"field '%s' is required\")\n\t}\n", fieldName, fieldName)
	default:
		return fmt.Sprintf("\t// TODO: implement %s validation for %s\n", ruleName, fieldName)
	}
}

// Helper functions

func extractValidateTag(tagString string) string {
	// Parse: validate:"required,minlen=3"
	parts := strings.Split(tagString, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "validate:") {
			return strings.Trim(strings.TrimPrefix(part, "validate:"), "\"")
		}
	}
	return ""
}

func parseValidationRulesFromTag(validateTag string) map[string]string {
	rules := make(map[string]string)
	parts := strings.Split(validateTag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			rules[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		} else {
			rules[part] = ""
		}
	}

	return rules
}

func getFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return getFieldType(t.X)
	default:
		return "unknown"
	}
}

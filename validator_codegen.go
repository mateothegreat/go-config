package goconfig

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
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

	// Add import statements
	builder.WriteString("import (\n")
	builder.WriteString("\t\"github.com/mateothegreat/go-config/validation\"\n")
	builder.WriteString(")\n\n")

	builder.WriteString(fmt.Sprintf("// Generated validator for %s - zero reflection, maximum performance\n", structName))
	builder.WriteString(fmt.Sprintf("func (c *%s) Validate() validation.ValidationErrors {\n", structName))
	builder.WriteString("\tvar errors validation.ValidationErrors\n\n")

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
	builder.WriteString("}\n\n")

	// Generate validation result with field-specific errors
	builder.WriteString(fmt.Sprintf("// ValidateWithResult validates %s and returns detailed results\n", structName))
	builder.WriteString(fmt.Sprintf("func (c *%s) ValidateWithResult() *validation.ValidationResult {\n", structName))
	builder.WriteString("\tresult := validation.NewValidationResult()\n\n")

	// Generate field-specific validation with detailed error tracking
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

		builder.WriteString(vg.generateFieldValidationWithResult(fieldName, fieldType, validateTag))
		builder.WriteString("\n")
	}

	builder.WriteString("\treturn result\n")
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
		case "float32":
			builder.WriteString(vg.generateFloat32Validation(fieldName, ruleName, ruleValue))
		case "float64":
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
		return fmt.Sprintf("\tif !validation.IsRequired(c.%s) {\n\t\terrors |= validation.ErrRequired\n\t}\n", fieldName)
	case "minlen":
		return fmt.Sprintf("\tif !validation.HasMinLength(c.%s, %s) {\n\t\terrors |= validation.ErrMinLength\n\t}\n", fieldName, ruleValue)
	case "maxlen":
		return fmt.Sprintf("\tif !validation.HasMaxLength(c.%s, %s) {\n\t\terrors |= validation.ErrMaxLength\n\t}\n", fieldName, ruleValue)
	case "len":
		return fmt.Sprintf("\tif !validation.HasLength(c.%s, %s) {\n\t\terrors |= validation.ErrLength\n\t}\n", fieldName, ruleValue)
	case "email":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsEmail(c.%s) {\n\t\terrors |= validation.ErrEmail\n\t}\n", fieldName, fieldName)
	case "url":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsURL(c.%s) {\n\t\terrors |= validation.ErrURL\n\t}\n", fieldName, fieldName)
	case "alpha":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsAlpha(c.%s) {\n\t\terrors |= validation.ErrAlpha\n\t}\n", fieldName, fieldName)
	case "alphanum":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsAlphaNum(c.%s) {\n\t\terrors |= validation.ErrAlphaNum\n\t}\n", fieldName, fieldName)
	case "numeric":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsNumeric(c.%s) {\n\t\terrors |= validation.ErrNumeric\n\t}\n", fieldName, fieldName)
	case "base64":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsBase64(c.%s) {\n\t\terrors |= validation.ErrBase64\n\t}\n", fieldName, fieldName)
	case "uuid":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsUUID(c.%s) {\n\t\terrors |= validation.ErrUUID\n\t}\n", fieldName, fieldName)
	case "ip":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsIP(c.%s) {\n\t\terrors |= validation.ErrIP\n\t}\n", fieldName, fieldName)
	case "ipv4":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsIPv4(c.%s) {\n\t\terrors |= validation.ErrIPv4\n\t}\n", fieldName, fieldName)
	case "ipv6":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsIPv6(c.%s) {\n\t\terrors |= validation.ErrIPv6\n\t}\n", fieldName, fieldName)
	case "hostname":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsHostname(c.%s) {\n\t\terrors |= validation.ErrHostname\n\t}\n", fieldName, fieldName)
	case "cidr":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsCIDR(c.%s) {\n\t\terrors |= validation.ErrCIDR\n\t}\n", fieldName, fieldName)
	case "mac":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsMAC(c.%s) {\n\t\terrors |= validation.ErrMAC\n\t}\n", fieldName, fieldName)
	case "oneof":
		// Handle both pipe and space separators
		var values []string
		if strings.Contains(ruleValue, "|") {
			values = strings.Split(ruleValue, "|")
		} else {
			values = strings.Split(ruleValue, " ")
		}
		var quotedValues []string
		for _, v := range values {
			v = strings.TrimSpace(v)
			if v != "" {
				quotedValues = append(quotedValues, fmt.Sprintf("\"%s\"", v))
			}
		}
		return fmt.Sprintf("\tif !validation.IsOneOfString(c.%s, []string{%s}) {\n\t\terrors |= validation.ErrOneOf\n\t}\n", fieldName, strings.Join(quotedValues, ", "))
	default:
		return fmt.Sprintf("\t// TODO: implement %s validation for %s\n", ruleName, fieldName)
	}
}

func (vg *ValidatorGenerator) generateIntValidation(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !validation.IsRequiredInt(c.%s) {\n\t\terrors |= validation.ErrRequired\n\t}\n", fieldName)
	case "min":
		return fmt.Sprintf("\tif !validation.HasMinInt(c.%s, %s) {\n\t\terrors |= validation.ErrMin\n\t}\n", fieldName, ruleValue)
	case "max":
		return fmt.Sprintf("\tif !validation.HasMaxInt(c.%s, %s) {\n\t\terrors |= validation.ErrMax\n\t}\n", fieldName, ruleValue)
	case "oneof":
		// Handle both pipe and space separators
		var values []string
		if strings.Contains(ruleValue, "|") {
			values = strings.Split(ruleValue, "|")
		} else {
			values = strings.Split(ruleValue, " ")
		}
		var intValues []string
		for _, v := range values {
			v = strings.TrimSpace(v)
			if v != "" {
				if _, err := strconv.Atoi(v); err == nil {
					intValues = append(intValues, v)
				}
			}
		}
		return fmt.Sprintf("\tif !validation.IsOneOfInt(c.%s, []int{%s}) {\n\t\terrors |= validation.ErrOneOf\n\t}\n", fieldName, strings.Join(intValues, ", "))
	default:
		return fmt.Sprintf("\t// TODO: implement %s validation for %s\n", ruleName, fieldName)
	}
}

func (vg *ValidatorGenerator) generateFloatValidation(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !validation.IsRequiredFloat64(c.%s) {\n\t\terrors |= validation.ErrRequired\n\t}\n", fieldName)
	case "min":
		return fmt.Sprintf("\tif !validation.HasMinFloat64(c.%s, %s) {\n\t\terrors |= validation.ErrMin\n\t}\n", fieldName, ruleValue)
	case "max":
		return fmt.Sprintf("\tif !validation.HasMaxFloat64(c.%s, %s) {\n\t\terrors |= validation.ErrMax\n\t}\n", fieldName, ruleValue)
	default:
		return fmt.Sprintf("\t// TODO: implement %s validation for %s\n", ruleName, fieldName)
	}
}

func (vg *ValidatorGenerator) generateFloat32Validation(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !validation.IsRequiredFloat32(c.%s) {\n\t\terrors |= validation.ErrRequired\n\t}\n", fieldName)
	case "min":
		return fmt.Sprintf("\tif !validation.HasMinFloat32(c.%s, %s) {\n\t\terrors |= validation.ErrMin\n\t}\n", fieldName, ruleValue)
	case "max":
		return fmt.Sprintf("\tif !validation.HasMaxFloat32(c.%s, %s) {\n\t\terrors |= validation.ErrMax\n\t}\n", fieldName, ruleValue)
	default:
		return fmt.Sprintf("\t// TODO: implement %s validation for %s\n", ruleName, fieldName)
	}
}

func (vg *ValidatorGenerator) generateBoolValidation(fieldName, ruleName, _ string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !validation.IsRequiredBool(c.%s) {\n\t\terrors |= validation.ErrRequired\n\t}\n", fieldName)
	default:
		return fmt.Sprintf("\t// TODO: implement %s validation for %s\n", ruleName, fieldName)
	}
}

func (vg *ValidatorGenerator) generateFieldValidationWithResult(fieldName, fieldType, validateTag string) string {
	var builder strings.Builder
	rules := parseValidationRulesFromTag(validateTag)

	builder.WriteString(fmt.Sprintf("\t// Validate %s with detailed results\n", fieldName))

	for ruleName, ruleValue := range rules {
		switch fieldType {
		case "string":
			builder.WriteString(vg.generateStringValidationWithResult(fieldName, ruleName, ruleValue))
		case "int", "int32", "int64":
			builder.WriteString(vg.generateIntValidationWithResult(fieldName, ruleName, ruleValue))
		case "float32":
			builder.WriteString(vg.generateFloat32ValidationWithResult(fieldName, ruleName, ruleValue))
		case "float64":
			builder.WriteString(vg.generateFloatValidationWithResult(fieldName, ruleName, ruleValue))
		case "bool":
			builder.WriteString(vg.generateBoolValidationWithResult(fieldName, ruleName, ruleValue))
		}
	}

	return builder.String()
}

func (vg *ValidatorGenerator) generateStringValidationWithResult(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !validation.IsRequired(c.%s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrRequired)\n\t}\n", fieldName, fieldName)
	case "minlen":
		return fmt.Sprintf("\tif !validation.HasMinLength(c.%s, %s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrMinLength)\n\t}\n", fieldName, ruleValue, fieldName)
	case "maxlen":
		return fmt.Sprintf("\tif !validation.HasMaxLength(c.%s, %s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrMaxLength)\n\t}\n", fieldName, ruleValue, fieldName)
	case "len":
		return fmt.Sprintf("\tif !validation.HasLength(c.%s, %s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrLength)\n\t}\n", fieldName, ruleValue, fieldName)
	case "email":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsEmail(c.%s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrEmail)\n\t}\n", fieldName, fieldName, fieldName)
	case "url":
		return fmt.Sprintf("\tif c.%s != \"\" && !validation.IsURL(c.%s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrURL)\n\t}\n", fieldName, fieldName, fieldName)
	default:
		return ""
	}
}

func (vg *ValidatorGenerator) generateIntValidationWithResult(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !validation.IsRequiredInt(c.%s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrRequired)\n\t}\n", fieldName, fieldName)
	case "min":
		return fmt.Sprintf("\tif !validation.HasMinInt(c.%s, %s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrMin)\n\t}\n", fieldName, ruleValue, fieldName)
	case "max":
		return fmt.Sprintf("\tif !validation.HasMaxInt(c.%s, %s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrMax)\n\t}\n", fieldName, ruleValue, fieldName)
	default:
		return ""
	}
}

func (vg *ValidatorGenerator) generateFloatValidationWithResult(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !validation.IsRequiredFloat64(c.%s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrRequired)\n\t}\n", fieldName, fieldName)
	case "min":
		return fmt.Sprintf("\tif !validation.HasMinFloat64(c.%s, %s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrMin)\n\t}\n", fieldName, ruleValue, fieldName)
	case "max":
		return fmt.Sprintf("\tif !validation.HasMaxFloat64(c.%s, %s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrMax)\n\t}\n", fieldName, ruleValue, fieldName)
	default:
		return ""
	}
}

func (vg *ValidatorGenerator) generateFloat32ValidationWithResult(fieldName, ruleName, ruleValue string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !validation.IsRequiredFloat32(c.%s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrRequired)\n\t}\n", fieldName, fieldName)
	case "min":
		return fmt.Sprintf("\tif !validation.HasMinFloat32(c.%s, %s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrMin)\n\t}\n", fieldName, ruleValue, fieldName)
	case "max":
		return fmt.Sprintf("\tif !validation.HasMaxFloat32(c.%s, %s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrMax)\n\t}\n", fieldName, ruleValue, fieldName)
	default:
		return ""
	}
}

func (vg *ValidatorGenerator) generateBoolValidationWithResult(fieldName, ruleName, _ string) string {
	switch ruleName {
	case "required":
		return fmt.Sprintf("\tif !validation.IsRequiredBool(c.%s) {\n\t\tresult.AddFieldError(\"%s\", validation.ErrRequired)\n\t}\n", fieldName, fieldName)
	default:
		return ""
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

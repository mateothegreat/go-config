package generator

import (
	"fmt"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/mateothegreat/go-config/internal/scanner"
)

// CodeGenerator handles generating zero-reflection validation code
type CodeGenerator struct {
	config GeneratorConfig
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator(config GeneratorConfig) *CodeGenerator {
	return &CodeGenerator{
		config: config,
	}
}

// GenerateValidationCode generates validation code for a struct
func (cg *CodeGenerator) GenerateValidationCode(structInfo scanner.StructInfo) (string, error) {
	tmpl := template.New("validation").Funcs(template.FuncMap{
		"hasRule":        hasRule,
		"getRuleValue":   getRuleValue,
		"capitalize":     capitalize,
		"lower":          strings.ToLower,
		"quote":          func(s string) string { return fmt.Sprintf(`"%s"`, s) },
		"buildValidator": cg.buildValidatorCode,
	})

	tmpl, err := tmpl.Parse(validationTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var builder strings.Builder
	data := TemplateData{
		PackageName: structInfo.PackageName,
		Structs:     []scanner.StructInfo{structInfo},
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	err = tmpl.Execute(&builder, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return builder.String(), nil
}

// GenerateFileContent generates complete file content with multiple structs
func (cg *CodeGenerator) GenerateFileContent(structs []scanner.StructInfo, packageName string) (string, error) {
	if len(structs) == 0 {
		return "", fmt.Errorf("no structs provided")
	}

	tmpl := template.New("file").Funcs(template.FuncMap{
		"hasRule":        hasRule,
		"getRuleValue":   getRuleValue,
		"capitalize":     capitalize,
		"lower":          strings.ToLower,
		"quote":          func(s string) string { return fmt.Sprintf(`"%s"`, s) },
		"buildValidator": cg.buildValidatorCode,
		"extractImports": cg.extractImports,
	})

	tmpl, err := tmpl.Parse(fileTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse file template: %w", err)
	}

	var builder strings.Builder
	data := TemplateData{
		PackageName: packageName,
		Structs:     structs,
		Timestamp:   time.Now().Format(time.RFC3339),
		Imports:     cg.extractImports(structs),
	}

	err = tmpl.Execute(&builder, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute file template: %w", err)
	}

	return builder.String(), nil
}

// buildValidatorCode generates validation code for a specific field and rule
func (cg *CodeGenerator) buildValidatorCode(field scanner.FieldInfo, rule string, value string) string {
	switch rule {
	case "required":
		return cg.buildRequiredValidator(field)
	case "min":
		return cg.buildMinValidator(field, value)
	case "max":
		return cg.buildMaxValidator(field, value)
	case "minlen":
		return cg.buildMinLenValidator(field, value)
	case "maxlen":
		return cg.buildMaxLenValidator(field, value)
	case "len":
		return cg.buildLenValidator(field, value)
	case "email":
		return cg.buildEmailValidator(field)
	case "url":
		return cg.buildURLValidator(field)
	case "alpha":
		return cg.buildAlphaValidator(field)
	case "alphanumeric":
		return cg.buildAlphaNumericValidator(field)
	case "numeric":
		return cg.buildNumericValidator(field)
	case "regex":
		return cg.buildRegexValidator(field, value)
	case "oneof":
		return cg.buildOneOfValidator(field, value)
	case "range":
		return cg.buildRangeValidator(field, value)
	default:
		return fmt.Sprintf("// Unknown validation rule: %s", rule)
	}
}

// buildRequiredValidator generates code for required field validation
func (cg *CodeGenerator) buildRequiredValidator(field scanner.FieldInfo) string {
	switch field.Type {
	case "string":
		return fmt.Sprintf(`
	if s.%s == "" {
		return goconfig.NewError().Field("%s").Required()
	}`, field.Name, field.Name)
	case "int", "int8", "int16", "int32", "int64":
		return fmt.Sprintf(`
	if s.%s == 0 {
		return goconfig.NewError().Field("%s").Required()
	}`, field.Name, field.Name)
	default:
		return fmt.Sprintf(`
	if s.%s == nil || reflect.ValueOf(s.%s).IsZero() {
		return goconfig.NewError().Field("%s").Required()
	}`, field.Name, field.Name, field.Name)
	}
}

// capitalize returns a string with the first letter capitalized
// Replaces deprecated strings.Title for simple title case
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// buildMinValidator generates code for minimum value validation
func (cg *CodeGenerator) buildMinValidator(field scanner.FieldInfo, minValue string) string {
	switch {
	case strings.HasPrefix(field.Type, "int"):
		return fmt.Sprintf(`
	if s.%s < %s {
		return goconfig.NewError().Field("%s").Format("must be at least %s")
	}`, field.Name, minValue, field.Name, minValue)
	case strings.HasPrefix(field.Type, "float"):
		return fmt.Sprintf(`
	if s.%s < %s {
		return goconfig.NewError().Field("%s").Format("must be at least %s")
	}`, field.Name, minValue, field.Name, minValue)
	default:
		return fmt.Sprintf("// Min validation not supported for type %s", field.Type)
	}
}

// buildMaxValidator generates code for maximum value validation
func (cg *CodeGenerator) buildMaxValidator(field scanner.FieldInfo, maxValue string) string {
	switch {
	case strings.HasPrefix(field.Type, "int"):
		return fmt.Sprintf(`
	if s.%s > %s {
		return goconfig.NewError().Field("%s").Format("must be at most %s")
	}`, field.Name, maxValue, field.Name, maxValue)
	case strings.HasPrefix(field.Type, "float"):
		return fmt.Sprintf(`
	if s.%s > %s {
		return goconfig.NewError().Field("%s").Format("must be at most %s")
	}`, field.Name, maxValue, field.Name, maxValue)
	default:
		return fmt.Sprintf("// Max validation not supported for type %s", field.Type)
	}
}

// buildMinLenValidator generates code for minimum length validation
func (cg *CodeGenerator) buildMinLenValidator(field scanner.FieldInfo, minLen string) string {
	if field.Type == "string" {
		return fmt.Sprintf(`
	if len(s.%s) < %s {
		return goconfig.NewError().Field("%s").MinLength(%s)
	}`, field.Name, minLen, field.Name, minLen)
	}
	return fmt.Sprintf("// MinLen validation not supported for type %s", field.Type)
}

// buildMaxLenValidator generates code for maximum length validation
func (cg *CodeGenerator) buildMaxLenValidator(field scanner.FieldInfo, maxLen string) string {
	if field.Type == "string" {
		return fmt.Sprintf(`
	if len(s.%s) > %s {
		return goconfig.NewError().Field("%s").MaxLength(%s)
	}`, field.Name, maxLen, field.Name, maxLen)
	}
	return fmt.Sprintf("// MaxLen validation not supported for type %s", field.Type)
}

// buildLenValidator generates code for exact length validation
func (cg *CodeGenerator) buildLenValidator(field scanner.FieldInfo, expectedLen string) string {
	if field.Type == "string" {
		return fmt.Sprintf(`
	if len(s.%s) != %s {
		return goconfig.NewError().Field("%s").Format("must be exactly %s characters long")
	}`, field.Name, expectedLen, field.Name, expectedLen)
	}
	return fmt.Sprintf("// Len validation not supported for type %s", field.Type)
}

// buildEmailValidator generates code for email validation
func (cg *CodeGenerator) buildEmailValidator(field scanner.FieldInfo) string {
	if field.Type == "string" {
		return fmt.Sprintf(`
	if !emailRegex.MatchString(s.%s) {
		return goconfig.NewError().Field("%s").Format("must be a valid email address")
	}`, field.Name, field.Name)
	}
	return fmt.Sprintf("// Email validation not supported for type %s", field.Type)
}

// buildURLValidator generates code for URL validation
func (cg *CodeGenerator) buildURLValidator(field scanner.FieldInfo) string {
	if field.Type == "string" {
		return fmt.Sprintf(`
	if !urlRegex.MatchString(s.%s) {
		return goconfig.NewError().Field("%s").Format("must be a valid URL")
	}`, field.Name, field.Name)
	}
	return fmt.Sprintf("// URL validation not supported for type %s", field.Type)
}

// buildAlphaValidator generates code for alphabetic validation
func (cg *CodeGenerator) buildAlphaValidator(field scanner.FieldInfo) string {
	if field.Type == "string" {
		return fmt.Sprintf(`
	if !alphaRegex.MatchString(s.%s) {
		return goconfig.NewError().Field("%s").Format("must contain only alphabetic characters")
	}`, field.Name, field.Name)
	}
	return fmt.Sprintf("// Alpha validation not supported for type %s", field.Type)
}

// buildAlphaNumericValidator generates code for alphanumeric validation
func (cg *CodeGenerator) buildAlphaNumericValidator(field scanner.FieldInfo) string {
	if field.Type == "string" {
		return fmt.Sprintf(`
	if !alphaNumRegex.MatchString(s.%s) {
		return goconfig.NewError().Field("%s").Format("must contain only alphanumeric characters")
	}`, field.Name, field.Name)
	}
	return fmt.Sprintf("// AlphaNumeric validation not supported for type %s", field.Type)
}

// buildNumericValidator generates code for numeric validation
func (cg *CodeGenerator) buildNumericValidator(field scanner.FieldInfo) string {
	if field.Type == "string" {
		return fmt.Sprintf(`
	if !numericRegex.MatchString(s.%s) {
		return goconfig.NewError().Field("%s").Format("must contain only numeric characters")
	}`, field.Name, field.Name)
	}
	return fmt.Sprintf("// Numeric validation not supported for type %s", field.Type)
}

// buildRegexValidator generates code for regex pattern validation
func (cg *CodeGenerator) buildRegexValidator(field scanner.FieldInfo, pattern string) string {
	if field.Type == "string" {
		regexVarName := fmt.Sprintf("%sRegex", strings.ToLower(field.Name))
		return fmt.Sprintf(`
	%s := regexp.MustCompile(%s)
	if !%s.MatchString(s.%s) {
		return goconfig.NewError().Field("%s").Format("does not match required pattern")
	}`, regexVarName, pattern, regexVarName, field.Name, field.Name)
	}
	return fmt.Sprintf("// Regex validation not supported for type %s", field.Type)
}

// buildOneOfValidator generates code for one-of validation
func (cg *CodeGenerator) buildOneOfValidator(field scanner.FieldInfo, options string) string {
	if field.Type == "string" {
		optionsList := strings.Split(options, "|")
		var conditions []string
		for _, option := range optionsList {
			option = strings.TrimSpace(option)
			conditions = append(conditions, fmt.Sprintf(`s.%s == "%s"`, field.Name, option))
		}
		condition := strings.Join(conditions, " || ")

		return fmt.Sprintf(`
	if !(%s) {
		return goconfig.NewError().Field("%s").Format("must be one of [%s]")
	}`, condition, field.Name, options)
	}
	return fmt.Sprintf("// OneOf validation not supported for type %s", field.Type)
}

// buildRangeValidator generates code for range validation
func (cg *CodeGenerator) buildRangeValidator(field scanner.FieldInfo, rangeValue string) string {
	parts := strings.Split(rangeValue, ":")
	if len(parts) != 2 {
		return fmt.Sprintf("// Invalid range format: %s", rangeValue)
	}

	min, max := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

	switch {
	case strings.HasPrefix(field.Type, "int"):
		return fmt.Sprintf(`
	if s.%s < %s || s.%s > %s {
		return goconfig.NewError().Field("%s").Range(%s, %s)
	}`, field.Name, min, field.Name, max, field.Name, min, max)
	case strings.HasPrefix(field.Type, "float"):
		return fmt.Sprintf(`
	if s.%s < %s || s.%s > %s {
		return goconfig.NewError().Field("%s").Range(%s, %s)
	}`, field.Name, min, field.Name, max, field.Name, min, max)
	default:
		return fmt.Sprintf("// Range validation not supported for type %s", field.Type)
	}
}

// extractImports extracts required imports for generated code
func (cg *CodeGenerator) extractImports(structs []scanner.StructInfo) []string {
	importSet := make(map[string]bool)

	// Always include these imports for validation
	importSet["fmt"] = true
	importSet["regexp"] = true
	importSet["github.com/mateothegreat/go-config"] = true
	importSet["github.com/mateothegreat/go-config/validate"] = true

	// Check if reflect is needed
	for _, structInfo := range structs {
		for _, field := range structInfo.Fields {
			if field.Required && !isPrimitiveType(field.Type) {
				importSet["reflect"] = true
				break
			}
		}
	}

	var imports []string
	for imp := range importSet {
		imports = append(imports, imp)
	}

	return imports
}

// isPrimitiveType checks if a type is a Go primitive type
func isPrimitiveType(typeName string) bool {
	primitiveTypes := map[string]bool{
		"string": true,
		"int":    true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true,
		"bool": true,
		"byte": true, "rune": true,
	}
	return primitiveTypes[typeName]
}

// Template helper functions
func hasRule(field scanner.FieldInfo, rule string) bool {
	_, exists := field.ValidateRules[rule]
	return exists
}

func getRuleValue(field scanner.FieldInfo, rule string) string {
	return field.ValidateRules[rule]
}

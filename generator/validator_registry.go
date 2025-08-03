package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// ValidatorRegistry manages the mapping between validation rules and go-validation library validators
type ValidatorRegistry struct {
	// Maps rule names to their corresponding go-validation validators
	validators map[string]ValidatorInfo
	// Maps rule names to AST expressions for code generation
	expressions map[string]ast.Expr
}

// ValidatorInfo contains information about a validator
type ValidatorInfo struct {
	Name         string
	PackagePath  string
	FunctionName string
	RequiresArgs bool
	ArgType      string // "int", "string", "float", etc.
}

// NewValidatorRegistry creates a new validator registry
func NewValidatorRegistry() *ValidatorRegistry {
	r := &ValidatorRegistry{
		validators:  make(map[string]ValidatorInfo),
		expressions: make(map[string]ast.Expr),
	}
	r.registerBuiltinValidators()
	return r
}

// registerBuiltinValidators registers all built-in validators from go-validation library
func (r *ValidatorRegistry) registerBuiltinValidators() {
	// Basic validators
	r.validators["required"] = ValidatorInfo{
		Name:         "required",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Required",
		RequiresArgs: false,
	}

	// String validators
	r.validators["email"] = ValidatorInfo{
		Name:         "email",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Email",
		RequiresArgs: false,
	}

	r.validators["url"] = ValidatorInfo{
		Name:         "url",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "URL",
		RequiresArgs: false,
	}

	r.validators["alpha"] = ValidatorInfo{
		Name:         "alpha",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Alpha",
		RequiresArgs: false,
	}

	r.validators["alphanum"] = ValidatorInfo{
		Name:         "alphanum",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "AlphaNum",
		RequiresArgs: false,
	}

	r.validators["numeric"] = ValidatorInfo{
		Name:         "numeric",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Numeric",
		RequiresArgs: false,
	}

	// Numeric validators
	r.validators["min"] = ValidatorInfo{
		Name:         "min",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Min",
		RequiresArgs: true,
		ArgType:      "numeric",
	}

	r.validators["max"] = ValidatorInfo{
		Name:         "max",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Max",
		RequiresArgs: true,
		ArgType:      "numeric",
	}

	// Length validators
	r.validators["len"] = ValidatorInfo{
		Name:         "len",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Length",
		RequiresArgs: true,
		ArgType:      "int",
	}

	r.validators["minlen"] = ValidatorInfo{
		Name:         "minlen",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "MinLength",
		RequiresArgs: true,
		ArgType:      "int",
	}

	r.validators["maxlen"] = ValidatorInfo{
		Name:         "maxlen",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "MaxLength",
		RequiresArgs: true,
		ArgType:      "int",
	}

	// Network validators
	r.validators["ip"] = ValidatorInfo{
		Name:         "ip",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "IP",
		RequiresArgs: false,
	}

	r.validators["ipv4"] = ValidatorInfo{
		Name:         "ipv4",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "IPv4",
		RequiresArgs: false,
	}

	r.validators["ipv6"] = ValidatorInfo{
		Name:         "ipv6",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "IPv6",
		RequiresArgs: false,
	}

	r.validators["cidr"] = ValidatorInfo{
		Name:         "cidr",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "CIDR",
		RequiresArgs: false,
	}

	r.validators["mac"] = ValidatorInfo{
		Name:         "mac",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "MAC",
		RequiresArgs: false,
	}

	r.validators["hostname"] = ValidatorInfo{
		Name:         "hostname",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Hostname",
		RequiresArgs: false,
	}

	// Format validators
	r.validators["uuid"] = ValidatorInfo{
		Name:         "uuid",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "UUID",
		RequiresArgs: false,
	}

	r.validators["uuid4"] = ValidatorInfo{
		Name:         "uuid4",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "UUIDv4",
		RequiresArgs: false,
	}

	r.validators["datetime"] = ValidatorInfo{
		Name:         "datetime",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "DateTime",
		RequiresArgs: false,
	}

	r.validators["date"] = ValidatorInfo{
		Name:         "date",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Date",
		RequiresArgs: false,
	}

	r.validators["time"] = ValidatorInfo{
		Name:         "time",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Time",
		RequiresArgs: false,
	}

	r.validators["json"] = ValidatorInfo{
		Name:         "json",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "JSON",
		RequiresArgs: false,
	}

	r.validators["base64"] = ValidatorInfo{
		Name:         "base64",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Base64",
		RequiresArgs: false,
	}

	r.validators["creditcard"] = ValidatorInfo{
		Name:         "creditcard",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "CreditCard",
		RequiresArgs: false,
	}

	r.validators["phone"] = ValidatorInfo{
		Name:         "phone",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Phone",
		RequiresArgs: false,
	}

	// Conditional validators
	r.validators["oneof"] = ValidatorInfo{
		Name:         "oneof",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "OneOf",
		RequiresArgs: true,
		ArgType:      "string",
	}

	// Cross-field validators
	r.validators["eqfield"] = ValidatorInfo{
		Name:         "eqfield",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "EqField",
		RequiresArgs: true,
		ArgType:      "field",
	}

	r.validators["nefield"] = ValidatorInfo{
		Name:         "nefield",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "NeField",
		RequiresArgs: true,
		ArgType:      "field",
	}

	r.validators["gtfield"] = ValidatorInfo{
		Name:         "gtfield",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "GtField",
		RequiresArgs: true,
		ArgType:      "field",
	}

	r.validators["ltfield"] = ValidatorInfo{
		Name:         "ltfield",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "LtField",
		RequiresArgs: true,
		ArgType:      "field",
	}

	// Conditional validators
	r.validators["required_if"] = ValidatorInfo{
		Name:         "required_if",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "RequiredIf",
		RequiresArgs: true,
		ArgType:      "conditional",
	}

	r.validators["required_unless"] = ValidatorInfo{
		Name:         "required_unless",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "RequiredUnless",
		RequiresArgs: true,
		ArgType:      "conditional",
	}

	r.validators["required_with"] = ValidatorInfo{
		Name:         "required_with",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "RequiredWith",
		RequiresArgs: true,
		ArgType:      "field",
	}

	r.validators["required_without"] = ValidatorInfo{
		Name:         "required_without",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "RequiredWithout",
		RequiresArgs: true,
		ArgType:      "field",
	}

	// Collection validators
	r.validators["dive"] = ValidatorInfo{
		Name:         "dive",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Dive",
		RequiresArgs: false,
	}

	// Regex validator
	r.validators["regex"] = ValidatorInfo{
		Name:         "regex",
		PackagePath:  "github.com/mateothegreat/go-validation/rules",
		FunctionName: "Regex",
		RequiresArgs: true,
		ArgType:      "string",
	}
}

// GetValidatorExpression returns an AST expression for a validator
func (r *ValidatorRegistry) GetValidatorExpression(rule, value string, fieldType string) ast.Expr {
	info, exists := r.validators[rule]
	if !exists {
		// If not found in registry, try to get from go-validation dynamically
		return r.createDynamicValidatorExpression(rule)
	}

	// Extract package alias from path
	parts := strings.Split(info.PackagePath, "/")
	pkgAlias := parts[len(parts)-1]

	if !info.RequiresArgs {
		// Simple validator without arguments
		return &ast.SelectorExpr{
			X:   ast.NewIdent(pkgAlias),
			Sel: ast.NewIdent(info.FunctionName),
		}
	}

	// Validator with arguments
	return r.createParameterizedValidator(info, value, fieldType, pkgAlias)
}

// createParameterizedValidator creates a validator expression with parameters
func (r *ValidatorRegistry) createParameterizedValidator(info ValidatorInfo, value, fieldType, pkgAlias string) ast.Expr {
	var args []ast.Expr

	switch info.ArgType {
	case "int":
		args = append(args, &ast.BasicLit{
			Kind:  token.INT,
			Value: value,
		})
	case "string":
		// Handle special cases like oneof
		if info.Name == "oneof" {
			// Split values by space or pipe
			options := strings.Split(value, " ")
			for _, opt := range options {
				opt = strings.TrimSpace(opt)
				if opt != "" {
					args = append(args, &ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprintf(`"%s"`, opt),
					})
				}
			}
		} else {
			args = append(args, &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf(`"%s"`, value),
			})
		}
	case "numeric":
		// Determine if int or float based on field type
		if strings.Contains(fieldType, "float") {
			args = append(args, &ast.BasicLit{
				Kind:  token.FLOAT,
				Value: value,
			})
		} else {
			args = append(args, &ast.BasicLit{
				Kind:  token.INT,
				Value: value,
			})
		}
	case "field":
		// For cross-field validation
		args = append(args, &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf(`"%s"`, value),
		})
	case "conditional":
		// Parse conditional like "Field=value"
		parts := strings.SplitN(value, "=", 2)
		if len(parts) == 2 {
			args = append(args,
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf(`"%s"`, parts[0]),
				},
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf(`"%s"`, parts[1]),
				},
			)
		}
	}

	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(pkgAlias),
			Sel: ast.NewIdent(info.FunctionName),
		},
		Args: args,
	}
}

// createDynamicValidatorExpression creates an expression for validators not in registry
func (r *ValidatorRegistry) createDynamicValidatorExpression(rule string) ast.Expr {
	// Try to get validator from go-validation registry at runtime
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent("validation"),
			Sel: ast.NewIdent("GetValidator"),
		},
		Args: []ast.Expr{
			&ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf(`"%s"`, rule),
			},
		},
	}
}

// GetRequiredImports returns the required imports for the validators used
func (r *ValidatorRegistry) GetRequiredImports(rules []string) map[string]string {
	imports := make(map[string]string)

	// Always include the base validation library
	imports["validation"] = "github.com/mateothegreat/go-validation"

	// Check which validator packages are needed
	needsRules := false
	for _, rule := range rules {
		if info, exists := r.validators[rule]; exists {
			if strings.Contains(info.PackagePath, "/rules") {
				needsRules = true
			}
		}
	}

	if needsRules {
		imports["rules"] = "github.com/mateothegreat/go-validation/rules"
	}

	return imports
}

// IsValidRule checks if a rule is valid
func (r *ValidatorRegistry) IsValidRule(rule string) bool {
	_, exists := r.validators[rule]
	return exists
}

// GetValidatorInfo returns information about a validator
func (r *ValidatorRegistry) GetValidatorInfo(rule string) (ValidatorInfo, bool) {
	info, exists := r.validators[rule]
	return info, exists
}

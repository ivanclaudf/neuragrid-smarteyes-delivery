
# Template Documentation

This document provides information about creating and using templates in the Delivery Service.

---

## Template Variables

Templates support variable substitution using Go's [`text/template`](https://pkg.go.dev/text/template) package. This allows you to create dynamic templates where variables are replaced with actual values at runtime.

### Basic Usage

To include a variable in your template content, use the `{{ .VariableName }}` syntax. For example:

```text
Hello {{.name}}, your order {{.orderNumber}} has been confirmed!
```

When sending a message, provide the variable values in the `params` object:

```json
{
  "params": {
    "name": "John",
    "orderNumber": "ORD-12345"
  }
}
```

The system will render the template and replace the variables with the provided values before sending:

```text
Hello John, your order ORD-12345 has been confirmed!
```

### Template Functions

Go templates support a range of functions for formatting and conditional logic:

#### Basic Functions

- `{{.variable}}` — Insert a variable value
- `{{if .condition}}...{{else}}...{{end}}` — Conditional blocks
- `{{range .items}}...{{end}}` — Loop over items

#### String Operations

- `{{.variable | upper}}` — Convert to uppercase
- `{{.variable | lower}}` — Convert to lowercase
- `{{.variable | title}}` — Title case

### Example Templates

#### Simple Greeting

```text
Hello {{.name}}! Thank you for using our service.
```

#### Conditional Content

```text
Hello {{.name}}!
{{if .isPremium}}
Thank you for being a premium customer!
{{else}}
Consider upgrading to our premium plan for additional benefits.
{{end}}
```

#### Alert with Formatting

```text
ALERT: {{.alertType | upper}}
Time: {{.timestamp}}
Description: {{.description}}
{{if .actionRequired}}
Action Required: {{.action}}
{{end}}
```

---

## Best Practices

1. Always provide default values for required variables
2. Test your templates thoroughly before using them in production
3. Keep templates simple and readable
4. Document the required variables for each template

---

## Technical Details

Template rendering is performed using Go's standard library `text/template` package. The variables are passed as a map from the request's `params` field and rendered just before sending the message to the provider.

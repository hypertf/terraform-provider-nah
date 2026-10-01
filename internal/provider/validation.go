package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// The API measures string limits in bytes rather than Unicode code points.
type apiString struct {
	min, max int
	pattern  *regexp.Regexp
}

func (v apiString) Description(context.Context) string {
	s := fmt.Sprintf("Must contain at least %d bytes", v.min)
	if v.max > 0 {
		s += fmt.Sprintf(" and at most %d bytes", v.max)
	}
	if v.pattern != nil {
		s += " and match " + v.pattern.String()
	}
	return s + "."
}
func (v apiString) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }
func (v apiString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	s := req.ConfigValue.ValueString()
	if len(s) < v.min || (v.max > 0 && len(s) > v.max) || (v.pattern != nil && !v.pattern.MatchString(s)) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid value", v.Description(ctx))
	}
}

var slugValidator = apiString{1, 63, regexp.MustCompile(`^[a-z][a-z0-9-]*$`)}
var nameValidator = apiString{1, 255, regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)}

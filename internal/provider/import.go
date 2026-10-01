package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func importScoped(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, attributes ...string) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != len(attributes) {
		resp.Diagnostics.AddError("Invalid import ID", "Expected "+strings.Join(attributes, "/")+".")
		return
	}
	for _, part := range parts {
		if strings.TrimSpace(part) == "" || part == "." || part == ".." {
			resp.Diagnostics.AddError("Invalid import ID", "Import segments must be nonempty identifiers.")
			return
		}
	}
	for i, attribute := range attributes {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attribute), parts[i])...)
	}
}

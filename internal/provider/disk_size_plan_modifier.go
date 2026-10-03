package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type replaceOnShrink struct{}

func (replaceOnShrink) Description(context.Context) string {
	return "Disk expansion updates in place; shrinking replaces the disk."
}

func (m replaceOnShrink) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }

func (replaceOnShrink) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	resp.RequiresReplace = req.PlanValue.ValueInt64() < req.StateValue.ValueInt64()
}

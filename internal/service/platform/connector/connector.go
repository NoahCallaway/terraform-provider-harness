package connector

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/antihax/optional"
	"github.com/harness/harness-go-sdk/harness/nextgen"
	"github.com/harness/terraform-provider-harness/helpers"
	"github.com/harness/terraform-provider-harness/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ReadConnectorData func(*schema.ResourceData, *nextgen.ConnectorInfo) error

func resourceConnectorReadBase(ctx context.Context, d *schema.ResourceData, meta interface{}, connType nextgen.ConnectorType) (*nextgen.ConnectorInfo, diag.Diagnostics) {
	c, ctx := meta.(*internal.Session).GetPlatformClientWithContext(ctx)

	if id, ok := d.GetOk("identifier"); ok {
		return readConnectorByID(ctx, c, d, id.(string), connType)
	} else if name, ok := d.GetOk("name"); ok {
		getReadConnectorOpts2(d)
		return readConnectorByName(ctx, c, d, name.(string), connType)
	} else{
		return nil,diag.FromErr(errors.New("Either Identifier or Name must be specified"))
	}
}

func readConnectorByID(ctx context.Context, c *nextgen.APIClient, d *schema.ResourceData, id string, connType nextgen.ConnectorType) (*nextgen.ConnectorInfo, diag.Diagnostics) {
	resp, httpResp, err := c.ConnectorsApi.GetConnector(ctx, c.AccountId, id, getReadConnectorOpts(d))
	if err != nil {
		return nil, helpers.HandleReadApiError(err, d, httpResp)
	}

	if connType != resp.Data.Connector.Type_ {
		return nil, diag.FromErr(fmt.Errorf("expected connector to be of type %s, but got %s", connType, resp.Data.Connector.Type_))
	}

	readCommonConnectorData(d, resp.Data.Connector)

	return resp.Data.Connector, nil
}

func readConnectorByName(ctx context.Context, c *nextgen.APIClient, d *schema.ResourceData, name string, connType nextgen.ConnectorType) (*nextgen.ConnectorInfo, diag.Diagnostics) {
	var pageIndex int32 = 0
	var pageSize int32 = 2
	opts := getReadConnectorOpts2(d)
	filters := &nextgen.ConnectorFilterProperties{
		ConnectorNames: []string{name},
		Types:          []string{string(connType)},
		FilterType:     nextgen.ConnectorFilterTypes.Connector,
	}

	resp, httpResp, err := c.ConnectorsApi.GetConnectorListV2(ctx, *filters, c.AccountId, &nextgen.ConnectorsApiGetConnectorListV2Opts{
		PageIndex:         optional.NewInt32(pageIndex),
		PageSize:          optional.NewInt32(pageSize),
		OrgIdentifier:     opts.OrgIdentifier,
		ProjectIdentifier: opts.ProjectIdentifier,
	})
	if err != nil {
		return nil, helpers.HandleReadApiError(err, d, httpResp)
	}
	if len(resp.Data.Content) == 0 {
		return nil, nil
	}

	for _, svc := range resp.Data.Content {
		if svc.Connector.Name == name{
			readCommonConnectorData(d, svc.Connector)
			return svc.Connector, nil
		}
	}

	pageIndex += pageSize

	return nil, nil
}

func dataConnectorReadBase(ctx context.Context, d *schema.ResourceData, meta interface{}, connType nextgen.ConnectorType) (*nextgen.ConnectorInfo, diag.Diagnostics) {
	c, ctx := meta.(*internal.Session).GetPlatformClientWithContext(ctx)

	id := d.Id()
	if id == "" {
		id = d.Get("identifier").(string)
	}

	resp, httpResp, err := c.ConnectorsApi.GetConnector(ctx, c.AccountId, id, getReadConnectorOpts(d))
	if err != nil {
		return nil, helpers.HandleApiError(err, d, httpResp)
	}

	if connType != resp.Data.Connector.Type_ {
		return nil, diag.FromErr(fmt.Errorf("expected connector to be of type %s, but got %s", connType, resp.Data.Connector.Type_))
	}

	readCommonConnectorData(d, resp.Data.Connector)

	return resp.Data.Connector, nil
}

func getReadConnectorOpts(d *schema.ResourceData) *nextgen.ConnectorsApiGetConnectorOpts {
	connOpts := &nextgen.ConnectorsApiGetConnectorOpts{}

	if attr, ok := d.GetOk("org_id"); ok {
		connOpts.OrgIdentifier = optional.NewString(attr.(string))
	}

	if attr, ok := d.GetOk("project_id"); ok {
		connOpts.ProjectIdentifier = optional.NewString(attr.(string))
	}

	if attr, ok := d.GetOk("git_sync"); ok {
		opts := attr.([]interface{})[0].(map[string]interface{})
		connOpts.Branch = optional.NewString(opts["branch"].(string))
		connOpts.RepoIdentifier = optional.NewString(opts["repo_id"].(string))
	}

	return connOpts
}

func getReadConnectorOpts2(d *schema.ResourceData) *nextgen.ConnectorsApiGetConnectorByNameOpts {
	connOpts := &nextgen.ConnectorsApiGetConnectorByNameOpts{}

	if attr, ok := d.GetOk("org_id"); ok {
		connOpts.OrgIdentifier = optional.NewString(attr.(string))
	}

	if attr, ok := d.GetOk("project_id"); ok {
		connOpts.ProjectIdentifier = optional.NewString(attr.(string))
	}

	return connOpts
}

func resourceConnectorCreateOrUpdateBase(ctx context.Context, d *schema.ResourceData, meta interface{}, connector *nextgen.ConnectorInfo) (*nextgen.ConnectorInfo, diag.Diagnostics) {
	c, ctx := meta.(*internal.Session).GetPlatformClientWithContext(ctx)

	id := d.Id()
	buildConnector(d, connector)

	var err error
	var resp nextgen.ResponseDtoConnectorResponse
	var httpResp *http.Response

	if id == "" {
		resp, httpResp, err = c.ConnectorsApi.CreateConnector(ctx, nextgen.Connector{Connector: connector}, c.AccountId, &nextgen.ConnectorsApiCreateConnectorOpts{})
	} else {
		resp, httpResp, err = c.ConnectorsApi.UpdateConnector(ctx, nextgen.Connector{Connector: connector}, c.AccountId, &nextgen.ConnectorsApiUpdateConnectorOpts{})
	}

	if err != nil {
		return nil, helpers.HandleApiError(err, d, httpResp)
	}

	readCommonConnectorData(d, resp.Data.Connector)

	return resp.Data.Connector, nil
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ctx := meta.(*internal.Session).GetPlatformClientWithContext(ctx)

	_, httpResp, err := c.ConnectorsApi.DeleteConnector(ctx, c.AccountId, d.Id(), &nextgen.ConnectorsApiDeleteConnectorOpts{
		OrgIdentifier:     helpers.BuildField(d, "org_id"),
		ProjectIdentifier: helpers.BuildField(d, "project_id")})

	if err != nil {
		return helpers.HandleApiError(err, d, httpResp)
	}

	return nil
}

func buildConnector(d *schema.ResourceData, connector *nextgen.ConnectorInfo) {
	if attr := d.Get("name").(string); attr != "" {
		connector.Name = attr
	}

	if attr := d.Get("identifier").(string); attr != "" {
		connector.Identifier = attr
	}

	if attr := d.Get("description").(string); attr != "" {
		connector.Description = attr
	}

	if attr := d.Get("org_id").(string); attr != "" {
		connector.OrgIdentifier = attr
	}

	if attr := d.Get("project_id").(string); attr != "" {
		connector.ProjectIdentifier = attr
	}

	if attr := d.Get("tags").(*schema.Set).List(); len(attr) > 0 {
		connector.Tags = helpers.ExpandTags(attr)
	}
}

func readCommonConnectorData(d *schema.ResourceData, connector *nextgen.ConnectorInfo) {
	d.SetId(connector.Identifier)
	d.Set("identifier", connector.Identifier)
	d.Set("description", connector.Description)
	d.Set("name", connector.Name)
	d.Set("org_id", connector.OrgIdentifier)
	d.Set("project_id", connector.ProjectIdentifier)
	d.Set("tags", helpers.FlattenTags(connector.Tags))
}

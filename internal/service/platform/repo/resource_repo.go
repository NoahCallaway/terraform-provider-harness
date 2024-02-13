package repo

import (
	"context"
	"net/http"

	"github.com/antihax/optional"
	"github.com/harness/harness-go-sdk/harness/code"
	"github.com/harness/terraform-provider-harness/helpers"
	"github.com/harness/terraform-provider-harness/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type RepoBody struct {
	Identifier    string `json:"identifier"`
	DefaultBranch string `json:"default_branch"`
	Description   string `json:"description"`
	IsPublic      bool   `json:"is_public"`
	Readme        bool   `json:"readme"`
	License       string `json:"license"`
	GitIgnore     string `json:"git_ignore"`
}

type Provider struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Type     string `json:"type"`
	Username string `json:"username"`
}

type RepoImportBody struct {
	Description  string   `json:"description"`
	Identifier   string   `json:"identifier"`
	Provider     Provider `json:"provider"`
	ProviderRepo string   `json:"provider_repo"`
}

func ResourceRepo() *schema.Resource {
	resource := &schema.Resource{
		Description: "Resource for creating a Harness Repo.",

		ReadContext:   resourceRepoRead,
		CreateContext: resourceRepoCreateOrUpdate,
		UpdateContext: resourceRepoCreateOrUpdate,
		DeleteContext: resourceRepoDelete,
		Importer:      helpers.MultiLevelResourceImporter,

		Schema: createSchema(),
	}

	helpers.SetMultiLevelResourceSchema(resource.Schema)

	return resource
}

func resourceRepoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ctx := meta.(*internal.Session).GetCodeClientWithContext(ctx)

	id := d.Id()
	if id == "" {
		d.MarkNewResource()
		return nil
	}
	accId := d.Get("account_id").(string)
	orgId := d.Get("org_id").(string)
	projId := d.Get("project_id").(string)

	repo, resp, err := c.RepositoryApi.FindRepository(
		ctx,
		accId,
		id,
		&code.RepositoryApiFindRepositoryOpts{
			OrgIdentifier:     optional.NewString(orgId),
			ProjectIdentifier: optional.NewString(projId),
		},
	)
	if err != nil {
		return helpers.HandleReadApiError(err, d, resp)
	}

	readRepo(d, &repo)

	return nil
}

func resourceRepoCreateOrUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	c, ctx := meta.(*internal.Session).GetCodeClientWithContext(ctx)
	var err error
	var repo code.TypesRepository
	var resp *http.Response

	accId := d.Get("account_id").(string)
	orgId := d.Get("org_id").(string)
	projId := d.Get("project_id").(string)

	var sourceRepo string
	source := source(d)
	if source != nil {
		sourceRepo = source["repo"].(string)
	}

	// Import repo
	if sourceRepo != "" {
		body := optional.NewInterface(buildRepoImportBody(d))
		repo, resp, err = c.RepositoryApi.ImportRepository(
			ctx, accId, &code.RepositoryApiImportRepositoryOpts{
				Body:              body,
				OrgIdentifier:     optional.NewString(orgId),
				ProjectIdentifier: optional.NewString(projId),
			},
		)
		if err != nil {
			return helpers.HandleApiError(err, d, resp)
		}

		readRepo(d, &repo)
		return nil
	}

	// Create repo
	body := optional.NewInterface(buildRepoBody(d))
	id := d.Id()
	if id == "" {
		repo, resp, err = c.RepositoryApi.CreateRepository(
			ctx, accId,
			&code.RepositoryApiCreateRepositoryOpts{
				Body:              body,
				OrgIdentifier:     optional.NewString(orgId),
				ProjectIdentifier: optional.NewString(projId),
			},
		)
		if err != nil {
			return helpers.HandleApiError(err, d, resp)
		}
		readRepo(d, &repo)
		return nil
	}

	// Update repo
	repo, resp, err = c.RepositoryApi.UpdateRepository(
		ctx,
		accId,
		d.Get("identifier").(string),
		&code.RepositoryApiUpdateRepositoryOpts{
			Body:              optional.NewInterface(body),
			OrgIdentifier:     optional.NewString(orgId),
			ProjectIdentifier: optional.NewString(projId),
		},
	)
	if err != nil {
		return helpers.HandleApiError(err, d, resp)
	}
	readRepo(d, &repo)
	return nil
}

func resourceRepoDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	c, ctx := meta.(*internal.Session).GetCodeClientWithContext(ctx)

	id := d.Id()

	accountId := d.Get("account_id").(string)
	orgId := d.Get("org_id").(string)
	projId := d.Get("project_id").(string)

	resp, err := c.RepositoryApi.DeleteRepository(
		ctx,
		accountId,
		id, &code.RepositoryApiDeleteRepositoryOpts{
			OrgIdentifier:     optional.NewString(orgId),
			ProjectIdentifier: optional.NewString(projId),
		},
	)
	if err != nil {
		return helpers.HandleApiError(err, d, resp)
	}

	return nil
}

func buildRepoBody(d *schema.ResourceData) *RepoBody {
	return &RepoBody{
		DefaultBranch: d.Get("default_branch").(string),
		Description:   d.Get("description").(string),
		GitIgnore:     d.Get("git_ignore").(string),
		Identifier:    d.Get("identifier").(string),
		IsPublic:      d.Get("is_public").(bool),
		License:       d.Get("license").(string),
		Readme:        d.Get("readme").(bool),
	}
}

func source(d *schema.ResourceData) map[string]interface{} {
	srcSet := d.Get("source").(*schema.Set)
	srcList := srcSet.List()
	if len(srcList) == 0 {
		return nil
	}
	srcElem := srcList[0]
	return srcElem.(map[string]interface{})
}

func buildRepoImportBody(d *schema.ResourceData) *RepoImportBody {
	importBody := &RepoImportBody{
		Description: d.Get("description").(string),
		Identifier:  d.Get("identifier").(string),
	}

	source := source(d)
	if source != nil {
		importBody.Provider = Provider{
			Host:     source["host"].(string),
			Password: source["password"].(string),
			Type:     source["type"].(string),
			Username: source["username"].(string),
		}
		importBody.ProviderRepo = source["repo"].(string)
	}

	return importBody
}

func readRepo(d *schema.ResourceData, repo *code.TypesRepository) {
	d.SetId(repo.Identifier)
	d.Set("name", repo.Identifier)

	d.Set("identifier", repo.Identifier)
	d.Set("created", repo.Created)
	d.Set("created_by", repo.CreatedBy)
	d.Set("default_branch", repo.DefaultBranch)
	d.Set("description", repo.Description)
	d.Set("git_url", repo.GitUrl)
	d.Set("id", repo.Id)
	d.Set("importing", repo.Importing)
	d.Set("is_public", repo.IsPublic)
	d.Set("num_closed_pulls", repo.NumClosedPulls)
	d.Set("num_merged_pulls", repo.NumMergedPulls)
	d.Set("num_open_pulls", repo.NumOpenPulls)
	d.Set("num_pulls", repo.NumPulls)
	d.Set("path", repo.Path)
	d.Set("size", repo.Size)
	d.Set("size_updated", repo.SizeUpdated)
	d.Set("updated", repo.Updated)
}

func createSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"account_id": {
			Description: "ID of the account who created the repository.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Description: "Name of the resource.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"created_by": {
			Description: "ID of the user who created the repository.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"created": {
			Description: "Timestamp when the repository was created.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"default_branch": {
			Description: "Default branch of the repository.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"description": {
			Description: "Description of the repository.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"git_url": {
			Description: "Git URL of the repository.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"id": {
			Description: "ID of the repository.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"importing": {
			Description: "Whether the repository is being imported.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"is_public": {
			Description: "Whether the repository is public.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"num_closed_pulls": {
			Description: "Number of closed pull requests.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"num_merged_pulls": {
			Description: "Number of merged pull requests.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"num_open_pulls": {
			Description: "Number of open pull requests.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"num_pulls": {
			Description: "Total number of pull requests.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"path": {
			Description: "Path of the repository.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"size": {
			Description: "Size of the repository.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"size_updated": {
			Description: "Timestamp when the repository size was last updated.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"updated": {
			Description: "Timestamp when the repository was last updated.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"git_ignore": {
			Description: "Gitignore file for the repository.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"identifier": {
			Description: "Identifier of the repository.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"license": {
			Description: "License for the repository.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"readme": {
			Description: "Whether the repository has a readme.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"source": {
			Description: "Provider related configurations.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"username": {
						Description: "The username for authentication when importing.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"password": {
						Description: "The password for authentication when importing.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"type": {
						Description: "The type of SCM provider (e.g., github) when importing.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"host": {
						Description: "The host URL for the import source.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"repo": {
						Description: "The provider repository to import from.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
	}
}

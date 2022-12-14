package chart

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nais/knorten/pkg/database"
	"github.com/nais/knorten/pkg/database/crypto"
	"github.com/nais/knorten/pkg/database/gensql"
	helmApps "github.com/nais/knorten/pkg/helm/applications"
	"github.com/nais/knorten/pkg/k8s"
	"github.com/nais/knorten/pkg/reflect"
	"github.com/sirupsen/logrus"
	"strconv"
)

type JupyterhubClient struct {
	repo         *database.Repo
	k8sClient    *k8s.Client
	cryptClient  *crypto.EncrypterDecrypter
	chartVersion string
	log          *logrus.Entry
}

type JupyterForm struct {
	TeamID string
	Slug   string
	JupyterValues
}

func (v *JupyterConfigurableValues) MemoryWithoutUnit() string {
	if v.MemoryLimit == "" {
		return ""
	}
	return v.MemoryLimit[:len(v.MemoryLimit)-1]
}

type JupyterConfigurableValues struct {
	CPULimit        string `form:"cpu" helm:"singleuser.cpu.limit"`
	CPUGuarantee    string `form:"cpu" helm:"singleuser.cpu.guarantee"`
	MemoryLimit     string `form:"memory" helm:"singleuser.memory.limit"`
	MemoryGuarantee string `form:"memory" helm:"singleuser.memory.guarantee"`
	ImageName       string `form:"imagename" helm:"singleuser.image.name"`
	ImageTag        string `form:"imagetag" helm:"singleuser.image.tag"`
	CullTimeout     string `form:"culltimeout" helm:"cull.timeout"`
}

type JupyterValues struct {
	JupyterConfigurableValues

	// Generated config
	AdminUsers       []string `helm:"hub.config.Authenticator.admin_users"`
	AllowedUsers     []string `helm:"hub.config.Authenticator.allowed_users"`
	Hosts            string   `helm:"ingress.hosts"`
	IngressTLS       string   `helm:"ingress.tls"`
	ServiceAccount   string   `helm:"singleuser.serviceAccountName"`
	OAuthCallbackURL string   `helm:"hub.config.AzureAdOAuthenticator.oauth_callback_url"`
	KnadaTeamSecret  string   `helm:"singleuser.extraEnv.KNADA_TEAM_SECRET"`
	ProfileList      string   `helm:"singleuser.profileList"`
}

func NewJupyterhubClient(repo *database.Repo, k8sClient *k8s.Client, cryptClient *crypto.EncrypterDecrypter, chartVersion string, log *logrus.Entry) JupyterhubClient {
	return JupyterhubClient{
		repo:         repo,
		k8sClient:    k8sClient,
		cryptClient:  cryptClient,
		chartVersion: chartVersion,
		log:          log,
	}
}

func (j JupyterhubClient) Create(c *gin.Context, slug string) error {
	var form JupyterForm
	err := c.ShouldBindWith(&form, binding.Form)
	if err != nil {
		return err
	}

	team, err := j.repo.TeamGet(c, slug)
	if err != nil {
		return err
	}

	existing, err := j.repo.TeamValuesGet(c, gensql.ChartTypeJupyterhub, team.ID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return fmt.Errorf("there already exists a jupyterhub for team '%v'", team.ID)
	}

	if team.PendingJupyterUpgrade {
		j.log.Info("pending jupyterhub install")
		return nil
	}

	form.Slug = slug
	form.TeamID = team.ID
	form.AdminUsers = team.Users
	form.AllowedUsers = team.Users

	addGeneratedJupyterhubConfig(&form)

	return j.UpdateTeamValuesAndInstallOrUpdate(c, form)
}

func (j JupyterhubClient) Update(c *gin.Context, form JupyterForm) error {
	team, err := j.repo.TeamGet(c, form.Slug)
	if err != nil {
		return err
	}
	if team.PendingJupyterUpgrade {
		j.log.Info("pending jupyterhub upgrade")
		return nil
	}

	form.TeamID = team.ID
	form.AdminUsers = team.Users
	form.AllowedUsers = team.Users

	return j.UpdateTeamValuesAndInstallOrUpdate(c, form)
}

func (j JupyterhubClient) UpdateTeamValuesAndInstallOrUpdate(ctx context.Context, form JupyterForm) error {
	err := form.formatValues()
	if err != nil {
		return err
	}

	if form.ImageName != "" && form.ImageTag != "" {
		err := j.addCustomImage(&form)
		if err != nil {
			return err
		}
	}

	if err := j.storeJupyterTeamValues(ctx, form); err != nil {
		return err
	}

	return j.Sync(ctx, form.TeamID)
}

func (j JupyterhubClient) Sync(ctx context.Context, teamID string) error {
	application := helmApps.NewJupyterhub(teamID, j.repo, j.cryptClient, j.chartVersion)
	charty, err := application.Chart(ctx)
	if err != nil {
		return err
	}

	// Release name must be unique across namespaces as the helm chart creates a clusterrole
	// for each jupyterhub with the same name as the release name.
	releaseName := JupyterReleaseName(k8s.NameToNamespace(teamID))
	return j.k8sClient.CreateHelmUpgradeJob(ctx, teamID, releaseName, charty.Values)
}

func (j JupyterhubClient) Delete(c context.Context, teamSlug string) error {
	team, err := j.repo.TeamGet(c, teamSlug)
	if err != nil {
		return err
	}
	if team.PendingJupyterUpgrade {
		j.log.Info("pending jupyterhub install")
		return nil
	}

	if err := j.repo.AppDelete(c, team.ID, gensql.ChartTypeJupyterhub); err != nil {
		return err
	}

	namespace := k8s.NameToNamespace(team.ID)
	releaseName := JupyterReleaseName(namespace)

	return j.k8sClient.CreateHelmUninstallJob(c, team.ID, releaseName)
}

func (j JupyterhubClient) addCustomImage(form *JupyterForm) error {
	type kubespawnerOverride struct {
		Image string `json:"image"`
	}

	type profile struct {
		DisplayName         string              `json:"display_name"`
		Description         string              `json:"description"`
		KubespawnerOverride kubespawnerOverride `json:"kubespawner_override"`
	}

	profileList := []profile{{
		DisplayName: "Custom image",
		Description: fmt.Sprintf("Custom image for team %v", form.Slug),
		KubespawnerOverride: kubespawnerOverride{
			Image: fmt.Sprintf("%v:%v", form.ImageName, form.ImageTag),
		},
	}}

	profilesBytes, err := json.Marshal(profileList)
	if err != nil {
		return err
	}
	form.ProfileList = string(profilesBytes)

	return nil
}

func (j JupyterhubClient) storeJupyterTeamValues(ctx context.Context, form JupyterForm) error {
	chartValues, err := reflect.CreateChartValues(form.JupyterValues)
	if err != nil {
		return err
	}

	err = j.repo.TeamValuesInsert(ctx, gensql.ChartTypeJupyterhub, chartValues, form.TeamID)
	if err != nil {
		return err
	}

	return nil
}

func JupyterReleaseName(namespace string) string {
	return fmt.Sprintf("%v-%v", string(gensql.ChartTypeJupyterhub), namespace)
}

func addGeneratedJupyterhubConfig(values *JupyterForm) {
	values.Hosts = fmt.Sprintf("[\"%v\"]", values.Slug+".jupyter.knada.io")
	values.IngressTLS = fmt.Sprintf("[{\"hosts\":[\"%v\"], \"secretName\": \"%v\"}]", values.Slug+".jupyter.knada.io", "jupyterhub-certificate")
	values.ServiceAccount = values.TeamID
	values.OAuthCallbackURL = fmt.Sprintf("https://%v.jupyter.knada.io/hub/oauth_callback", values.Slug)
	values.KnadaTeamSecret = fmt.Sprintf("projects/knada-gcp/secrets/%v", values.TeamID)
}
func (v *JupyterConfigurableValues) formatValues() error {
	floatVal, err := strconv.ParseFloat(v.CPUGuarantee, 64)
	if err != nil {
		return err
	}
	v.CPUGuarantee = fmt.Sprintf("%.1f", floatVal)

	floatVal, err = strconv.ParseFloat(v.CPULimit, 64)
	if err != nil {
		return err
	}
	v.CPULimit = fmt.Sprintf("%.1f", floatVal)

	_, err = strconv.ParseFloat(v.MemoryGuarantee, 64)
	if err != nil {
		return err
	}
	v.MemoryGuarantee = v.MemoryGuarantee + "G"

	_, err = strconv.ParseFloat(v.MemoryLimit, 64)
	if err != nil {
		return err
	}
	v.MemoryLimit = v.MemoryLimit + "G"

	return nil
}

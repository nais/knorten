package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nais/knorten/pkg/database/gensql"
	"github.com/nais/knorten/pkg/helm"
	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HelmAction string

const (
	InstallOrUpgrade HelmAction = "install-or-upgrade"
	Uninstall        HelmAction = "uninstall"

	namespace                = "knada-system"
	saName                   = "knorten"
	ttlSecondsAfterFinished  = 60
	backoffLimit             = 1
	helmRepoConfigMap        = "helm-repos"
	helmRepoConfigMapSubPath = "repositories.yaml"
	helmRepoConfigMountPath  = "/root/.config/helm/repositories.yaml"
)

func CreateJobPrefix(teamID, chartType, action string) string {
	return fmt.Sprintf("%v-%v-%v-", teamID, chartType, action)
}

func (c *Client) CreateHelmUpgradeJob(ctx context.Context, teamID, releaseName string, values map[string]any) error {
	if c.dryRun {
		c.log.Infof("NOOP: Running in dry run mode")
		out, err := yaml.Marshal(values)
		if err != nil {
			c.log.WithError(err).Errorf("error while marshaling chart for %v", releaseName)
			return err
		}

		err = os.WriteFile(fmt.Sprintf("%v.yaml", releaseName), out, 0o644)
		if err != nil {
			c.log.WithError(err).Errorf("error while writing to file %v.yaml", releaseName)
			return err
		}
		return nil
	}

	encValues, err := c.encryptValues(values)
	if err != nil {
		return err
	}

	chartType := helm.ReleaseNameToChartType(releaseName)
	chartVersion, err := c.versionForChart(chartType)
	if err != nil {
		return err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: CreateJobPrefix(teamID, chartType, string(InstallOrUpgrade)),
			Namespace:    namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  string(InstallOrUpgrade),
							Image: c.knelmImage,
							Command: []string{
								"/app/knelm",
								fmt.Sprintf("--action=%v", string(InstallOrUpgrade)),
								fmt.Sprintf("--releasename=%v", releaseName),
								fmt.Sprintf("--team=%v", teamID),
							},
							EnvFrom: []v1.EnvFromSource{
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "knelm",
										},
									},
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "HELM_VALUES",
									Value: encValues,
								},
								{
									Name:  "CHART_VERSION",
									Value: chartVersion,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "helm-repos-config",
									MountPath: helmRepoConfigMountPath,
									SubPath:   helmRepoConfigMapSubPath,
								},
							},
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: saName,
					Volumes: []v1.Volume{
						{
							Name: "helm-repos-config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: helmRepoConfigMap,
									},
								},
							},
						},
					},
				},
			},
			TTLSecondsAfterFinished: intToInt32Ptr(ttlSecondsAfterFinished),
			BackoffLimit:            intToInt32Ptr(backoffLimit),
		},
	}

	_, err = c.clientSet.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	if err := c.repo.TeamSetPendingUpgrade(ctx, teamID, helm.ReleaseNameToChartType(releaseName), true); err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateHelmUninstallJob(ctx context.Context, teamID, releaseName string) error {
	if c.dryRun {
		c.log.Infof("NOOP: Running in dry run mode")
		return nil
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: CreateJobPrefix(teamID, helm.ReleaseNameToChartType(releaseName), string(Uninstall)),
			Namespace:    namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  string(Uninstall),
							Image: c.knelmImage,
							Command: []string{
								"/app/knelm",
								fmt.Sprintf("--action=%v", string(Uninstall)),
								fmt.Sprintf("--releasename=%v", releaseName),
								fmt.Sprintf("--team=%v", teamID),
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "helm-repos-config",
									MountPath: helmRepoConfigMountPath,
									SubPath:   helmRepoConfigMapSubPath,
								},
							},
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: saName,
					Volumes: []v1.Volume{
						{
							Name: "helm-repos-config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: helmRepoConfigMap,
									},
								},
							},
						},
					},
				},
			},
			TTLSecondsAfterFinished: intToInt32Ptr(ttlSecondsAfterFinished),
			BackoffLimit:            intToInt32Ptr(backoffLimit),
		},
	}

	_, err := c.clientSet.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) encryptValues(values map[string]any) (string, error) {
	data, err := json.Marshal(values)
	if err != nil {
		return "", err
	}

	valuesEncoded := base64.StdEncoding.EncodeToString(data)
	return c.cryptClient.EncryptValue(valuesEncoded)
}

func (c *Client) versionForChart(chartType string) (string, error) {
	switch chartType {
	case string(gensql.ChartTypeAirflow):
		return c.airflowChartVersion, nil
	case string(gensql.ChartTypeJupyterhub):
		return c.jupyterChartVersion, nil
	default:
		return "", fmt.Errorf("chart type %v does not exist", chartType)
	}
}

func intToInt32Ptr(val int) *int32 {
	valInt32 := int32(val)
	return &valInt32
}

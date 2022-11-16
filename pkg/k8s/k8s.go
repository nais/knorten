package k8s

import (
	"context"
	"errors"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type Client struct {
	clientSet  *kubernetes.Clientset
	dryRun     bool
	inCluster  bool
	gcpProject string
	gcpRegion  string
}

func New(dryRun, inCluster bool, gcpProject, gcpRegion string) (*Client, error) {
	client := &Client{
		dryRun:     dryRun,
		gcpProject: gcpProject,
		gcpRegion:  gcpRegion,
	}

	config, err := createConfig(inCluster)
	if err != nil {
		return nil, err
	}

	client.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) CreateTeamNamespace(ctx context.Context, name string) error {
	if c.dryRun {
		return nil
	}

	nsSpec := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"cert-secret-jupyterhub":   "include",
				"cert-secret-airflow":      "include",
				"azureadapp-secret":        "include",
				"smtp-secret":              "include",
				"slack-secret":             "include",
				"github-app-secret":        "include",
				"ghcr-secret":              "include",
				"ca-bundle-cm":             "include",
				"airflow-webserver-config": "include",
				"airflow-auth-config":      "include",
			},
		},
	}

	_, err := c.clientSet.CoreV1().Namespaces().Create(ctx, nsSpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateTeamServiceAccount(ctx context.Context, namespace string) error {
	if c.dryRun {
		return nil
	}

	saSpec := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace,
			Namespace: namespace,
			Annotations: map[string]string{
				"iam.gke.io/gcp-service-account": fmt.Sprintf("%v@%v.iam.gserviceaccount.com", namespace, c.gcpProject),
			},
		},
	}

	_, err := c.clientSet.CoreV1().ServiceAccounts(namespace).Create(ctx, saSpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateCloudSQLProxy(ctx context.Context, name, team, dbInstance string) error {
	port := int32(5432)

	if c.dryRun {
		return nil
	}

	deploySpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: team,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: team,
					Containers: []v1.Container{
						{
							Name:  "cloudsql-proxy",
							Image: "gcr.io/cloudsql-docker/gce-proxy:1.29.0-alpine",
							Ports: []v1.ContainerPort{
								{
									Protocol:      v1.ProtocolTCP,
									ContainerPort: port,
								},
							},
							Command: []string{
								"/cloud_sql_proxy",
								"-term_timeout=30s",
								fmt.Sprintf("-instances=%v:%v:%v=tcp:0.0.0.0:%v", c.gcpProject, c.gcpRegion, dbInstance, port),
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientSet.AppsV1().Deployments(team).Create(ctx, deploySpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	serviceSpec := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: team,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []v1.ServicePort{
				{
					Protocol:   v1.ProtocolTCP,
					Port:       port,
					TargetPort: intstr.IntOrString{IntVal: port},
				},
			},
		},
	}

	_, err = c.clientSet.CoreV1().Services(team).Create(ctx, serviceSpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createConfig(inCluster bool) (*rest.Config, error) {
	if inCluster {
		return rest.InClusterConfig()
	}

	configPath := os.Getenv("KUBECONFIG")
	if configPath == "" {
		return nil, errors.New("KUBECONFIG env not set")
	}

	return clientcmd.BuildConfigFromFlags("", configPath)
}
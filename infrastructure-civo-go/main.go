package main

import (
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		firewall, err := civo.NewFirewall(ctx, "civo-firewall", &civo.FirewallArgs{
			CreateDefaultRules: pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}
		cluster, err := civo.NewKubernetesCluster(ctx, "civo-k3s-cluster", &civo.KubernetesClusterArgs{
			Pools: civo.KubernetesClusterPoolsArgs{
				Size:      pulumi.String("g4s.kube.medium"),
				NodeCount: pulumi.Int(2),
			},
			FirewallId: firewall.ID(),
		})
		if err != nil {
			return err
		}

		ctx.Export("ClusterName", cluster.Name)
		ctx.Export("ClusterId", cluster.ID())
		ctx.Export("kubeconfig", pulumi.ToSecret(cluster.Kubeconfig))

		k8sProvider, err := kubernetes.NewProvider(ctx, "kubernetes", &kubernetes.ProviderArgs{
			Kubeconfig:            cluster.Kubeconfig,
			EnableServerSideApply: pulumi.BoolPtr(true),
		}, pulumi.DependsOn([]pulumi.Resource{cluster}))
		if err != nil {
			return err
		}
		argo, err := helm.NewRelease(ctx, "argo", &helm.ReleaseArgs{
			Chart:           pulumi.String("argo-cd"),
			Version:         pulumi.String("5.42.1"),
			Namespace:       pulumi.String("argocd"),
			CreateNamespace: pulumi.BoolPtr(true),
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://argoproj.github.io/argo-helm"),
			},
			Values: pulumi.Map{
				"config": pulumi.Map{
					"params": pulumi.Map{
						"server.insecure": pulumi.BoolPtr(true),
					},
				},
			},
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return err
		}
		ctx.Export("ArgoNamespace", argo.Namespace)
		return nil
	})
}

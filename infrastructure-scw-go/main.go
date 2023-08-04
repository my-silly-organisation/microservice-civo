package main

import (
	"github.com/dirien/pulumi-scaleway/sdk/v2/go/scaleway"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		clusterConfig := config.New(ctx, "cluster")

		// Create a Scaleway resource (Object Bucket).
		cluster, err := scaleway.NewK8sCluster(ctx, "k8s-cluster", &scaleway.K8sClusterArgs{
			Version:                   pulumi.String(clusterConfig.Require("version")),
			Cni:                       pulumi.String("cilium"),
			DeleteAdditionalResources: pulumi.Bool(true),
			AutoUpgrade: scaleway.K8sClusterAutoUpgradeArgs{
				Enable:                     pulumi.Bool(clusterConfig.RequireBool("auto_upgrade")),
				MaintenanceWindowDay:       pulumi.String("monday"),
				MaintenanceWindowStartHour: pulumi.Int(2),
			},
		})
		if err != nil {
			return err
		}

		nodeConfig := config.New(ctx, "node")

		defaultNodePool, err := scaleway.NewK8sPool(ctx, "default-node-pool", &scaleway.K8sPoolArgs{
			ClusterId:   cluster.ID(),
			NodeType:    pulumi.String(nodeConfig.Require("node_type")),
			Size:        pulumi.Int(nodeConfig.RequireInt("node_count")),
			Autoscaling: pulumi.BoolPtr(nodeConfig.RequireBool("auto_scale")),
			Autohealing: pulumi.BoolPtr(nodeConfig.RequireBool("auto_heal")),
		})

		ctx.Export("ClusterName", cluster.Name)
		ctx.Export("ClusterId", cluster.ID())
		ctx.Export("NodePoolName", defaultNodePool.Name)
		// Export the name of the bucket.
		ctx.Export("kubeconfig", pulumi.ToSecret(cluster.Kubeconfigs.Index(pulumi.Int(0)).ConfigFile()))

		k8sProvider, err := kubernetes.NewProvider(ctx, "kubernetes", &kubernetes.ProviderArgs{
			Kubeconfig:            cluster.Kubeconfigs.Index(pulumi.Int(0)).ConfigFile(),
			EnableServerSideApply: pulumi.BoolPtr(true),
		}, pulumi.DependsOn([]pulumi.Resource{cluster, defaultNodePool}))
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

package main

import (
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
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
		return nil
	})
}

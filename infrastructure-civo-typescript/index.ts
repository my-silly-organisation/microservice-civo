import * as pulumi from "@pulumi/pulumi";
import * as civo from "@pulumi/civo";

const firewall = new civo.Firewall("civo-firewall", {
    createDefaultRules: true,
});

const cluster = new civo.KubernetesCluster("civo-k3s-cluster", {
    pools: {
        nodeCount: 2,
        size: "g4s.kube.medium"
    },
    firewallId: firewall.id,
})

export const ClusterName = cluster.name
export const ClusterId = cluster.id
export const kubeconfig = pulumi.secret(cluster.kubeconfig);

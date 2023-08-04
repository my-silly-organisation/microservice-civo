import * as pulumi from "@pulumi/pulumi";
import * as civo from "@pulumi/civo";
import * as k8s from "@pulumi/kubernetes";

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

const k8sProvider = new k8s.Provider("k8s", {
    kubeconfig: cluster.kubeconfig,
    enableServerSideApply: true,
}, {
    dependsOn: [cluster]
})

const argo = new k8s.helm.v3.Release("argo", {
    chart: "argo-cd",
    version: "5.42.1",
    repositoryOpts: {
        repo: "https://argoproj.github.io/argo-helm",
    },
    namespace: "argo",
    createNamespace: true,
    values: {
        configs: {
            params: {
                "server\.insecure": true,
            }
        }
    }
}, {
    provider: k8sProvider,
});

export const ClusterName = cluster.name
export const ClusterId = cluster.id
export const kubeconfig = pulumi.secret(cluster.kubeconfig)
export const argoNamespace = argo.namespace

import * as k8s from "@pulumi/kubernetes";
import * as scaleway from "@ediri/scaleway";
import * as pulumi from "@pulumi/pulumi";

const clusterConfig = new pulumi.Config("cluster")

const cluster = new scaleway.K8sCluster("k8s-cluster", {
    version: clusterConfig.require("version"),
    cni: "cilium",
    deleteAdditionalResources: true,
    tags: [
        "pulumi",
        "workshop",
    ],
    autoUpgrade: {
        enable: clusterConfig.requireBoolean("auto_upgrade"),
        maintenanceWindowStartHour: 3,
        maintenanceWindowDay: "monday"
    },
});

const nodeConfig = new pulumi.Config("node")

const defaultNodePool = new scaleway.K8sPool("default-node-pool", {
    nodeType: nodeConfig.require("node_type"),
    size: nodeConfig.requireNumber("node_count"),
    autoscaling: nodeConfig.requireBoolean("auto_scale"),
    autohealing: nodeConfig.requireBoolean("auto_heal"),
    clusterId: cluster.id,
});

const k8sProvider = new k8s.Provider("k8s", {
    kubeconfig: cluster.kubeconfigs[0].configFile,
    enableServerSideApply: true,
}, {
    dependsOn: [cluster, defaultNodePool]
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
export const kubeconfig = pulumi.secret(cluster.kubeconfigs[0].configFile)
export const argoNamespace = argo.namespace

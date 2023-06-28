import * as k8s from "@pulumi/kubernetes";
import * as pulumi from "@pulumi/pulumi";

const infraConfig = new pulumi.Config("infra");
const cfg = new pulumi.Config()
const infraStackReference = new pulumi.StackReference(infraConfig.require("stack-reference-name"));


const k8sProvider = new k8s.Provider("k8s", {
    kubeconfig: infraStackReference.getOutput("kubeconfig"),
    enableServerSideApply: true,
});

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

new k8s.apiextensions.CustomResource("microservice", {
    apiVersion: "argoproj.io/v1alpha1",
    kind: "Application",
    metadata: {
        name: cfg.require("application-name"),
        namespace: argo.namespace,
    },
    otherFields: {
        spec: {
            destination: {
                server: "https://kubernetes.default.svc",
                namespace: cfg.require("application-namespace"),
            },
            source: {
                path: cfg.require("application-repo-path"),
                repoURL: cfg.require("application-repo-url"),
                targetRevision: cfg.require("application-repo-target-revision"),
            },
            project: "default",
            syncPolicy: {
                automated: {
                    prune: true,
                    selfHeal: true,
                }
            },
        }
    }
}, {
    provider: k8sProvider,
    dependsOn: [argo],
})

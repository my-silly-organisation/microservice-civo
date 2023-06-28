import pulumi
import pulumi_civo as civo

firewall = civo.Firewall("civo-firewall",
                         create_default_rules=True)

cluster = civo.KubernetesCluster('civo-k3s-cluster',
                                 firewall_id=firewall.id,
                                 pools=civo.KubernetesClusterPoolsArgs(
                                     node_count=2,
                                     size="g4s.kube.medium",
                                 ))

pulumi.export('cluster_name', cluster.name)
pulumi.export('kubeconfig', pulumi.Output.secret(cluster.kubeconfig))

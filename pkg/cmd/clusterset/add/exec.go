// Copyright Contributors to the Open Cluster Management project
package add

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterclientset "open-cluster-management.io/api/client/cluster/clientset/versioned"
)

func (o *Options) complete(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		return fmt.Errorf("the name of the clusterset must be specified")
	}

	if len(args) > 1 {
		return fmt.Errorf("only one clusterset can be specified")
	}

	o.Clusterset = args[0]

	return nil
}

func (o *Options) validate() (err error) {
	if len(o.Clusters) == 0 {
		return fmt.Errorf("cluster name must be specified in --clusters")
	}
	return nil
}

func (o *Options) run() (err error) {
	restConfig, err := o.ClusteradmFlags.KubectlFactory.ToRESTConfig()
	if err != nil {
		return err
	}
	clusterClient, err := clusterclientset.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	for _, clusterName := range o.Clusters {
		cluster, err := clusterClient.ClusterV1().ManagedClusters().Get(context.TODO(), clusterName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if len(cluster.Labels) == 0 {
			cluster.Labels = map[string]string{}
		}

		if clusterset := cluster.Labels["cluster.open-cluster-management.io/clusterset"]; clusterset == o.Clusterset {
			fmt.Fprintf(o.Streams.Out, "Cluster %s is already added into Clusterset %s\n", clusterName, o.Clusterset)
			continue
		}

		cluster.Labels["cluster.open-cluster-management.io/clusterset"] = o.Clusterset
		_, err = clusterClient.ClusterV1().ManagedClusters().Update(context.TODO(), cluster, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		fmt.Fprintf(o.Streams.Out, "Cluster %s is added into Clusterset %s\n", clusterName, o.Clusterset)
	}

	return nil
}

package dbaasredhatcom

import (
	"strconv"

	"github.com/go-logr/logr"

	dbaasredhatcomv1alpha2 "github.com/CrunchyData/crunchy-bridge-operator/apis/dbaas.redhat.com/v1alpha2"
	"github.com/CrunchyData/crunchy-bridge-operator/internal/bridgeapi"
	dbaasv1alpha2 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha2"
)

const (
	TEAM_ID       = "team_id"
	PROVIDER_ID   = "provider_id"
	REGION_ID     = "region_id"
	CREATED_AT    = "created_at"
	UPDATED_AT    = "updated_at"
	MAJOR_VERSION = "major_version"
	STORAGE       = "storage"
	CPU           = "cpu"
	MEMORY        = "memory"
	IS_HA         = "is_ha"
	CLUSTER_NAME  = "name"
	STATE         = "state"
)

// discoverInventories query crunchy bridge and return list of inverntories by team
func (r *CrunchyBridgeInventoryReconciler) discoverInventories(inventory *dbaasredhatcomv1alpha2.CrunchyBridgeInventory, bridgeapi *bridgeapi.Client, logger logr.Logger) error {
	var bridgeInstances []dbaasv1alpha2.DatabaseService
	clusterList, clusterListErr := bridgeapi.ListAllClusters()
	if clusterListErr != nil {
		logger.Error(clusterListErr, "Error Listing the instance")
		return clusterListErr
	}
	logger.Info("cluster List ", " Total clusters ", len(clusterList.Clusters))
	if len(clusterList.Clusters) == 0 {
		logger.Info("cluster List ", " No Clusters found for account details ", inventory.Spec.CredentialsRef)
		inventory.Status.DatabaseServices = bridgeInstances
		return nil
	}
	for _, cluster := range clusterList.Clusters {
		clusterSvc := dbaasv1alpha2.DatabaseService{
			ServiceID:   cluster.ID,
			ServiceName: cluster.Name,
			ServiceType: dbaasv1alpha2.InstanceDatabaseService,
			ServiceInfo: map[string]string{
				TEAM_ID:       cluster.TeamID,
				PROVIDER_ID:   cluster.ProviderID,
				REGION_ID:     cluster.RegionID,
				CREATED_AT:    cluster.Created.String(),
				UPDATED_AT:    cluster.Updated.String(),
				MAJOR_VERSION: strconv.Itoa(cluster.PGMajorVersion),
				STORAGE:       strconv.Itoa(cluster.StorageGB),
				CPU:           strconv.Itoa(cluster.CPU),
				MEMORY:        strconv.Itoa(cluster.MemoryGB),
				IS_HA:         strconv.FormatBool(cluster.HighAvailability),
				STATE:         cluster.State,
			},
		}
		bridgeInstances = append(bridgeInstances, clusterSvc)
	}

	inventory.Status.DatabaseServices = bridgeInstances

	return nil
}

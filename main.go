package main

import (
	"fmt"

	"github.com/google/btree"
)

type KeyValue string
type KeyValueObj struct {
	key   string
	value string
}

func (kv KeyValue) Less(than btree.Item) bool {
	if kv < than.(KeyValue) {
		return true
	} else if kv < than.(KeyValue) {
		return false
	}
	return false

}

// type KeyValueItem struct {
// 	btree.Item
// }

var KeyValues = [...]KeyValueObj{
	{"/admin", "/admin_value"},
	{"/admin/delete_topics", "/admin/delete_topics_value"},
	{"/admin/preferred_replica_election", "/admin/preferred_replica_election_value"},
	{"/brokers", "/brokers_value"},
	{"/brokers/ids", "/brokers/ids_value"},
	{"/brokers/topics", "/brokers/topics_value"},
	{"/config", "/config_value"},
	{"/config/clients", "/config/clients_value"},
	{"/config/topics", "/config/topics_value"},
	{"/consumers", "/consumers_value"},
	{"/consumers/group", "/consumers/group_value"},
	{"/controller", "/controller_value"},
	{"/controller_epoch", "/controller_epoch_value"},
	{"/isr_change_notification", "/isr_change_notification_value"},
	{"/log_dir_event_notification", "/log_dir_event_notification_value"},
	{"/registry", "/registry_value"},
	{"/registry/componentstatuses", "/registry/componentstatuses_value"},
	{"/registry/configmaps", "/registry/configmaps_value"},
	{"/registry/cronjobs", "/registry/cronjobs_value"},
	{"/registry/daemonsets", "/registry/daemonsets_value"},
	{"/registry/deployments", "/registry/deployments_value"},
	{"/registry/endpoints", "/registry/endpoints_value"},
	{"/registry/endpointslices", "/registry/endpointslices_value"},
	{"/registry/events", "/registry/events_value"},
	{"/registry/ingresses", "/registry/ingresses_value"},
	{"/registry/jobs", "/registry/jobs_value"},
	{"/registry/leases", "/registry/leases_value"},
	{"/registry/limitranges", "/registry/limitranges_value"},
	{"/registry/namespaces", "/registry/namespaces_value"},
	{"/registry/networkpolicies", "/registry/networkpolicies_value"},
	{"/registry/node", "/registry/node_value"},
	{"/registry/persistentvolumeclaims", "/registry/persistentvolumeclaims_value"},
	{"/registry/persistentvolumes", "/registry/persistentvolumes_value"},
	{"/registry/podtemplates", "/registry/podtemplates_value"},
	{"/registry/pods", "/registry/pods_value"},
	{"/registry/replicasets", "/registry/replicasets_value"},
	{"/registry/replicationcontrollers", "/registry/replicationcontrollers_value"},
	{"/registry/resourcequotas", "/registry/resourcequotas_value"},
	{"/registry/runtimeclasses", "/registry/runtimeclasses_value"},
	{"/registry/secrets", "/registry/secrets_value"},
	{"/registry/services", "/registry/services_value"},
	{"/registry/statefulsets", "/registry/statefulsets_value"},
	{"/registry/volumeattachments", "/registry/volumeattachments_value"},
	{"garbage_key_1", "garbage_value_1"},
	{"garbage_key_2", "garbage_value_2"},
	{"garbage_key_3", "garbage_value_3"},
}

func main() {

	bt := btree.New(10)

	for i := range KeyValues {

		var x KeyValue
		x = KeyValue(KeyValues[i].key)
		bt.ReplaceOrInsert(x)
	}

	dels := make([]btree.Item, 0, bt.Len())
	bt.Ascend(func(item btree.Item) bool {
		if item.(KeyValue) < "garbage_key_1" {
			dels = append(dels, item)
		}

		return true
	})
	for i := range dels {
		fmt.Println(dels[i])
	}
	fmt.Println(len(dels), bt.Len())
}

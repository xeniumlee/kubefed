/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"context"
	"sync"
	"time"

	typesv1beta1 "github.com/xeniumlee/kubefed/apis/types/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type State int

const (
	ChanSize         = 10
	SyncPeriod       = time.Second * 30
	Init       State = iota
	Created
	Updated
)

var (
	workerMap  map[ctrlclient.ObjectKey]*SyncWorker = make(map[ctrlclient.ObjectKey]*SyncWorker)
	workerlock sync.RWMutex
)

type NotifyObj struct {
	obj         *typesv1beta1.FederatedObject
	clusterName string
}

type SyncWorker struct {
	notifyChan chan NotifyObj
	state      State
	objMap     map[string]*typesv1beta1.FederatedObject
	status     []typesv1beta1.ClusterStatus
	targets    int
}

func (w *SyncWorker) Run() {
	for {
		select {
		case msg := <-w.notifyChan:
			switch w.state {
			case Init:
				if msg.clusterName == FederationClusterName &&
					msg.obj.Status == nil {

					obj := msg.obj
					for _, cluster := range obj.Spec.Placement.Clusters {
						if client := GetclusterClient(cluster.Name); client != nil {
							if err := client.Create(context.TODO(), obj); err == nil {
								w.targets++
							}
						}
					}
					w.objMap[msg.clusterName] = obj
					w.status = make([]typesv1beta1.ClusterStatus, w.targets)
					w.state = Created
				}

			case Created:
				if msg.clusterName != FederationClusterName {
					// Created or Timestamp updated
					obj := msg.obj
					w.objMap[msg.clusterName] = obj
					if obj.Status != nil && len(obj.Status) == 1 {
						w.status = append(w.status, obj.Status[0])
						if len(w.status) == w.targets {

							// Update member cluster
							for cluster, o := range w.objMap {
								if client := GetclusterClient(cluster); client != nil {
									o.Status = w.status
									client.Update(context.TODO(), o)
								}
							}
							w.state = Updated
						}
					}
				}

			case Updated:
				return
			}
		case <-time.After(SyncPeriod):
			return
		}
	}
}

func (w *SyncWorker) Notify(clusterName string, obj *typesv1beta1.FederatedObject) {
	w.notifyChan <- NotifyObj{clusterName: clusterName, obj: obj}
}

func StartSync(clusterName string, key ctrlclient.ObjectKey, obj *typesv1beta1.FederatedObject) {
	w := &SyncWorker{
		notifyChan: make(chan NotifyObj, ChanSize),
		state:      Init,
		objMap:     make(map[string]*typesv1beta1.FederatedObject),
	}
	workerlock.Lock()
	workerMap[key] = w
	workerlock.Unlock()

	go w.Run()
	w.Notify(clusterName, obj)
}

func GetSyncworker(key ctrlclient.ObjectKey) *SyncWorker {
	workerlock.RLock()
	defer workerlock.RUnlock()
	if w, ok := workerMap[key]; ok {
		return w
	} else {
		return nil
	}
}

func RemoveSyncworker(key ctrlclient.ObjectKey) {
	workerlock.Lock()
	defer workerlock.Unlock()
	delete(workerMap, key)
}

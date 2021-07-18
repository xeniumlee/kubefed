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
	"fmt"
	"sync"
	"time"

	typesv1beta1 "github.com/xeniumlee/kubefed/apis/types/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type state int

const (
	chanSize         = 10
	syncPeriod       = time.Second * 40
	created    state = iota
	dispatched
	updated
)

var (
	workerMap  map[ctrlclient.ObjectKey]*syncWorker = make(map[ctrlclient.ObjectKey]*syncWorker)
	workerlock sync.RWMutex
)

type notifyObj struct {
	obj         *typesv1beta1.FederatedObject
	clusterName string
}

type syncWorker struct {
	notifyChan chan notifyObj
	objMap     map[string]*typesv1beta1.FederatedObject
	statusList []typesv1beta1.ClusterStatus
	key        ctrlclient.ObjectKey
	state      state
	dispatched int
}

func (w *syncWorker) run() {
	for {
		select {
		case msg := <-w.notifyChan:
			switch w.state {
			case created:
				if msg.clusterName == FederationClusterName &&
					msg.obj.Status == nil {

					obj := msg.obj.DeepCopy()
					w.dispatched = 0
					for _, cluster := range obj.Spec.Placement.Clusters {
						if client := GetclusterClient(cluster.Name); client != nil {
							obj.ObjectMeta.ResourceVersion = ""
							if err := client.Create(context.TODO(), obj); err == nil {
								w.dispatched++
							} else {
								fmt.Println(err)
							}
						}
					}
					w.objMap[msg.clusterName] = obj
					w.state = dispatched
					fmt.Println("Dispatched", w.dispatched)
				}

			case dispatched:
				if msg.clusterName != FederationClusterName &&
					msg.obj.Status != nil {

					obj := msg.obj.DeepCopy()
					w.objMap[msg.clusterName] = obj

					// Double check
					if obj.Status[0].Name == msg.clusterName {
						w.statusList = append(w.statusList, obj.Status[0])

						if len(w.statusList) == w.dispatched {

							for cluster, o := range w.objMap {
								if client := GetclusterClient(cluster); client != nil {
									o.Status = w.statusList
									fmt.Println(cluster, o)
									err := client.Status().Update(context.TODO(), o)
									fmt.Println(err)
								}
							}
							w.state = updated
						}
					}
				}

			case updated:
				fmt.Println("Got an update", msg.clusterName, msg.obj.Name)
			}
		case <-time.After(syncPeriod):
			workerlock.Lock()
			defer workerlock.Unlock()
			delete(workerMap, w.key)
			fmt.Println("Ending loop")
			return
		}
	}
}

func (w *syncWorker) notify(clusterName string, obj *typesv1beta1.FederatedObject) {
	w.notifyChan <- notifyObj{clusterName: clusterName, obj: obj}
}

func TryStartSync(clusterName string, key ctrlclient.ObjectKey, obj *typesv1beta1.FederatedObject) {
	workerlock.Lock()
	defer workerlock.Unlock()
	if _, ok := workerMap[key]; ok {
		fmt.Println("Already Started")
		return
	}

	w := &syncWorker{
		notifyChan: make(chan notifyObj, chanSize),
		objMap:     make(map[string]*typesv1beta1.FederatedObject),
		state:      created,
		key:        key,
	}

	workerMap[key] = w
	go w.run()
	w.notify(clusterName, obj)
	fmt.Println("TryStartSync")
}

func TryNotify(clusterName string, key ctrlclient.ObjectKey, obj *typesv1beta1.FederatedObject) {
	workerlock.RLock()
	defer workerlock.RUnlock()
	if w, ok := workerMap[key]; ok {
		w.notify(clusterName, obj)
		fmt.Println("TryNotify")
	}
}

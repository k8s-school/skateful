/*
Copyright 2016 The Kubernetes Authors.
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

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/scheme"

	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	clientCfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		log.Fatal(err.Error())
	}
	namespace := clientCfg.Contexts[clientCfg.CurrentContext].Namespace

	if namespace == "" {
		namespace = "default"
	}

	//log.Printf("Namespace %s", namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/managed-by=qserv-operator,app=qserv",
	}

	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), listOptions)
	//log.Printf("PVCs %v\n", pvcs)
	if err != nil {
		log.Fatal(err.Error())
	}

	var curatedPvcs v1.PersistentVolumeClaimList
	curatedPvcs.TypeMeta = metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "List",
	}
	var curatedPvs v1.PersistentVolumeList
	curatedPvs.TypeMeta = metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "List",
	}
	for _, pvc := range pvcs.Items {
		//log.Printf("PVC %v\n", pvc)
		curatedPvc := v1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "PersistentVolumeClaim",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: pvc.ObjectMeta.GetName(),
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes:      pvc.Spec.AccessModes,
				Resources:        pvc.Spec.Resources,
				StorageClassName: pvc.Spec.StorageClassName,
				VolumeName:       pvc.Spec.VolumeName,
			},
		}
		curatedPvcs.Items = append(curatedPvcs.Items, curatedPvc)
		//log.Printf("PVCs %v\n", curatedPvcs)
		pv, err := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pvc.Spec.VolumeName, metav1.GetOptions{})
		if err != nil {
			log.Fatal(err.Error())
		}
		curatedPv := v1.PersistentVolume{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "PersistentVolume",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: pv.ObjectMeta.GetName(),
			},
			Spec: v1.PersistentVolumeSpec{
				AccessModes:                   pv.Spec.AccessModes,
				Capacity:                      pv.Spec.Capacity,
				NodeAffinity:                  pv.Spec.NodeAffinity,
				PersistentVolumeReclaimPolicy: pv.Spec.PersistentVolumeReclaimPolicy,
				PersistentVolumeSource:        pv.Spec.PersistentVolumeSource,
				StorageClassName:              pv.Spec.StorageClassName,
			},
		}
		curatedPvs.Items = append(curatedPvs.Items, curatedPv)
		//log.Printf("PV %v\n", curatedPv)
	}

	// Examples for error handling:
	// - Use helper functions like e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	// pod := "example-xxxxx"
	// _, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
	// if errors.IsNotFound(err) {
	// 	fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
	// } else if statusError, isStatus := err.(*errors.StatusError); isStatus {
	// 	fmt.Printf("Error getting pod %s in namespace %s: %v\n",
	// 		pod, namespace, statusError.ErrStatus.Message)
	// } else if err != nil {
	// 	panic(err.Error())
	// } else {
	// 	fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	// }

	err = s.Encode(&curatedPvcs, os.Stdout)
	if err != nil {
		panic(err)
	}
	fmt.Println("---")
	err = s.Encode(&curatedPvs, os.Stdout)
	if err != nil {
		panic(err)
	}
}

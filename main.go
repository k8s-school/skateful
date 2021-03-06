package main

import (
	"context"
	"flag"
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	const (
		csi   = "csi"
		usage = "the variety of gopher"
	)

	storage := flag.String("storage", csi, "Type of storage: 'csi' or 'local'")
	outFile := flag.String("out", "pvc-pv.yaml", "Path to output file")
	flag.Parse()

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	switch *storage {
	case csi:

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

		file, err := os.Create(*outFile) // For read access.
		if err != nil {
			log.Fatal(err)
		}
		err = s.Encode(&curatedPvcs, file)
		if err != nil {
			panic(err)
		}
		if _, err = file.WriteString("---\n"); err != nil {
			file.Close()
			check(err)
		}
		err = s.Encode(&curatedPvs, file)
		if err != nil {
			panic(err)
		}
		file.Close()
	default:
		log.Fatalf("Unknow storage %s", *storage)
	}
}

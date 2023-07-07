/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get curated yaml for pvc and pv",
	Long: `Extract PVCs and PVs from a Qserv deployment

Create a YAML file which can be used to recreate the same PVCs and PVs`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called")
		getYaml()
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

const (
	csi = "csi"
)

func getYaml() {

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	storage := viper.GetString("storage")
	outFile := viper.GetString("out")

	switch storage {
	case csi:
		logger.Debug("CSI storage")
		clientset, _ := setKubeClient()

		clientCfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
		if err != nil {
			log.Fatal(err.Error())
		}
		namespace := clientCfg.Contexts[clientCfg.CurrentContext].Namespace

		if namespace == "" {
			namespace = "default"
		}

		logger.Debugf("Namespace %s", namespace)

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
			logger.Debug("PVCs %v\n", curatedPvcs)
			pv, err := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pvc.Spec.VolumeName, metav1.GetOptions{})
			if err != nil {
				log.Fatal(err.Error())
			}
			claimRef := pv.Spec.ClaimRef
			claimRef.UID = ""
			claimRef.ResourceVersion = ""
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
					ClaimRef:                      claimRef,
					NodeAffinity:                  pv.Spec.NodeAffinity,
					PersistentVolumeReclaimPolicy: pv.Spec.PersistentVolumeReclaimPolicy,
					PersistentVolumeSource:        pv.Spec.PersistentVolumeSource,
					StorageClassName:              pv.Spec.StorageClassName,
				},
			}
			curatedPvs.Items = append(curatedPvs.Items, curatedPv)
			//log.Printf("PV %v\n", curatedPv)
		}

		logger.Debugf("Create yaml file %s", outFile)
		file, err := os.Create(outFile) // For read access.
		if err != nil {
			log.Fatal(err)
		}

		logger.Debug("Write PVC")
		err = s.Encode(&curatedPvcs, file)
		if err != nil {
			panic(err)
		}
		if _, err = file.WriteString("---\n"); err != nil {
			file.Close()
			check(err)
		}

		logger.Debug("Write PV")
		err = s.Encode(&curatedPvs, file)
		if err != nil {
			panic(err)
		}
		file.Close()
	default:
		log.Fatalf("Unknow storage %s", storage)
	}
}

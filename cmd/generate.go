/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate curated yaml for pvc and pv",
	Long:  `Use a time range to retrieve a set of PV and generate curated yaml for these PVs and their Qserv PVCs`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generate called")
		generateYaml()
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func generateYaml() {

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	storage := viper.GetString("storage")
	outFile := viper.GetString("out")

	start, _ := time.Parse(time.RFC3339, "2021-08-19T18:00:14Z")
	end, _ := time.Parse(time.RFC3339, "2021-08-19T19:00:14Z")

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

		listOptions := metav1.ListOptions{}

		pvs, err := clientset.CoreV1().PersistentVolumes().List(context.TODO(), listOptions)
		//log.Printf("PVCs %v\n", pvcs)
		if err != nil {
			log.Fatal(err.Error())
		}

		var seekedPvs v1.PersistentVolumeList
		seekedPvs.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		}
		for _, pv := range pvs.Items {
			// Check that pv creation date is not older than 1 day
			creationTimestamp := pv.GetObjectMeta().GetCreationTimestamp()
			if inTimeSpan(start, end, creationTimestamp.Time) {
				logger.Debugf("PV: %s:\tclaim: %s,\tcreated at: %s\n", pv.GetObjectMeta().GetName(), pv.Spec.ClaimRef.Name, creationTimestamp.Time)
				logger.Debugf("PV: %s", &pv.TypeMeta)

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
				seekedPvs.Items = append(seekedPvs.Items, curatedPv)
			}
		}

		var curatedPvcs v1.PersistentVolumeClaimList
		curatedPvcs.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		}
		for _, pv := range seekedPvs.Items {
			//log.Printf("PVC %v\n", pvc)
			storageClassName := pv.Spec.StorageClassName
			curatedPvc := v1.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "PersistentVolumeClaim",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      pv.Spec.ClaimRef.Name,
					Namespace: pv.Spec.ClaimRef.Namespace,
				},
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes: pv.Spec.AccessModes,
					Resources: v1.ResourceRequirements{
						Requests: pv.Spec.Capacity,
					},
					StorageClassName: &storageClassName,
					VolumeName:       pv.GetObjectMeta().GetName(),
				},
			}
			curatedPvcs.Items = append(curatedPvcs.Items, curatedPvc)
			logger.Debug("PVCs %v\n", curatedPvcs)
			//log.Printf("PV %v\n", curatedPv)
		}

		logger.Debugf("Create yaml file %s", outFile)
		file, err := os.Create(outFile) // For read access.
		if err != nil {
			log.Fatal(err)
		}

		logger.Debug("Write PVs")
		err = s.Encode(&seekedPvs, file)
		if err != nil {
			panic(err)
		}
		if _, err = file.WriteString("---\n"); err != nil {
			file.Close()
			check(err)
		}

		logger.Debug("Write PVCs")
		err = s.Encode(&curatedPvcs, file)
		if err != nil {
			panic(err)
		}
		file.Close()
	default:
		log.Fatalf("Unknow storage %s", storage)
	}
}

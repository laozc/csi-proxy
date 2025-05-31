package main

import (
	"fmt"

	"github.com/kubernetes-csi/csi-proxy/pkg/os/system"
)

func main() {
	/*
		disks, err := cim.ListDisks(nil)

		if err != nil {
			fmt.Println(err)
		}

		//cim.QueryInstances(cim.WMINamespaceStorage,
		//	query.NewWmiQuery("ASSOCIATORS OF {MSFT_Disk.Manufacturer = 'VMware'} WHERE AssocClass=MSFT_DiskToPartition"))

		for i := range disks {
			path := disks[i].RelativePath()
			//objectId, _ := disks[i].GetPropertyObjectId()

			fmt.Println(path)
			//fmt.Println(disks[i].GetAllRelatedWithQuery(query.NewWmiQuery("MSFT_Partition", "DiskId", objectId)))
			coll, err := disks[i].GetAssociated("MSFT_DiskToPartition", "MSFT_Partition", "", "")
			if err != nil {
				fmt.Println(err)
			}

			fmt.Printf("Disk %d Found  %d parts\n", i, len(coll))
			for j := range coll {
				partID, _ := coll[j].GetProperty("ObjectId")
				fmt.Println(coll[j])
				r, err := coll[j].GetAssociated("MSFT_PartitionToVolume", "MSFT_Volume", "", "")
				if err != nil {
					fmt.Println(err)
				}

				fmt.Printf("Found  %d volumes\n", len(r))
				fmt.Println(r)

				if len(r) > 0 {
					vol := r[0]
					part, err := vol.GetAssociated("MSFT_PartitionToVolume", "MSFT_Partition", "Partition", "Volume")
					if err != nil {
						fmt.Println(err)
					}

					fmt.Println(part)
					inst := part[0]
					partID2, _ := inst.GetProperty("ObjectId")
					if partID == partID2 {
						fmt.Println("You've got a match")
					} else {
						fmt.Println("You've got an non- match")
					}
				}
			}
		}*/
	name := "csiproxy"

	api := system.APIImplementor{}

	err := api.StartService(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Note: both StartService and StopService are not implemented by WMI
	//script := `Start-Service -Name $env:ServiceName`
	//cmdEnv := fmt.Sprintf("ServiceName=%s", name)
	//out, err := utils.RunPowershellCmd(script, cmdEnv)
	//if err != nil {
	//	fmt.Println(err)
	//}

	//_ = out

	err = api.StopService("containerd", true)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

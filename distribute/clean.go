package distribute

import (
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/jinzhu/copier"
	"github.com/panjf2000/ants"
)

func ClearAllInstances(opt libs.Options) {
	if opt.Cloud.Retry == 0 {
		opt.Cloud.Retry = 8
	}

	// select all cloud instance
	instances := GetAllInstances(opt)
	if len(instances) == 0 {
		utils.WarnF("no active cloud instance running")
		return
	}

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(opt.Concurrency*5, func(i interface{}) {
		instance := i.(CloudRunner)

		var options libs.Options
		copier.Copy(&options, &opt)
		instance.Opt = opt
		instance.Provider.IsBackgroundCheck = true
		instance.Provider.InitClient()
		instance.Prepare()
		utils.InforF("Deleting the instance: %v -- %v", instance.Provider.ProviderName, instance.PublicIP)
		if err := instance.Provider.DeleteInstance(instance.InstanceID); err == nil {
			utils.InforF("Instance deleted %s -- %s", color.HiYellowString(instance.PublicIP), color.HiYellowString(instance.InstanceID))
			instanceFile := instance.Opt.Env.InstancesFolder + "/" + instance.InstanceName + "-" + instance.PublicIP + ".json"
			os.Remove(instanceFile)
		}

		wg.Done()
	}, ants.WithPreAlloc(true))
	defer p.Release()

	utils.InforF("Evaluating the status of %v cloud instances", color.HiMagentaString("%v", len(instances)))
	for _, instance := range instances {
		wg.Add(1)
		_ = p.Invoke(instance)
	}
	wg.Wait()

}

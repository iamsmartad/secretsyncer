package main

import (
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"

	log "github.com/sirupsen/logrus"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type ConfigController struct {
	informerFactory informers.SharedInformerFactory
	configInformer  coreinformers.ConfigMapInformer
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *ConfigController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.configInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}

func (c *ConfigController) add(obj interface{}) {
	config := obj.(*v1.ConfigMap)
	log.Infof("[on local cluster]: watching configmap %s/%s", config.Namespace, config.Name)
}

func (c *ConfigController) update(old, new interface{}) {
	oldconfig := old.(*v1.ConfigMap)
	config := new.(*v1.ConfigMap)
	log.Infof("[on local cluster]: configmap %s/%s UPDATED (verion %v -> %v)", config.Name, config.Namespace, oldconfig.ResourceVersion, config.ResourceVersion)
	log.Fatalf("rebooting pod to reload config changes... good bye...")
	os.Exit(0)
}

func (c *ConfigController) delete(obj interface{}) {
	config := obj.(*v1.ConfigMap)
	log.Infof("[on local cluster]: config %s/%s DELETED", config.Namespace, config.Name)
}

// NewConfigController creates a NewSecretController
func NewConfigController(informerFactory informers.SharedInformerFactory) *ConfigController {
	configInformer := informerFactory.Core().V1().ConfigMaps()

	c := &ConfigController{
		informerFactory: informerFactory,
		configInformer:  configInformer,
	}
	configInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.add,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.update,
			// Called on resource deletion.
			DeleteFunc: c.delete,
		},
	)
	return c
}

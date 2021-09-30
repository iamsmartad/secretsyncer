package pkg

import (
	"fmt"
	"time"

	"github.com/iamsmartad/secretsyncer/helpers"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"

	log "github.com/sirupsen/logrus"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// SecretController ...
type SecretController struct {
	informerFactory informers.SharedInformerFactory
	secretInformer  coreinformers.SecretInformer
	syncrule        helpers.SyncRule
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *SecretController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.secretInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}

func (c *SecretController) secretAdd(obj interface{}) {
	secret := obj.(*v1.Secret)
	if helpers.FindMatchingAnnotation(secret.GetAnnotations(), c.syncrule.Annotations) {
		log.Infof("[on remote cluster]: watching secret %s/%s", secret.Namespace, secret.Name)
		sec := c.createApplyableSecret(secret)
		if sec != nil {
			updateLocalSecret(sec)
		}
	}

}

func (c *SecretController) secretUpdate(old, new interface{}) {
	oldsecret := old.(*v1.Secret)
	secret := new.(*v1.Secret)
	if oldsecret.ResourceVersion != secret.ResourceVersion && helpers.FindMatchingAnnotation(secret.GetAnnotations(), c.syncrule.Annotations) {
		log.Infof("[on remote cluster]: secret %s/%s UPDATED (verion %v -> %v)", secret.Namespace, secret.Name, oldsecret.ResourceVersion, secret.ResourceVersion)
	}
	sec := c.createApplyableSecret(secret)
	if sec != nil {
		updateLocalSecret(sec)
	}
}

func (c *SecretController) secretDelete(obj interface{}) {
	secret := obj.(*v1.Secret)
	log.Infof("[on remote cluster]: secret %s/%s DELETED", secret.Namespace, secret.Name)
}

func (c *SecretController) localSecretAdd(obj interface{}) {
	secret := obj.(*v1.Secret)
	log.Infof("[on local cluster]: watching secret %s/%s", secret.Namespace, secret.Name)
	if secret.Labels[SYNCLABEL] == "receiver" {
		updateOnReceiver(secret)
		return
	}
	triggerSyncOnReceivers(secret)
}

func (c *SecretController) localSecretUpdate(old, new interface{}) {
	oldsecret := old.(*v1.Secret)
	secret := new.(*v1.Secret)
	log.Infof("[on local cluster]: secret %s/%s UPDATED (verion %v -> %v)", secret.Name, secret.Namespace, oldsecret.ResourceVersion, secret.ResourceVersion)
	if secret.Labels[SYNCLABEL] == "receiver" {
		updateOnReceiver(secret)
		return
	}
	triggerSyncOnReceivers(secret)
}

func (c *SecretController) localSecretDelete(obj interface{}) {
	secret := obj.(*v1.Secret)
	log.Infof("[on local cluster]: secret %s/%s DELETED", secret.Namespace, secret.Name)
}

func (c *SecretController) createApplyableSecret(secret *v1.Secret) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: *TRUTHNAMESPACE,
			Annotations: map[string]string{
				SYNCSOURCENAMESPACE:           secret.Namespace,
				SYNCSOURCENAME:                secret.Name,
				SYNCRESOURCEVERSIONKEY:        secret.ResourceVersion,
				"field.cattle.io/description": fmt.Sprintf("synced %s from REMOTE %s/%s (UTC+2)", timestamp(), secret.Namespace, secret.Name),
			},
			Labels: map[string]string{SYNCLABEL: "true"},
		},
		Data:       secret.Data,
		StringData: secret.StringData,
		Type:       secret.Type,
	}
}

func updateOnReceiver(secret *v1.Secret) {

	sourceSecret, err := getLocalSecret(secret.GetAnnotations()[SYNCSOURCENAME], secret.GetAnnotations()[SYNCSOURCENAMESPACE])
	if err != nil {
		log.Warnf(err.Error())
		return
	}
	if sourceSecret.ResourceVersion == secret.GetAnnotations()[SYNCRESOURCEVERSIONKEY] {
		log.Debugf("Secret %s/%s (%s) is up-to-date with secret %s/%s (%s)",
			secret.Name,
			secret.Namespace,
			secret.GetAnnotations()[SYNCRESOURCEVERSIONKEY],
			secret.GetAnnotations()[SYNCSOURCENAME],
			secret.GetAnnotations()[SYNCSOURCENAMESPACE],
			sourceSecret.ResourceVersion)
		return
	}
	log.Infof("Secret %s/%s (%s) has to be updated to version  %s ", secret.Name, secret.Namespace, secret.GetAnnotations()[SYNCRESOURCEVERSIONKEY], sourceSecret.ResourceVersion)
	secret.Data = sourceSecret.Data
	secret.StringData = sourceSecret.StringData
	secret.Annotations[SYNCRESOURCEVERSIONKEY] = sourceSecret.ResourceVersion
	secret.Annotations["field.cattle.io/description"] = "synced " + timestamp() + " from " + sourceSecret.Namespace + "/" + sourceSecret.Name
	updateLocalSecret(secret)
}

func triggerSyncOnReceivers(secret *v1.Secret) {
	localSecrets, err := getAllLocalSecrets("", SYNCLABEL+"=receiver")
	if err != nil {
		log.Warnf(err.Error())
		return
	}
	for _, sec := range localSecrets {
		log.Debugf("Checking %s", sec.Name)
		name, ok := sec.GetAnnotations()[SYNCSOURCENAME]
		if !ok {
			continue
		}
		namespace, ok := sec.GetAnnotations()[SYNCSOURCENAMESPACE]
		if !ok {
			namespace = sec.Namespace
		}
		if name == secret.Name && namespace == secret.Namespace {
			updateOnReceiver(&sec)
		}
	}
}

func timestamp() string {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return time.Now().Format("Jan 02 15:04:05")
	}
	return time.Now().In(loc).Format("Jan 02 15:04:05")

}

// NewSecretController creates a NewSecretController
func NewSecretController(informerFactory informers.SharedInformerFactory, syncrule helpers.SyncRule, local bool) *SecretController {
	secretInformer := informerFactory.Core().V1().Secrets()

	c := &SecretController{
		informerFactory: informerFactory,
		secretInformer:  secretInformer,
		syncrule:        syncrule,
	}
	if local {
		secretInformer.Informer().AddEventHandler(
			// Your custom resource event handlers.
			cache.ResourceEventHandlerFuncs{
				// Called on creation
				AddFunc: c.localSecretAdd,
				// Called on resource update and every resyncPeriod on existing resources.
				UpdateFunc: c.localSecretUpdate,
				// Called on resource deletion.
				DeleteFunc: c.localSecretDelete,
			},
		)
		return c
	}
	secretInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.secretAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.secretUpdate,
			// Called on resource deletion.
			DeleteFunc: c.secretDelete,
		},
	)
	return c
}

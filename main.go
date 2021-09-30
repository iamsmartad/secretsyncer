package main

import (
	"context"
	"flag"
	"path/filepath"
	"time"

	"github.com/iamsmartad/secretsyncer/helpers"
	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig         *string
	syncrulespath      *string
	syncintervalRemote *int64
	syncintervalLocal  *int64
	basesynclabel      *string
	// TRUTHNAMESPACE ...
	TRUTHNAMESPACE *string
	// SYNCLABEL ...
	SYNCLABEL string
	// SYNCRESOURCEVERSIONKEY ...
	SYNCRESOURCEVERSIONKEY string
	// SYNCSOURCENAME ...
	SYNCSOURCENAME string
	// SYNCSOURCENAMESPACE ...
	SYNCSOURCENAMESPACE string
)

func main() {
	kubeconfig = flag.String("kubeconfig", filepath.Join("/root", ".kube", "config"), "absolute path to the kubeconfig file")
	syncrulespath = flag.String("syncrulespath", filepath.Join("/root", "syncrules"), "absolute path to the syncrules file")
	syncintervalRemote = flag.Int64("syncinterval-remote", 24*60, "forced sync with remote cluster [in minutes]")
	syncintervalLocal = flag.Int64("syncinterval-local", 5, "forced sync of secrets on local cluster [in minutes]")
	TRUTHNAMESPACE = flag.String("truth-namespace", "truth", "default namespace in remote an local cluster to sync between")
	basesynclabel = flag.String("labelbase", "iamstudent.dev/sync", "default base name for all related labels and annotations")
	flag.Parse()

	SYNCLABEL = *basesynclabel
	SYNCRESOURCEVERSIONKEY = *basesynclabel + "ResourceVersion"
	SYNCSOURCENAME = *basesynclabel + "SourceName"
	SYNCSOURCENAMESPACE = *basesynclabel + "SourceNamespace"

	log.SetFormatter(&log.TextFormatter{FullTimestamp: false, TimestampFormat: "15:04:05"})

	syncrules := helpers.GetSyncRules(*syncrulespath)

	for syncRuleName, sr := range syncrules {
		log.Infof("handling syncrule `%s`", syncRuleName)

		switch d := sr.Direction; d {

		case "fromPrimary":
			for _, ns := range sr.Namespaces {
				log.Infof("[on remote cluster]: watching namespace `%s`", ns)
				defer close(watchRemoteSecrets(ns, sr))
			}

		case "local":
			log.Infof("[on local cluster]: watching all namespaces")
			defer close(watchLocalSecrets(sr))

		default:
			log.Infof("unknown syncrule direction `%s`", d)
		}
	}

	defer close(watchOwnConfigsMap())

	select {}
}

func watchRemoteSecrets(namespace string, syncrule helpers.SyncRule) chan struct{} {
	namespaceOptions := informers.WithNamespace(namespace)
	labelOptions := informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = SYNCLABEL + "=true"
	})
	factory := informers.NewSharedInformerFactoryWithOptions(getOutOfClusterClientset(),
		time.Minute*time.Duration(*syncintervalRemote),
		namespaceOptions,
		labelOptions)
	controller := NewSecretController(factory, syncrule, false)
	stop := make(chan struct{})
	err := controller.Run(stop)
	if err != nil {
		log.Fatal(err)
	}
	return stop
}

func watchLocalSecrets(syncrule helpers.SyncRule) chan struct{} {
	labelOptions := informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = SYNCLABEL
	})
	factory := informers.NewSharedInformerFactoryWithOptions(getInClusterClientset(), time.Minute*time.Duration(*syncintervalLocal), labelOptions)
	controller := NewSecretController(factory, syncrule, true)
	stop := make(chan struct{})
	err := controller.Run(stop)
	if err != nil {
		log.Fatal(err)
	}

	return stop
}

func watchOwnConfigsMap() chan struct{} {
	labelOptions := informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = SYNCLABEL + "=config"
	})
	factory := informers.NewSharedInformerFactoryWithOptions(getInClusterClientset(),
		time.Hour*24,
		labelOptions)
	controller := NewConfigController(factory)
	stop := make(chan struct{})
	err := controller.Run(stop)
	if err != nil {
		log.Fatal(err)
	}
	return stop
}

func getLocalSecret(name string, namespace string) (*v1.Secret, error) {
	clientset := getInClusterClientset()
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func getAllLocalSecrets(namespace string, labels string) ([]v1.Secret, error) {
	clientset := getInClusterClientset()
	secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labels})
	if err != nil {
		return nil, err
	}
	return secrets.Items, nil
}

func updateLocalSecret(secret *v1.Secret) error {
	clientset := getInClusterClientset()
	// first check if secret already exists with resourceversion
	existing, _ := clientset.CoreV1().Secrets(secret.Namespace).Get(context.TODO(), secret.Name, metav1.GetOptions{})
	if existing != nil && existing.GetAnnotations()[SYNCRESOURCEVERSIONKEY] == secret.Annotations[SYNCRESOURCEVERSIONKEY] {
		return nil
	}

	log.Infof("updating... secret %s to resVer: %v", secret.Name, secret.Annotations[SYNCRESOURCEVERSIONKEY])
	_, err := clientset.CoreV1().Secrets(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		log.Infof(err.Error())
		log.Infof("creating new secret now")
		_, err = clientset.CoreV1().Secrets(secret.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			log.Warnf(err.Error())
		}
	}

	return nil
}

func getOutOfClusterClientset() *kubernetes.Clientset {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func getInClusterClientset() *kubernetes.Clientset {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

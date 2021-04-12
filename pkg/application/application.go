package application

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/argoproj/argocd-autopilot/pkg/log"
	"github.com/argoproj/argocd-autopilot/pkg/store"
	"github.com/argoproj/argocd-autopilot/pkg/util"
	"github.com/ghodss/yaml"

	argocdutil "github.com/argoproj/argo-cd/v2/cmd/util"
	argocdapp "github.com/argoproj/argo-cd/v2/pkg/apis/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argocdsettings "github.com/argoproj/argo-cd/v2/util/settings"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	kusttypes "sigs.k8s.io/kustomize/api/types"
)

const (
	defaultDestServer = "https://kubernetes.default.svc"
)

var (
	// Errors
	ErrEmptyAppSpecifier = errors.New("empty app specifier not allowed")
	ErrEmptyAppName      = errors.New("app name cannot be empty, please specify application name with: --app-name")
)

type Application interface {
	// Base returns the base kustomization file for this app.
	Base() *kusttypes.Kustomization

	// Overlay returns the overlay kustomization object that is looking on this
	// app.Base() file.
	Overlay() *kusttypes.Kustomization

	// Config returns this app's config.json file that should be next to the overlay
	// kustomization.yaml file. This is used by the environment's application set
	// to generate the final argo-cd application.
	ConfigJson() *Config
}

type BootstrapApplication interface {
	// GenerateManifests runs kustomize build on the app and returns the result.
	GenerateManifests() ([]byte, error)

	// ArgoCD parses the app flags and returns the constructed argo-cd application.
	ArgoCD() *v1alpha1.Application

	// Kustomization returns the kustomization for the bootstrap application.
	//  only available when creating bootstrap application.
	Kustomization() (*kusttypes.Kustomization, error)

	// RootApp returns the root application that watches the envs/ directory
	// to create new environments application-sets.
	RootApp(revision, srcPath string) *v1alpha1.Application
}

type Config struct {
	AppName       string `json:"appName,omitempty"`
	UserGivenName string `json:"userGivenName,omitempty"`
	DestNamespace string `json:"destNamespace,omitempty"`
	DestServer    string `json:"destServer,omitempty"`
}

type CreateOptions struct {
	AppSpecifier   string
	AppName        string
	SrcPath        string
	Namespace      string
	Server         string
	argoAppOptions argocdutil.AppOptions
	flags          *pflag.FlagSet
}

type application struct {
	tag       string
	name      string
	namespace string
	path      string
	fs        filesys.FileSystem
	argoApp   *v1alpha1.Application
}

type bootstrapApp struct {
	*application
	repoUrl string
}

func AddFlags(cmd *cobra.Command, defAppName string) *CreateOptions {
	co := &CreateOptions{}

	cmd.Flags().StringVar(&co.AppSpecifier, "app", "", "The application specifier (e.g. argocd@v1.0.2 | https://github.com")
	cmd.Flags().StringVar(&co.AppName, "app-name", defAppName, "The application name")
	cmd.Flags().StringVar(&co.Server, "dest-server", "", "K8s cluster URL (e.g. https://kubernetes.default.svc)")
	cmd.Flags().StringVar(&co.Namespace, "dest-namespace", "", "K8s target namespace (overrides the namespace specified in the ksonnet app.yaml)")

	co.flags = cmd.Flags()

	return co
}

/*********************************/
/*       CreateOptions impl      */
/*********************************/
func (o *CreateOptions) Parse() (Application, error) {
	return parseApplication(o)
}

func (o *CreateOptions) ParseBootstrap() (BootstrapApplication, error) {
	app, err := parseApplication(o)
	if err != nil {
		return nil, err
	}
	app.argoApp.ObjectMeta.Namespace = app.namespace
	app.argoApp.Spec.Destination.Server = defaultDestServer
	app.argoApp.Spec.Destination.Namespace = app.namespace
	app.argoApp.Spec.Source.Path = o.SrcPath

	return &bootstrapApp{
		application: app,
		repoUrl:     o.flags.Lookup("repo").Value.String(),
	}, nil
}

func parseApplication(o *CreateOptions) (*application, error) {
	if o.AppSpecifier == "" {
		return nil, ErrEmptyAppSpecifier
	}

	if o.AppName == "" {
		return nil, ErrEmptyAppName
	}

	argoApp, err := argocdutil.ConstructApp("", o.AppName, getLabels(o.AppName), []string{}, o.argoAppOptions, o.flags)
	if err != nil {
		return nil, err
	}

	// set default options
	argoApp.Spec.SyncPolicy = &v1alpha1.SyncPolicy{
		Automated: &v1alpha1.SyncPolicyAutomated{
			SelfHeal: true,
			Prune:    true,
		},
	}

	app := &application{
		name:      o.AppName,
		path:      o.AppSpecifier, // TODO: supporting only path for now
		namespace: o.Namespace,
		fs:        filesys.MakeFsOnDisk(),
		argoApp:   argoApp,
	}

	return app, nil
}

/*********************************/
/*        Application impl       */
/*********************************/
func (app *application) Base() *kusttypes.Kustomization {
	return &kusttypes.Kustomization{
		Resources: []string{app.path},
		TypeMeta: kusttypes.TypeMeta{
			APIVersion: kusttypes.KustomizationVersion,
			Kind:       kusttypes.KustomizationKind,
		},
	}
}

func (app *application) Overlay() *kusttypes.Kustomization {
	return &kusttypes.Kustomization{
		Resources: []string{"../../base"},
		TypeMeta: kusttypes.TypeMeta{
			APIVersion: kusttypes.KustomizationVersion,
			Kind:       kusttypes.KustomizationKind,
		},
	}
}

func (app *application) ConfigJson() *Config {
	return &Config{
		AppName:       app.argoApp.Name,
		UserGivenName: app.argoApp.Name,
		DestNamespace: app.argoApp.Spec.Destination.Namespace,
		DestServer:    app.argoApp.Spec.Destination.Server,
	}
}

func (app *application) kustomizeBuild() ([]byte, error) {
	kopts := krusty.MakeDefaultOptions()
	kopts.DoLegacyResourceSort = true

	k := krusty.MakeKustomizer(kopts)

	log.G().WithField("path", app.path).Debug("running kustomize")
	res, err := k.Run(app.fs, app.path)
	if err != nil {
		return nil, err
	}

	return res.AsYaml()
}

/*********************************/
/*   Bootstrap application impl  */
/*********************************/

func (app *bootstrapApp) GenerateManifests() ([]byte, error) {
	td, err := ioutil.TempDir("", "auto-pilot")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(td)
	kustPath := filepath.Join(td, "kustomization.yaml")
	k, err := app.Kustomization()
	if err != nil {
		return nil, err
	}

	kyaml, err := yaml.Marshal(k)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(kustPath, kyaml, 0400)
	if err != nil {
		return nil, err
	}

	log.G().WithFields(log.Fields{
		"bootstrapKustPath": kustPath,
		"resourcePath":      app.path,
	}).Debugf("running bootstrap kustomization: %s\n", string(kyaml))

	opts := krusty.MakeDefaultOptions()
	opts.DoLegacyResourceSort = true
	kust := krusty.MakeKustomizer(opts)
	fs := filesys.MakeFsOnDisk()
	res, err := kust.Run(fs, filepath.Dir(kustPath))
	if err != nil {
		return nil, err
	}

	bootstrapManifests, err := res.AsYaml()
	if err != nil {
		return nil, err
	}

	return util.JoinManifests(createNamespace(app.namespace), bootstrapManifests), nil
}

func (app *bootstrapApp) Kustomization() (*kusttypes.Kustomization, error) {
	credsYAML, err := createCreds(app.repoUrl)
	if err != nil {
		return nil, err
	}

	k := &kusttypes.Kustomization{
		Resources: []string{app.path},
		TypeMeta: kusttypes.TypeMeta{
			APIVersion: kusttypes.KustomizationVersion,
			Kind:       kusttypes.KustomizationKind,
		},
		ConfigMapGenerator: []kusttypes.ConfigMapArgs{
			{
				GeneratorArgs: kusttypes.GeneratorArgs{
					Name:     "argocd-cm",
					Behavior: kusttypes.BehaviorMerge.String(),
					KvPairSources: kusttypes.KvPairSources{
						LiteralSources: []string{
							"repository.credentials=" + string(credsYAML),
						},
					},
				},
			},
		},
		Namespace: app.namespace,
	}

	k.FixKustomizationPostUnmarshalling()
	errs := k.EnforceFields()
	if len(errs) > 0 {
		return nil, fmt.Errorf("kustomization errors: %s", strings.Join(errs, "\n"))
	}
	k.FixKustomizationPreMarshalling()

	return k, nil
}

func (app *bootstrapApp) ArgoCD() *v1alpha1.Application {
	return app.argoApp
}

func (app *bootstrapApp) RootApp(revision, srcPath string) *v1alpha1.Application {
	return &v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: argocdapp.Group + "/v1alpha1",
			Kind:       argocdapp.ApplicationKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: app.namespace,
			Name:      "root",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": store.Common.ManagedBy,
				"app.kubernetes.io/name":       store.Common.RootName,
			},
			Finalizers: []string{
				"resources-finalizer.argocd.argoproj.io",
			},
		},
		Spec: v1alpha1.ApplicationSpec{
			Source: v1alpha1.ApplicationSource{
				RepoURL:        app.repoUrl,
				Path:           srcPath,
				TargetRevision: revision,
			},
			Destination: v1alpha1.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: app.namespace,
			},
			SyncPolicy: &v1alpha1.SyncPolicy{
				Automated: &v1alpha1.SyncPolicyAutomated{
					SelfHeal: true,
					Prune:    true,
				},
			},
		},
	}
}

func getLabels(appName string) []string {
	return []string{
		"app.kubernetes.io/managed-by=argo-autopilot",
		"app.kubernetes.io/name=" + appName,
	}
}

func createNamespace(namespace string) []byte {
	ns := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	data, err := yaml.Marshal(ns)
	util.Die(err)

	return data
}

func createCreds(repoUrl string) ([]byte, error) {
	creds := []argocdsettings.Repository{
		{
			URL: repoUrl,
			UsernameSecret: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "autopilot-secret",
				},
				Key: "git_username",
			},
			PasswordSecret: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "autopilot-secret",
				},
				Key: "git_token",
			},
		},
	}

	return yaml.Marshal(creds)
}

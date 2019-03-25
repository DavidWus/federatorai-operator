package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	fedOperator "github.com/containers-ai/federatorai-operator"
	"github.com/containers-ai/federatorai-operator/pkg/apis"
	"github.com/containers-ai/federatorai-operator/pkg/controller"
	"github.com/containers-ai/federatorai-operator/pkg/version"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

const (
	envVarPrefix  = "FEDERATORAI_OPERATOR"
	allowEmptyEnv = true
)

var (
	metricsPort           int32
	configurationFilePath string

	federatoraiOperatorFlagSet = pflag.NewFlagSet("federatorai-operator", pflag.ExitOnError)

	fedOperatorConfig fedOperator.Config

	log = logf.Log.WithName("manager")
)

func init() {

	initFlags()
	initConfiguration()
}

func initFlags() {

	federatoraiOperatorFlagSet.Int32Var(&metricsPort, "metrics.port", 8383, "port to export metrics data")
	federatoraiOperatorFlagSet.StringVar(&configurationFilePath, "config", "/etc/federatorai/operator/operator.yml", "File path to federatorai-operator coniguration")

	pflag.CommandLine.AddFlagSet(zap.FlagSet())
	pflag.CommandLine.AddFlagSet(federatoraiOperatorFlagSet)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()
}

func initConfiguration() {

	fedOperatorConfig = fedOperator.NewDefaultConfig()

	initViperSetting()
	mergeViperValueWithDefaultConfig()
}

func initViperSetting() {

	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AllowEmptyEnv(allowEmptyEnv)
	if err := viper.BindPFlags(federatoraiOperatorFlagSet); err != nil {
		panic(err)
	}
}

func mergeViperValueWithDefaultConfig() {

	viper.SetConfigFile(configurationFilePath)

	if err := viper.ReadInConfig(); err != nil {
		panic(errors.New("Read configuration file failed: " + err.Error()))
	}

	if err := viper.Unmarshal(&fedOperatorConfig); err != nil {
		panic(errors.New("Unmarshal configuration failed: " + err.Error()))
	}
}

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Federatorai Operator Version: %v", version.String))
}

func printConfiguration() {
	if b, err := json.MarshalIndent(fedOperatorConfig, "", "    "); err != nil {
		panic(err.Error())
	} else {
		log.Info(fmt.Sprintf("%+v", string(b)))
	}
}

func main() {

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	printVersion()
	printConfiguration()

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "federatorai-operator-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          fedOperatorConfig.WatchNamespace,
		MetricsBindAddress: fmt.Sprintf("%s:%d", fedOperatorConfig.Metrics.Host, fedOperatorConfig.Metrics.Port),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create Service object to expose the metrics port.
	_, err = metrics.ExposeMetricsPort(ctx, metricsPort)
	if err != nil {
		log.Info(err.Error())
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ionfury/virtualmachine-hostdevice-defaulter/webhook"
	"github.com/urfave/cli/v3"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var version string

var (
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	cmd := &cli.Command{
		Name:  "harvester-pci-mutating-webhook",
		Usage: "tbd",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Value: 6969,
				Usage: "Webhook port",
				Action: func(ctx context.Context, cmd *cli.Command, v int64) error {
					if v >= 65536 {
						return fmt.Errorf("Flag port value %v out of range [0-65536]", v)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:  "certPath",
				Value: "/etc/kubernetes/",
				Usage: "Path containing ssl.pem and ssl.key files",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) error {
			ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.JSONEncoder()))

			return nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
				WebhookServer: webhook.NewServer(webhook.Options{
					Port: int(cmd.Int("port")),
				}),
			})

			if err != nil {
				setupLog.Error(err, "unable to start webhook server")
				return err
			}

			if err := kubevirtv1.AddToScheme(mgr.GetScheme()); err != nil {
				setupLog.Error(err, "unable to add kubevirtv1 to scheme")
				return err
			}

			if err := ctrl.NewWebhookManagedBy(mgr).For(new(kubevirtv1.VirtualMachine)).Complete(webhook.NewVirtualMachineReconciler(mgr.GetClient())); err != nil {
				setupLog.Error(err, "unable to start webhook manager vor virtual machine")
				return err
			}

			setupLog.Info("starting manager")
			if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
				setupLog.Error(err, "unable to start manager")
				return err
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

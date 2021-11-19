package podhelper

import (
	"time"

	"github.com/loft-sh/vcluster/cmd/vclusterctl/log"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

func GetVClusterConfig(config *rest.Config, pod, namespace string, log log.Logger) ([]byte, error) {
	var (
		out []byte
	)
	printedWaiting := false
	err := wait.PollImmediate(time.Second*2, time.Minute*10, func() (done bool, err error) {
		stdout, _, err := ExecBuffered(config, namespace, pod, "syncer", []string{"cat", "/root/.kube/config"}, nil)
		if err != nil {
			if !printedWaiting {
				log.Infof("Waiting for vcluster to come up...")
				printedWaiting = true
			}
			log.Infof("Observed error: %v", err)

			return false, nil
		}

		out = stdout
		return true, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "wait for vcluster")
	}

	return out, nil
}

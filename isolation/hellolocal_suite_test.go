package hellolocal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pingcap/test-infra/sdk/core"
	_ "github.com/pingcap/test-infra/sdk/resource/impl/k8s"
	_ "github.com/pingcap/test-infra/sdk/resource/impl/local"

	. "github.com/pingcap/endless/pkg/util"
)

var (
	suiteTestCtx core.TestContext
)

func init() {
	suiteTestCtx = AutoConf()
}

func TestHelloworld(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hellolocal Suite")
}

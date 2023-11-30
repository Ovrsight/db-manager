package databases_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatabases(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Databases Suite")
}

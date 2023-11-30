package databases_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backup", func() {

	BeforeEach(func() {

	})

	Context("with filesystem driver", func() {
		It("can backup a database", func() {
			Expect(true).To(BeTrue())
		})
	})

	Context("with dropbox driver", func() {
		It("can backup a database", func() {
			Expect(true).To(BeTrue())
		})
	})

	Context("with google driver driver", func() {
		It("can backup a database", func() {
			Expect(true).To(BeTrue())
		})
	})
})

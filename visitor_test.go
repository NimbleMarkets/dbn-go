// Copyright (c) 2025 Neomantra Corp

package dbn_test

import (
	"github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Visitor", func() {
	Context("interfaces", func() {
		It("NullVisitor should implement dbn.Visitor", func() {
			v := dbn.NullVisitor{}
			var _ dbn.Visitor = &v
		})
	})
})

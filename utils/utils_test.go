package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Context("utils.Contains() test", func() {
		It("Should return if the string is in the string slice or not", func() {
			words := []string{
				"one", "two", "three",
			}
			foundWord := "one"
			notFoundWord := "four"
			Expect(Contains(words, foundWord)).Should(BeTrue())
			Expect(Contains(words, notFoundWord)).Should(BeFalse())
		})
	})
})

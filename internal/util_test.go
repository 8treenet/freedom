package internal

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("util.go", func() {
	Describe("NewMap", func() {
		Context("given a non-pointer", func() {
			var (
				badDst map[string]struct{}
				err    error
			)

			BeforeEach(func() {
				badDst = map[string]struct{}{}

				maybePanic := func() {
					err = NewMap(badDst)
				}

				Expect(maybePanic).ShouldNot(Panic())
			})

			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
			})
		})

		Context("given a pointer to other type", func() {
			var (
				badDst *struct{}
				err    error
			)

			BeforeEach(func() {
				badDst = &struct{}{}

				maybePanic := func() {
					err = NewMap(badDst)
				}

				Expect(maybePanic).ShouldNot(Panic())
			})

			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
			})
		})

		Context("given a pointer to map", func() {
			var (
				goodDst *map[string]struct{}
				err     error
			)

			BeforeEach(func() {
				goodDst = &map[string]struct{}{}

				maybePanic := func() {
					err = NewMap(goodDst)
				}

				Expect(maybePanic).ShouldNot(Panic())
			})

			It("should not return error", func() {
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("InSlice", func() {
		Context("given a non-slice", func() {
			var (
				badArr              int
				arbitrarySearchItem int
				ret                 bool
			)

			BeforeEach(func() {
				badArr = 10
				arbitrarySearchItem = 100

				maybePanic := func() {
					ret = InSlice(badArr, arbitrarySearchItem)
				}

				Expect(maybePanic).ShouldNot(Panic())
			})

			It("should return false", func() {
				Expect(ret).To(BeFalse())
			})
		})

		Context("given a slice", func() {
			var (
				goodArr []int
			)

			BeforeEach(func() {
				goodArr = []int{1, 3, 5, 7, 9}
			})

			When("search item does not occurred", func() {
				var (
					notOccurredSearchItem int
					ret                   bool
				)

				BeforeEach(func() {
					notOccurredSearchItem = 2

					maybePanic := func() {
						ret = InSlice(goodArr, notOccurredSearchItem)
					}

					Expect(maybePanic).ShouldNot(Panic())
				})

				It("should return false", func() {
					Expect(ret).To(BeFalse())
				})
			})

			When("search item has occurred", func() {
				var (
					occurredSearchItem int
					ret                bool
				)

				BeforeEach(func() {
					occurredSearchItem = 1

					maybePanic := func() {
						ret = InSlice(goodArr, occurredSearchItem)
					}

					Expect(maybePanic).ShouldNot(Panic())
				})

				It("should return true", func() {
					Expect(ret).To(BeTrue())
				})
			})
		})
	})
})

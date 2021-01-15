package internal

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("util.go", func() {

	Describe("NewMap", func() {

		Context("given a non-pointer", func() {
			var (
				badMapWithoutPointer map[string]struct{}
				err                  error
			)

			BeforeEach(func() {
				badMapWithoutPointer = map[string]struct{}{}

				maybePanic := func() {
					err = NewMap(badMapWithoutPointer)
				}

				Expect(maybePanic).ShouldNot(Panic())
			})

			It("should fail", func() {
				Expect(err).Should(HaveOccurred())
			})
		})

		Context("given a pointer to other type", func() {
			var (
				badMapWithOtherType *struct{}
				err                 error
			)

			BeforeEach(func() {
				badMapWithOtherType = &struct{}{}

				maybePanic := func() {
					err = NewMap(badMapWithOtherType)
				}

				Expect(maybePanic).ShouldNot(Panic())
			})

			It("should fail", func() {
				Expect(err).Should(HaveOccurred())
			})
		})

		Context("given a pointer to map", func() {
			var (
				goodMap *map[string]struct{}
				err     error
			)

			BeforeEach(func() {
				goodMap = &map[string]struct{}{}

				maybePanic := func() {
					err = NewMap(goodMap)
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
				badSliceWithIntType int
				arbitrarySearchItem int
				found               bool
			)

			BeforeEach(func() {
				badSliceWithIntType = 10
				arbitrarySearchItem = 100

				maybePanic := func() {
					found = InSlice(badSliceWithIntType, arbitrarySearchItem)
				}

				Expect(maybePanic).ShouldNot(Panic())
			})

			It("should not be found", func() {
				Expect(found).To(BeFalse())
			})
		})

		Context("given a slice", func() {
			var (
				goodSlice []int
			)

			BeforeEach(func() {
				goodSlice = []int{1, 3, 5, 7, 9}
			})

			When("search item is absent", func() {
				var (
					absentSearchItem int
					found            bool
				)

				BeforeEach(func() {
					absentSearchItem = 2

					maybePanic := func() {
						found = InSlice(goodSlice, absentSearchItem)
					}

					Expect(maybePanic).ShouldNot(Panic())
				})

				It("should not be found", func() {
					Expect(found).To(BeFalse())
				})
			})

			When("search item is occurred", func() {
				var (
					occurredSearchItem int
					found              bool
				)

				BeforeEach(func() {
					occurredSearchItem = 1

					maybePanic := func() {
						found = InSlice(goodSlice, occurredSearchItem)
					}

					Expect(maybePanic).ShouldNot(Panic())
				})

				It("should be found", func() {
					Expect(found).To(BeTrue())
				})
			})
		})
	})
})

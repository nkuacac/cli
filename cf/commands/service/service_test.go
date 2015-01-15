package service_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/service"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewShowService(ui), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not provided the name of the service to show", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			runCommand()

			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.TargetedSpaceSuccess = true

			Expect(runCommand("come-ON")).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true

			Expect(runCommand("okay-this-time-please??")).To(BeFalse())
		})
	})

	Context("when logged in, a space is targeted, and provided the name of a service that exists", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
		})

		Context("when the service is externally provided", func() {
			var serviceInstance models.ServiceInstance

			BeforeEach(func() {
				offering := models.ServiceOfferingFields{Label: "mysql", DocumentationUrl: "http://documentation.url", Description: "the-description"}
				plan := models.ServicePlanFields{Guid: "plan-guid", Name: "plan-name"}

				serviceInstance = models.ServiceInstance{}
				serviceInstance.Name = "service1"
				serviceInstance.Guid = "service1-guid"
				serviceInstance.State = "creating"
				serviceInstance.StateDescription = "creating resource - step 1"
				serviceInstance.ServicePlan = plan
				serviceInstance.ServiceOffering = offering
				serviceInstance.DashboardUrl = "some-url"
			})

			It("shows the service", func() {
				runCommand("service1")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Service instance:", "service1"},
					[]string{"Service: ", "mysql"},
					[]string{"Plan: ", "plan-name"},
					[]string{"Description: ", "the-description"},
					[]string{"Documentation url: ", "http://documentation.url"},
					[]string{"Dashboard: ", "some-url"},
					[]string{"Status: ", "unavailable (creating)"},
					[]string{"Message: ", "creating resource - step 1"},
				))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
			})

			FContext("shows correct status information based on service instance state", func() {
				It("shows status: `unavailable (creating)` when state: `creating`", func() {
					serviceInstance.State = "creating"
					requirementsFactory.ServiceInstance = serviceInstance
					runCommand("service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Status: ", "unavailable (creating)"},
					))
					Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
				})

				It("shows status: `available` when state: `available`", func() {
					serviceInstance.State = "available"
					requirementsFactory.ServiceInstance = serviceInstance
					runCommand("service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Status: ", "available"},
					))
					Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
				})

				FIt("shows status: `failed (creating)` when state: `failed`", func() {
					serviceInstance.State = "failed"
					requirementsFactory.ServiceInstance = serviceInstance
					runCommand("service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Status: ", "failed (creating)"},
					))
					Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
				})

				It("shows status: `` when state: ``", func() {
					serviceInstance.State = ""
					requirementsFactory.ServiceInstance = serviceInstance
					runCommand("service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Status: ", ""},
					))
					Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
				})
			})

			Context("when the guid flag is provided", func() {
				It("shows only the service guid", func() {
					runCommand("--guid", "service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"service1-guid"},
					))

					Expect(ui.Outputs).ToNot(ContainSubstrings(
						[]string{"Service instance:", "service1"},
					))
				})
			})
		})

		Context("when the service is user provided", func() {
			BeforeEach(func() {
				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "service1"
				serviceInstance.Guid = "service1-guid"
				requirementsFactory.ServiceInstance = serviceInstance
			})

			It("shows user provided services", func() {
				runCommand("service1")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Service instance: ", "service1"},
					[]string{"Service: ", "user-provided"},
				))
			})
		})
	})
})

var _ = Describe("ServiceInstanceStateToStatus", func() {
	It("returns status: `unavailable (creating)` when state: `creating`", func() {
		status := ServiceInstanceStateToStatus("creating")
		Expect(status).To(Equal("unavailable (creating)"))
	})

	It("returns status: `available` when state: `available`", func() {
		status := ServiceInstanceStateToStatus("available")
		Expect(status).To(Equal("available"))
	})

	It("returns status: `failed (creating)` when state: `failed`", func() {
		status := ServiceInstanceStateToStatus("failed")
		Expect(status).To(Equal("failed (creating)"))
	})

	It("returns status: `` when state: ``", func() {
		status := ServiceInstanceStateToStatus("")
		Expect(status).To(Equal(""))
	})
})

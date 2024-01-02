package controllers

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"html/template"
	v13 "k8s.io/api/batch/v1"
	v14 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"reflect"
	"time"
	v1 "tutorial.kubebuilder.io/api/v1"
)

func parseTemplate(templateName string, app *v1.CronJob) []byte {
	tmpl, err := template.ParseFiles("templates/" + templateName + ".yaml")
	if err != nil {
		panic(err)
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, app)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

var _ = Describe("Cronjob controller", func() {
	const (
		CronjobName      = "test-cronjob"
		CronjobNamespace = "default"
		JobName          = "test-job"

		timeout  = time.Second * 5
		duration = time.Second * 5
		interval = time.Millisecond * 250
	)
	var cronjob *v1.CronJob
	var cronjobLookupKey types.NamespacedName
	var createdCronjob *v1.CronJob

	BeforeEach(func() {
		println("before")
		By("By creating a new CronJob")
		cronjob = new(v1.CronJob)
		cronjob.ObjectMeta.Name = CronjobName
		cronjob.ObjectMeta.Namespace = CronjobNamespace
		Expect(yaml.Unmarshal(parseTemplate("cronjob_test", cronjob), cronjob)).Should(Succeed())
		Expect(k8sClient.Create(ctx, cronjob)).Should(Succeed())
		cronjobLookupKey = types.NamespacedName{Name: CronjobName, Namespace: CronjobNamespace}
		createdCronjob = &v1.CronJob{}
		Eventually(func() bool {
			err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
			if err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
		Expect(createdCronjob.Spec.Schedule).Should(Equal("1 * * * *"))
	})

	Context("When updating CronJob Status", func() {
		It("Should increase CronJob Status.Active count when new Jobs are created", func() {
			println("spec1")
			By("By checking the CronJob has zero active Jobs")
			Consistently(func() (int, error) {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				if err != nil {
					return -1, err
				}
				return len(createdCronjob.Status.Active), nil
			}, duration, interval).Should(Equal(0))

			By("By creating a new Job")
			testJob := &v13.Job{
				ObjectMeta: v12.ObjectMeta{
					Name:      JobName,
					Namespace: CronjobNamespace,
				},
				Spec: v13.JobSpec{
					Template: v14.PodTemplateSpec{
						Spec: v14.PodSpec{
							// For simplicity, we only fill out the required fields.
							Containers: []v14.Container{
								{
									Name:  "test-container",
									Image: "test-image",
								},
							},
							RestartPolicy: v14.RestartPolicyOnFailure,
						},
					},
				},
				Status: v13.JobStatus{
					Active: 2,
				},
			}
			kind := reflect.TypeOf(v1.CronJob{}).Name()
			gvk := v1.GroupVersion.WithKind(kind)

			controllerRef := v12.NewControllerRef(createdCronjob, gvk)
			testJob.SetOwnerReferences([]v12.OwnerReference{*controllerRef})
			Expect(k8sClient.Create(ctx, testJob)).Should(Succeed())

			By("By checking that the CronJob has one active Job")
			Eventually(func() ([]string, error) {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				if err != nil {
					return nil, err
				}

				names := []string{}
				for _, job := range createdCronjob.Status.Active {
					names = append(names, job.Name)
				}
				return names, nil
			}, timeout, interval).Should(ConsistOf(JobName), "should list our active job %s in the active jobs list in status", JobName)
		})
	})
	Context("When creating job automatically", func() {
		It("Should creat job automatically", func() {
			println("spec2")
			time.Sleep(time.Second * 70)
			By("By checking that the CronJob has one active Job")
			Eventually(func() ([]string, error) {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				if err != nil {
					return nil, err
				}

				names := []string{}
				for _, job := range createdCronjob.Status.Active {
					names = append(names, job.Name)
				}
				return names, nil
			}, timeout, interval).Should(ConsistOf(JobName), "should list our active job %s in the active jobs list in status", JobName)
		})
	})

	AfterEach(func() {
		println("after")
		Eventually(func() bool {
			err := k8sClient.Delete(ctx, cronjob)
			if err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
	})
})

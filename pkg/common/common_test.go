package common

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	workv1 "open-cluster-management.io/api/work/v1"
)

func TestFilterTimestampManifestwork(t *testing.T) {

	cases := []struct {
		name           string
		work           *workv1.ManifestWork
		expectedReport bool
		expectedHosted string
	}{
		{
			name: "no label",
			work: &workv1.ManifestWork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-work1",
					Namespace: "test",
				},
			},
			expectedReport: false,
			expectedHosted: "",
		},
		{
			name: "import hosted cluster",
			work: &workv1.ManifestWork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-work1",
					Namespace: "test",
					Labels: map[string]string{
						LabelImportHostedCluster: "hosted1",
					},
				},
			},
			expectedReport: true,
			expectedHosted: "hosted1",
		},
		{
			name: "hosted cluster",
			work: &workv1.ManifestWork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-work1",
					Namespace: "test",
					Labels: map[string]string{
						LabelImportHostedCluster: "hosted1",
					},
				},
			},
			expectedReport: true,
			expectedHosted: "hosted1",
		},
		{
			name: "sd management cluster",
			work: &workv1.ManifestWork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-work1",
					Namespace: "test",
					Labels: map[string]string{
						LabelClusterServiceManagementCluster: "hosting1",
					},
				},
			},
			expectedReport: true,
			expectedHosted: "",
		},
		{
			name: "other label",
			work: &workv1.ManifestWork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-work1",
					Namespace: "test",
					Labels: map[string]string{
						"test": "test1",
					},
				},
			},
			expectedReport: false,
			expectedHosted: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			report, hostedCluster := FilterTimestampManifestwork(c.work)
			if report != c.expectedReport {
				t.Errorf("unexpected report result: %v", report)
			}
			if hostedCluster != c.expectedHosted {
				t.Errorf("unexpected hosted cluster: %s", hostedCluster)
			}
		})
	}

}

func TestGetObservedTimestamp(t *testing.T) {
	expectTime, err := time.Parse(time.RFC3339, "2021-09-01T00:01:03Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	cases := []struct {
		name         string
		work         *workv1.ManifestWork
		expectedTime *time.Time
	}{
		{
			name: "no annotation",
			work: &workv1.ManifestWork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-work1",
					Namespace: "test",
				},
			},
		},
		{
			name: "get expected time",
			work: &workv1.ManifestWork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-work1",
					Namespace: "test",
					Annotations: map[string]string{
						AnnotationObservedTimestamp: `{"appliedTime":"2021-09-01T00:01:03Z"}`,
					},
				},
			},
			expectedTime: &expectTime,
		},
		{
			name: "appliedTime invalid",
			work: &workv1.ManifestWork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-work1",
					Namespace: "test",
					Annotations: map[string]string{
						AnnotationObservedTimestamp: `{"appliedTime":"2021-09-01T00:01:03Z"`,
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			timestamp := GetObservedTimestamp(c.work)
			if timestamp != nil {
				if c.expectedTime == nil {
					t.Errorf("expect timestamp: %v, but got %v", c.expectedTime, timestamp)
				} else if !timestamp.AppliedTime.Equal(*c.expectedTime) {
					t.Errorf("expect timestamp: %v, but got %v", c.expectedTime, timestamp)
				}

			} else {
				if c.expectedTime != nil {
					t.Errorf("expect timestamp: %v, but got %v", c.expectedTime, timestamp)
				}
			}
		})
	}

}

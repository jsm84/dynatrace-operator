package v1beta1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetInternalFlags(t *testing.T) {
	type testData struct {
		description         string
		annotatedObject     metav1.Object
		expectedMapContents string
	}
	testCases := []testData{
		{
			description:         "Empty pod should have no internal flags",
			annotatedObject:     &corev1.Pod{},
			expectedMapContents: "map[]",
		},
		{
			description: "Only internal flags should be returned (1)",
			annotatedObject: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"not-a-flag": "oh, no",
					},
				},
			},
			expectedMapContents: "map[]",
		},
		{
			description: "Only internal flags should be returned (2)",
			annotatedObject: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						InternalFlagPrefix + "some-flag":                  "something",
						InternalFlagPrefix + "other-flag":                 "nothing",
						"unexpected." + InternalFlagPrefix + "not-a-flag": "oh, no",
					},
				},
			},
			expectedMapContents: "map[internal.operator.dynatrace.com/other-flag:nothing internal.operator.dynatrace.com/some-flag:something]",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t,
			testCase.expectedMapContents, fmt.Sprint(GetInternalFlags(testCase.annotatedObject)),
			testCase.description,
		)
	}
}

func TestIsInternalFlagsEqual(t *testing.T) {
	for _, objects := range [][]metav1.Object{
		{&corev1.Pod{}, &corev1.Pod{}},
		{&corev1.Pod{}, &corev1.Service{}},
		{&corev1.Namespace{}, &corev1.Service{}},
	} {
		assert.True(t, IsInternalFlagsEqual(objects[0], objects[1]),
			"Expected that no-flag objects are equal internal flags-wise",
		)
	}

	for _, objects := range [][]metav1.Object{
		{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"": "", "space": " ", "dot": ".",
			}}},
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"comma": ",",
			}}},
		},
		{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"": "", "space": " ", "dot": ".",
			}}},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"comma": ",",
			}}},
		},
		{
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"": "", "space": " ", "dot": ".",
			}}},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"comma": ",",
			}}},
		},
	} {
		assert.True(t, IsInternalFlagsEqual(objects[0], objects[1]),
			"Expected that objects without internal operator flags compare as equal",
		)
	}

	for _, objects := range [][]metav1.Object{
		{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trace", InternalFlagPrefix + "flag": "value",
			}}},
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "truce", InternalFlagPrefix + "flag": "value",
			}}},
		},
		{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trace", InternalFlagPrefix + "flag": "value",
			}}},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "tarce", InternalFlagPrefix + "flag": "value",
			}}},
		},
		{
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trace", InternalFlagPrefix + "flag": "value",
			}}},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trcue", InternalFlagPrefix + "flag": "value",
			}}},
		},
	} {
		assert.True(t, IsInternalFlagsEqual(objects[0], objects[1]),
			"Expected that objects with internal operator flags compare as equal if the flags are identical",
		)
	}

	for _, objects := range [][]metav1.Object{
		{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trace", InternalFlagPrefix + "flag": "value",
			}}},
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trace", InternalFlagPrefix + "flag": "other value",
			}}},
		},
		{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trace", InternalFlagPrefix + "flag": "value",
			}}},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trace", InternalFlagPrefix + "flag": "other value",
			}}},
		},
		{
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trace", InternalFlagPrefix + "flag": "value",
			}}},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				"dyna": "trace", InternalFlagPrefix + "flag": "other value",
			}}},
		},
	} {
		assert.False(t, IsInternalFlagsEqual(objects[0], objects[1]),
			"Expected that objects with internal operator flags compare as different if the flags have different values or are missing",
		)
	}
}

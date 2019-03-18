package resourceread

import (
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	appsScheme = runtime.NewScheme()
	appsCodecs = serializer.NewCodecFactory(appsScheme)
)

func init() {
	if err := appsv1.AddToScheme(appsScheme); err != nil {
		panic(err)
	}
}

func ReadDeploymentV1OrDie(objBytes []byte) *appsv1.Deployment {
	requiredObj, err := runtime.Decode(appsCodecs.UniversalDecoder(appsv1.SchemeGroupVersion), objBytes)
	if err != nil {
		logrus.Fatalf("Failed to ReadDeploymentV1OrDie: %v", err)
		panic(err)
	}
	return requiredObj.(*appsv1.Deployment)
}

package resourceread

import (
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	rbacScheme = runtime.NewScheme()
	rbacCodecs = serializer.NewCodecFactory(rbacScheme)
)

func init() {
	if err := rbacv1.AddToScheme(rbacScheme); err != nil {
		panic(err)
	}
}

func ReadClusterRoleBindingV1OrDie(objBytes []byte) *rbacv1.ClusterRoleBinding {
	requiredObj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	return requiredObj.(*rbacv1.ClusterRoleBinding)
}

func ReadRoleBindingV1OrDie(objBytes []byte) *rbacv1.RoleBinding {
	requiredObj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	return requiredObj.(*rbacv1.RoleBinding)
}

func ReadClusterRoleV1OrDie(objBytes []byte) *rbacv1.ClusterRole {
	requiredObj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		logrus.Fatalf("Fail ReadClusterRoleV1OrDie: %v", err)
		panic(err)
	}
	return requiredObj.(*rbacv1.ClusterRole)
}

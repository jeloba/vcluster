package persistentvolumes

import (
	"github.com/loft-sh/vcluster/pkg/util/translate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func (s *syncer) translate(vPv *corev1.PersistentVolume) (*corev1.PersistentVolume, error) {
	target, err := s.translator.Translate(vPv)
	if err != nil {
		return nil, err
	}

	// translate the persistent volume
	pPV := target.(*corev1.PersistentVolume)
	pPV.Spec.ClaimRef = nil
	pPV.Spec.StorageClassName = translateStorageClass(s.targetNamespace, vPv.Spec.StorageClassName)
	// TODO: translate the storage secrets

	return pPV, nil
}

func translateStorageClass(physicalNamespace, vStorageClassName string) string {
	if vStorageClassName == "" {
		return ""
	}
	return translate.PhysicalNameClusterScoped(vStorageClassName, physicalNamespace)
}

func (s *syncer) translateBackwards(pPv *corev1.PersistentVolume, vPvc *corev1.PersistentVolumeClaim) *corev1.PersistentVolume {
	// build virtual persistent volume
	vObj := pPv.DeepCopy()
	vObj.ResourceVersion = ""
	vObj.UID = ""
	vObj.ManagedFields = nil
	if vPvc != nil {
		vObj.Spec.ClaimRef.ResourceVersion = vPvc.ResourceVersion
		vObj.Spec.ClaimRef.UID = vPvc.UID
		vObj.Spec.ClaimRef.Name = vPvc.Name
		vObj.Spec.ClaimRef.Namespace = vPvc.Namespace
	}
	if vObj.Annotations == nil {
		vObj.Annotations = map[string]string{}
	}
	vObj.Annotations[HostClusterPersistentVolumeAnnotation] = pPv.Name
	return vObj
}

func (s *syncer) translateUpdateBackwards(vPv *corev1.PersistentVolume, pPv *corev1.PersistentVolume, vPvc *corev1.PersistentVolumeClaim) *corev1.PersistentVolume {
	var updated *corev1.PersistentVolume

	// build virtual persistent volume
	translatedSpec := *pPv.Spec.DeepCopy()
	if vPvc != nil {
		translatedSpec.ClaimRef.ResourceVersion = vPvc.ResourceVersion
		translatedSpec.ClaimRef.UID = vPvc.UID
		translatedSpec.ClaimRef.Name = vPvc.Name
		translatedSpec.ClaimRef.Namespace = vPvc.Namespace
	}

	// check storage class
	if translate.IsManagedCluster(s.targetNamespace, pPv) == false {
		if equality.Semantic.DeepEqual(vPv.Spec.StorageClassName, translatedSpec.StorageClassName) == false {
			updated = newIfNil(updated, vPv)
			updated.Spec.StorageClassName = translatedSpec.StorageClassName
		}
	}

	// check claim ref
	if equality.Semantic.DeepEqual(vPv.Spec.ClaimRef, translatedSpec.ClaimRef) == false {
		updated = newIfNil(updated, vPv)
		updated.Spec.ClaimRef = translatedSpec.ClaimRef
	}

	return updated
}

func (s *syncer) translateUpdate(vPv *corev1.PersistentVolume, pPv *corev1.PersistentVolume) *corev1.PersistentVolume {
	var updated *corev1.PersistentVolume

	// TODO: translate the storage secrets
	if equality.Semantic.DeepEqual(pPv.Spec.PersistentVolumeSource, vPv.Spec.PersistentVolumeSource) == false {
		updated = newIfNil(updated, pPv)
		updated.Spec.PersistentVolumeSource = vPv.Spec.PersistentVolumeSource
	}

	if equality.Semantic.DeepEqual(pPv.Spec.Capacity, vPv.Spec.Capacity) == false {
		updated = newIfNil(updated, pPv)
		updated.Spec.Capacity = vPv.Spec.Capacity
	}

	if equality.Semantic.DeepEqual(pPv.Spec.AccessModes, vPv.Spec.AccessModes) == false {
		updated = newIfNil(updated, pPv)
		updated.Spec.AccessModes = vPv.Spec.AccessModes
	}

	if equality.Semantic.DeepEqual(pPv.Spec.PersistentVolumeReclaimPolicy, vPv.Spec.PersistentVolumeReclaimPolicy) == false {
		updated = newIfNil(updated, pPv)
		updated.Spec.PersistentVolumeReclaimPolicy = vPv.Spec.PersistentVolumeReclaimPolicy
	}

	translatedStorageClassName := translateStorageClass(s.targetNamespace, vPv.Spec.StorageClassName)
	if equality.Semantic.DeepEqual(pPv.Spec.StorageClassName, translatedStorageClassName) == false {
		updated = newIfNil(updated, pPv)
		updated.Spec.StorageClassName = translatedStorageClassName
	}

	if equality.Semantic.DeepEqual(pPv.Spec.NodeAffinity, vPv.Spec.NodeAffinity) == false {
		updated = newIfNil(updated, pPv)
		updated.Spec.NodeAffinity = vPv.Spec.NodeAffinity
	}

	if equality.Semantic.DeepEqual(pPv.Spec.VolumeMode, vPv.Spec.VolumeMode) == false {
		updated = newIfNil(updated, pPv)
		updated.Spec.VolumeMode = vPv.Spec.VolumeMode
	}

	if equality.Semantic.DeepEqual(pPv.Spec.MountOptions, vPv.Spec.MountOptions) == false {
		updated = newIfNil(updated, pPv)
		updated.Spec.MountOptions = vPv.Spec.MountOptions
	}

	updatedAnnotations := s.translator.TranslateAnnotations(vPv, pPv)
	if !equality.Semantic.DeepEqual(updatedAnnotations, pPv.Annotations) {
		updated = newIfNil(updated, pPv)
		updated.Annotations = updatedAnnotations
	}

	// check labels
	updatedLabels := s.translator.TranslateLabels(vPv)
	if !equality.Semantic.DeepEqual(updatedLabels, pPv.Labels) {
		updated = newIfNil(updated, pPv)
		updated.Labels = updatedLabels
	}

	return updated
}

func newIfNil(updated *corev1.PersistentVolume, obj *corev1.PersistentVolume) *corev1.PersistentVolume {
	if updated == nil {
		return obj.DeepCopy()
	}
	return updated
}

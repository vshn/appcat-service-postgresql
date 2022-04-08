// Code generated by angryjet. DO NOT EDIT.

package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// GetCondition of this PostgresqlStandalone.
func (mg *PostgresqlStandalone) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this PostgresqlStandalone.
func (mg *PostgresqlStandalone) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetProviderConfigReference of this PostgresqlStandalone.
func (mg *PostgresqlStandalone) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

/*
GetProviderReference of this PostgresqlStandalone.
Deprecated: Use GetProviderConfigReference.
*/
func (mg *PostgresqlStandalone) GetProviderReference() *xpv1.Reference {
	return mg.Spec.ProviderReference
}

// GetWriteConnectionSecretToReference of this PostgresqlStandalone.
func (mg *PostgresqlStandalone) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this PostgresqlStandalone.
func (mg *PostgresqlStandalone) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this PostgresqlStandalone.
func (mg *PostgresqlStandalone) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetProviderConfigReference of this PostgresqlStandalone.
func (mg *PostgresqlStandalone) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

/*
SetProviderReference of this PostgresqlStandalone.
Deprecated: Use SetProviderConfigReference.
*/
func (mg *PostgresqlStandalone) SetProviderReference(r *xpv1.Reference) {
	mg.Spec.ProviderReference = r
}

// SetWriteConnectionSecretToReference of this PostgresqlStandalone.
func (mg *PostgresqlStandalone) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

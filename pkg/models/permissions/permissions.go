package permissions

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
)

type DatasetPermission int64

const (
	ViewGraphSchema DatasetPermission = iota
	ManageGraphSchema
	ManageModelTemplates
	ManageDatasetTemplates
	PublishDatasetTemplate
	CreateDeleteRecord
	CreateDeleteFiles
	EditRecords
	EditFiles
	ViewRecords
	ViewFiles
	ManageCollections
	ManageRecordRelationships
	ManageDatasetCollections
	AddPeople
	ChangeRoles
	ViewPeopleAndRoles
	TransferOwnership
	ReserveDoi
	ManageAnnotations
	ManageAnnotationLayers
	ViewAnnotations
	ManageDiscussionComments
	ViewDiscussionComments
	EditContributors
	EditDatasetName
	EditDatasetDescription
	EditDatasetAutomaticallyProcessingPackages
	DeleteDataset
	RequestRevise
	RequestCancelPublishRevise
	ShowSettingsPage
	ViewExternalPublications
	ManageExternalPublications
	ViewWebhooks
	ManageWebhooks
	TriggerCustomEvents
)

type OrganizationPermission int64

const (
	CreateDatasetFromTemplate OrganizationPermission = iota
)

var Permission = map[string]string{
	"Viewer": "[ ]",
}

func rolePermissions(r role.Role) []DatasetPermission {

	var permissionSet []DatasetPermission
	switch r {
	case role.Viewer:
		permissionSet = []DatasetPermission{
			ViewGraphSchema,
			ViewRecords,
			ViewFiles,
			ViewAnnotations,
			ViewPeopleAndRoles,
			ManageDiscussionComments,
			ViewDiscussionComments,
			ViewExternalPublications,
			ViewWebhooks,
		}

	case role.Editor:
		permissionSet = rolePermissions(role.Viewer)
		permissionSet = append(permissionSet, []DatasetPermission{
			CreateDeleteRecord,
			CreateDeleteFiles,
			EditRecords,
			EditFiles,
			ManageCollections,
			ManageRecordRelationships,
			ManageAnnotations,
			ManageAnnotationLayers,
			TriggerCustomEvents,
		}...)

	case role.Manager:
		permissionSet = rolePermissions(role.Editor)
		permissionSet = append(permissionSet, []DatasetPermission{
			ManageGraphSchema,
			ManageModelTemplates,
			ManageDatasetTemplates,
			PublishDatasetTemplate,
			AddPeople,
			ChangeRoles,
			EditDatasetName,
			EditDatasetDescription,
			EditDatasetAutomaticallyProcessingPackages,
			EditContributors,
			ShowSettingsPage,
			RequestRevise,
			ReserveDoi,
			ManageDatasetCollections,
			ManageExternalPublications,
			ManageWebhooks,
		}...)

	case role.Owner:
		permissionSet = rolePermissions(role.Manager)
		permissionSet = append(permissionSet, []DatasetPermission{
			TransferOwnership,
			DeleteDataset,
			RequestCancelPublishRevise,
		}...)
	}

	return permissionSet
}

func HasDatasetPermission(r role.Role, permission DatasetPermission) bool {
	permSet := rolePermissions(r)

	for _, v := range permSet {
		if v == permission {
			return true
		}
	}

	return false
}

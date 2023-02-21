package permissions

import (
	"github.com/pennsieve/pennsieve-go-core/models/dataset"
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

func rolePermissions(role dataset.Role) []DatasetPermission {

	var permissionSet []DatasetPermission
	switch role {
	case dataset.Viewer:
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

	case dataset.Editor:
		permissionSet = rolePermissions(dataset.Viewer)
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

	case dataset.Manager:
		permissionSet = rolePermissions(dataset.Editor)
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

	case dataset.Owner:
		permissionSet = rolePermissions(dataset.Manager)
		permissionSet = append(permissionSet, []DatasetPermission{
			TransferOwnership,
			DeleteDataset,
			RequestCancelPublishRevise,
		}...)
	}

	return permissionSet
}

func HasDatasetPermission(role dataset.Role, permission DatasetPermission) bool {
	permSet := rolePermissions(role)

	for _, v := range permSet {
		if v == permission {
			return true
		}
	}

	return false
}

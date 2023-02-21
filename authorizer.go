package pkg

// UploadSession contains the information that is shared based on the upload session ID
type PennsieveContext struct {
	organizationId  int    `json:"organization_id"`
	datasetId       int    `json:"dataset_id"`
	ownerId         int    `json:"owner_id"`
	targetPackageId string `json:"target_package_id"`
}
